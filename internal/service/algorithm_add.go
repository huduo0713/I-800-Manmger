package service

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"demo/internal/dao"
	"demo/internal/model/do"
)

// AlgorithmDownloadService 算法下载服务
type AlgorithmDownloadService struct {
	downloadPath string // 下载基础路径 "/usr/runtime"
}

// NewAlgorithmDownloadService 创建算法下载服务实例
func NewAlgorithmDownloadService() *AlgorithmDownloadService {
	ctx := gctx.New()

	// 从配置文件读取下载路径，支持跨平台
	downloadPath := g.Cfg().MustGet(ctx, "algorithm.downloadPath").String()

	// 如果配置文件未设置，使用默认路径
	if downloadPath == "" {
		if runtime.GOOS == "windows" {
			// Windows环境：使用当前工作目录下的runtime/algorithm文件夹
			downloadPath = "./runtime/algorithm"
		} else {
			// Linux/Unix环境：使用/usr/runtime/algorithm
			downloadPath = "/usr/runtime/algorithm"
		}
	}

	g.Log().Info(ctx, "算法下载服务初始化", g.Map{
		"downloadPath": downloadPath,
		"platform":     runtime.GOOS,
	})

	return &AlgorithmDownloadService{
		downloadPath: downloadPath,
	}
}

// DownloadAlgorithmFile 下载算法文件并验证MD5
func (s *AlgorithmDownloadService) DownloadAlgorithmFile(algorithmId, algorithmVersionId, url, md5sum string) (string, error) {
	// 创建目标目录: {algorithmId}/{algorithmVersionId}
	targetDir := filepath.Join(s.downloadPath, algorithmId, algorithmVersionId)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 从URL中提取文件名
	fileName := filepath.Base(url)
	if fileName == "" || fileName == "." {
		fileName = "algorithm.bin"
	}
	targetPath := filepath.Join(targetDir, fileName)

	// 下载文件
	g.Log().Info(gctx.New(), "开始下载算法文件", g.Map{
		"url":        url,
		"targetPath": targetPath,
	})

	resp, err := http.Get(url)
	if err != nil {
		// 下载失败，清理创建的目录
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("下载文件失败: %v", err)
	}
	defer func() {
		// 确保响应体总是被关闭
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// HTTP状态码异常，等待一小段时间再清理目录
		time.Sleep(100 * time.Millisecond) // 给系统时间释放文件句柄
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(targetPath)
	if err != nil {
		// 文件创建失败，清理创建的目录
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入文件并计算MD5
	hash := md5.New()
	writer := io.MultiWriter(file, hash)

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(targetPath) // 删除不完整的文件
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	// 验证MD5
	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
	if calculatedMD5 != md5sum {
		os.Remove(targetPath) // 删除MD5不匹配的文件
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("MD5校验失败，期望: %s, 实际: %s", md5sum, calculatedMD5)
	}

	g.Log().Info(gctx.New(), "算法文件MD5校验成功", g.Map{
		"targetPath": targetPath,
		"md5":        calculatedMD5,
	})

	// MD5校验成功后，解压算法文件到同级目录
	if err := s.extractAlgorithmFile(targetPath, targetDir); err != nil {
		os.Remove(targetPath) // 删除压缩文件
		s.cleanupDirectoryOnFailure(targetDir, algorithmId)
		return "", fmt.Errorf("解压算法文件失败: %v", err)
	}

	// 解压成功后删除原始压缩文件
	if err := os.Remove(targetPath); err != nil {
		g.Log().Warning(gctx.New(), "删除压缩文件失败", g.Map{
			"targetPath": targetPath,
			"error":      err,
		})
	}

	g.Log().Info(gctx.New(), "算法文件下载和解压完成", g.Map{
		"targetDir": targetDir,
		"md5":       calculatedMD5,
	})

	return targetDir, nil
}

// SyncAlgorithmToDatabase 同步算法信息到数据库
func (s *AlgorithmDownloadService) SyncAlgorithmToDatabase(req *AlgorithmAddRequest, localPath string) error {
	ctx := gctx.New()

	// 检查是否已存在相同版本的算法（基于algorithmId + algorithmVersion）
	existingVersion, err := dao.Algorithm.Ctx(ctx).
		Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).
		Where(dao.Algorithm.Columns().AlgorithmVersion, req.Params.AlgorithmVersion).
		One()
	if err != nil {
		return fmt.Errorf("查询算法版本失败: %v", err)
	}

	// 如果相同版本已存在，忽略此次下发
	if !existingVersion.IsEmpty() {
		g.Log().Info(ctx, "算法版本已存在，忽略下发", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"version":     req.Params.AlgorithmVersion,
			"localPath":   existingVersion["local_path"].String(),
		})
		return nil
	}

	// 查询该算法ID的所有旧版本
	existingAlgorithm, err := dao.Algorithm.Ctx(ctx).
		Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).
		One()
	if err != nil {
		return fmt.Errorf("查询旧版本算法失败: %v", err)
	}

	// 准备数据对象
	algorithmData := do.Algorithm{
		AlgorithmId:        req.Params.AlgorithmId,
		AlgorithmName:      req.Params.AlgorithmName,
		AlgorithmVersion:   req.Params.AlgorithmVersion,
		AlgorithmVersionId: req.Params.AlgorithmVersionId,
		AlgorithmDataUrl:   req.Params.AlgorithmDataUrl,
		FileSize:           req.Params.FileSize,
		Md5:                req.Params.Md5,
		LocalPath:          localPath,
	}

	// 如果是全新的算法ID，直接插入
	if existingAlgorithm.IsEmpty() {
		_, err = dao.Algorithm.Ctx(ctx).Data(algorithmData).Insert()
		if err != nil {
			return fmt.Errorf("插入新算法记录失败: %v", err)
		}
		g.Log().Info(ctx, "新增算法记录成功", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"version":     req.Params.AlgorithmVersion,
		})
	} else {
		// 算法ID已存在，但版本不同 - 需要删除旧版本，安装新版本
		oldVersion := existingAlgorithm["algorithm_version"].String()
		oldLocalPath := existingAlgorithm["local_path"].String()

		g.Log().Info(ctx, "检测到算法版本更新", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"oldVersion":  oldVersion,
			"newVersion":  req.Params.AlgorithmVersion,
		})

		// 清理旧版本文件
		if oldLocalPath != "" {
			// 清理旧算法的整个algorithmId目录（因为版本改变了）
			oldAlgorithmDir := filepath.Join(s.downloadPath, req.Params.AlgorithmId)
			if err := os.RemoveAll(oldAlgorithmDir); err != nil {
				g.Log().Warning(ctx, "删除旧版本算法目录失败", g.Map{
					"oldAlgorithmDir": oldAlgorithmDir,
					"error":           err,
				})
			} else {
				g.Log().Info(ctx, "清理旧版本算法目录成功", g.Map{
					"oldAlgorithmDir": oldAlgorithmDir,
				})
			}
		}

		// 更新为新版本记录
		_, err = dao.Algorithm.Ctx(ctx).
			Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).
			Data(algorithmData).
			Update()
		if err != nil {
			return fmt.Errorf("更新算法版本记录失败: %v", err)
		}
		g.Log().Info(ctx, "算法版本更新成功", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"oldVersion":  oldVersion,
			"newVersion":  req.Params.AlgorithmVersion,
		})
	}

	return nil
}

// cleanupDirectoryOnFailure 下载失败时清理创建的空目录
func (s *AlgorithmDownloadService) cleanupDirectoryOnFailure(targetDir, algorithmId string) {
	ctx := gctx.New()

	g.Log().Info(ctx, "开始清理失败目录", g.Map{
		"targetDir":   targetDir,
		"algorithmId": algorithmId,
	})

	// 1. 延迟一小段时间，确保文件句柄完全释放
	time.Sleep(100 * time.Millisecond)

	// 2. 先尝试删除algorithmVersionId目录（targetDir本身）
	if err := s.removeDirectoryWithRetry(targetDir, 3); err != nil {
		g.Log().Warning(ctx, "清理版本目录失败", g.Map{
			"targetDir": targetDir,
			"error":     err,
		})
	} else {
		g.Log().Info(ctx, "清理版本目录成功", g.Map{
			"targetDir": targetDir,
		})
	}

	// 3. 尝试删除algorithmId目录（如果为空的话）
	algorithmDir := filepath.Join(s.downloadPath, algorithmId)

	// 检查算法目录是否为空
	if isEmpty, err := s.isDirEmpty(algorithmDir); err != nil {
		g.Log().Warning(ctx, "检查算法目录状态失败", g.Map{
			"algorithmDir": algorithmDir,
			"error":        err,
		})
	} else if isEmpty {
		// 目录为空，可以安全删除
		if err := s.removeDirectoryWithRetry(algorithmDir, 3); err != nil {
			g.Log().Warning(ctx, "清理空算法目录失败", g.Map{
				"algorithmDir": algorithmDir,
				"error":        err,
			})
		} else {
			g.Log().Info(ctx, "清理空算法目录成功", g.Map{
				"algorithmDir": algorithmDir,
			})
		}
	} else {
		g.Log().Debug(ctx, "算法目录不为空，跳过清理", g.Map{
			"algorithmDir": algorithmDir,
		})
	}
}

// removeDirectoryWithRetry 带重试的目录删除
func (s *AlgorithmDownloadService) removeDirectoryWithRetry(dirPath string, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// 每次重试前等待一段时间
			time.Sleep(time.Duration(i*200) * time.Millisecond)
		}

		// 检查目录是否存在
		if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
			return nil // 目录不存在，删除成功
		}

		// 尝试删除
		err = os.RemoveAll(dirPath)
		if err == nil {
			return nil // 删除成功
		}

		g.Log().Debug(gctx.New(), "目录删除重试", g.Map{
			"dirPath":    dirPath,
			"attempt":    i + 1,
			"maxRetries": maxRetries,
			"error":      err,
		})
	}
	return err
}

// isDirEmpty 检查目录是否为空
func (s *AlgorithmDownloadService) isDirEmpty(dirPath string) (bool, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // 目录不存在，视为空
		}
		return false, err
	}
	defer dir.Close()

	// 尝试读取一个目录项
	_, err = dir.Readdir(1)
	if err == io.EOF {
		return true, nil // 没有内容，目录为空
	}
	if err != nil {
		return false, err
	}
	return false, nil // 有内容，目录不为空
}

// extractAlgorithmFile 解压算法文件到指定目录
func (s *AlgorithmDownloadService) extractAlgorithmFile(zipPath, targetDir string) error {
	ctx := gctx.New()

	g.Log().Info(ctx, "开始解压算法文件", g.Map{
		"zipPath":   zipPath,
		"targetDir": targetDir,
	})

	// 打开zip文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %v", err)
	}
	defer reader.Close()

	// 解压每个文件
	for _, file := range reader.File {
		// 构建目标文件路径
		destPath := filepath.Join(targetDir, file.Name)

		// 防止目录遍历攻击（确保文件在目标目录内）
		if !strings.HasPrefix(destPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("文件路径不安全: %s", file.Name)
		}

		g.Log().Debug(ctx, "解压文件", g.Map{
			"fileName": file.Name,
			"destPath": destPath,
		})

		// 如果是目录，创建目录并继续
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("创建目录失败 %s: %v", destPath, err)
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %v", err)
		}

		// 打开zip中的文件
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开zip文件内容失败 %s: %v", file.Name, err)
		}

		// 创建目标文件
		destFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("创建目标文件失败 %s: %v", destPath, err)
		}

		// 复制文件内容
		_, err = io.Copy(destFile, rc)
		destFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("复制文件内容失败 %s: %v", destPath, err)
		}

		// 设置文件权限
		if err := os.Chmod(destPath, file.FileInfo().Mode()); err != nil {
			g.Log().Warning(ctx, "设置文件权限失败", g.Map{
				"destPath": destPath,
				"error":    err,
			})
		}
	}

	g.Log().Info(ctx, "算法文件解压完成", g.Map{
		"zipPath":   zipPath,
		"targetDir": targetDir,
		"fileCount": len(reader.File),
	})

	return nil
}

package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
			// Windows环境：使用当前工作目录下的runtime文件夹
			downloadPath = "./runtime"
		} else {
			// Linux/Unix环境：使用/usr/runtime
			downloadPath = "/usr/runtime"
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
func (s *AlgorithmDownloadService) DownloadAlgorithmFile(algorithmId, md5sum, url string) (string, error) {
	// 创建目标目录
	targetDir := filepath.Join(s.downloadPath, algorithmId, md5sum)
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

	g.Log().Info(gctx.New(), "算法文件下载完成", g.Map{
		"targetPath": targetPath,
		"md5":        calculatedMD5,
	})

	return targetPath, nil
}

// SyncAlgorithmToDatabase 同步算法信息到数据库
func (s *AlgorithmDownloadService) SyncAlgorithmToDatabase(req *AlgorithmAddRequest, localPath string) error {
	ctx := gctx.New()

	// 检查是否已存在相同版本的算法
	existing, err := dao.Algorithm.Ctx(ctx).Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).One()
	if err != nil {
		return fmt.Errorf("查询数据库失败: %v", err)
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

	if existing.IsEmpty() {
		// 新增算法记录
		_, err = dao.Algorithm.Ctx(ctx).Data(algorithmData).Insert()
		if err != nil {
			return fmt.Errorf("插入算法记录失败: %v", err)
		}
		g.Log().Info(ctx, "新增算法记录成功", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"version":     req.Params.AlgorithmVersion,
		})
	} else {
		// 算法已存在，进行覆盖更新
		existingVersion := existing["algorithm_version"].String()
		oldLocalPath := existing["local_path"].String()

		// 无论版本是否相同，都要清理旧文件（实现真正的覆盖）
		if oldLocalPath != "" && oldLocalPath != localPath {
			// 清理旧算法的整个目录
			oldDir := filepath.Dir(filepath.Dir(oldLocalPath)) // 获取 algorithmId 级别的目录
			if err := os.RemoveAll(oldDir); err != nil {
				g.Log().Warning(ctx, "删除旧算法目录失败", g.Map{
					"oldDir": oldDir,
					"error":  err,
				})
			} else {
				g.Log().Info(ctx, "清理旧算法目录成功", g.Map{
					"oldDir": oldDir,
				})
			}
		}

		if req.Params.AlgorithmVersion != existingVersion {
			g.Log().Info(ctx, "检测到算法版本变更", g.Map{
				"algorithmId":     req.Params.AlgorithmId,
				"existingVersion": existingVersion,
				"newVersion":      req.Params.AlgorithmVersion,
			})
		} else {
			g.Log().Info(ctx, "覆盖相同版本算法", g.Map{
				"algorithmId": req.Params.AlgorithmId,
				"version":     req.Params.AlgorithmVersion,
			})
		}

		// 更新记录
		_, err = dao.Algorithm.Ctx(ctx).
			Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).
			Data(algorithmData).
			Update()
		if err != nil {
			return fmt.Errorf("更新算法记录失败: %v", err)
		}
		g.Log().Info(ctx, "算法覆盖更新成功", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"version":     req.Params.AlgorithmVersion,
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

	// 2. 先尝试删除md5sum目录（targetDir本身）
	if err := s.removeDirectoryWithRetry(targetDir, 3); err != nil {
		g.Log().Warning(ctx, "清理MD5目录失败", g.Map{
			"targetDir": targetDir,
			"error":     err,
		})
	} else {
		g.Log().Info(ctx, "清理MD5目录成功", g.Map{
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

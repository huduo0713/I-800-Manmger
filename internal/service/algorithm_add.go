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

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"demo/internal/dao"
	"demo/internal/model/do"
)

// AlgorithmAddRequest 算法添加请求结构
type AlgorithmAddRequest struct {
	CmdId     string `json:"cmdId"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Params    struct {
		AlgorithmId        string `json:"algorithmId"`
		AlgorithmName      string `json:"algorithmName"`
		AlgorithmVersion   string `json:"algorithmVersion"`
		AlgorithmVersionId string `json:"algorithmVersionId"`
		AlgorithmDataUrl   string `json:"algorithmDataUrl"`
		FileSize           int64  `json:"fileSize"`
		LastModifyTime     string `json:"lastModifyTime"`
		Md5                string `json:"md5"`
	} `json:"params"`
}

// AlgorithmReply 算法操作响应结构
type AlgorithmReply struct {
	CmdId     string      `json:"cmdId"`
	Version   string      `json:"version"`
	Method    string      `json:"method"`
	Timestamp string      `json:"timestamp"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

// 错误码定义
const (
	CodeSuccess         = 0
	CodeDownloadFailed  = 1001
	CodeMd5CheckFailed  = 1002
	CodeFileSystemError = 1003
	CodeDatabaseError   = 1004
	CodeInvalidParams   = 1005
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
		return "", fmt.Errorf("下载文件失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入文件并计算MD5
	hash := md5.New()
	writer := io.MultiWriter(file, hash)

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(targetPath) // 清理失败的文件
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	// 验证MD5
	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
	if calculatedMD5 != md5sum {
		os.Remove(targetPath) // MD5不匹配，删除文件
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
		// 更新现有记录，进行版本比较
		existingVersion := existing["algorithm_version"].String()
		if req.Params.AlgorithmVersion != existingVersion {
			g.Log().Info(ctx, "检测到算法版本变更", g.Map{
				"algorithmId":     req.Params.AlgorithmId,
				"existingVersion": existingVersion,
				"newVersion":      req.Params.AlgorithmVersion,
			})

			// 删除旧文件（如果路径不同）
			oldLocalPath := existing["local_path"].String()
			if oldLocalPath != "" && oldLocalPath != localPath {
				if err := os.RemoveAll(filepath.Dir(oldLocalPath)); err != nil {
					g.Log().Warning(ctx, "删除旧算法文件失败", g.Map{
						"oldPath": oldLocalPath,
						"error":   err,
					})
				}
			}
		}

		// 更新记录
		_, err = dao.Algorithm.Ctx(ctx).
			Where(dao.Algorithm.Columns().AlgorithmId, req.Params.AlgorithmId).
			Data(algorithmData).
			Update()
		if err != nil {
			return fmt.Errorf("更新算法记录失败: %v", err)
		}
		g.Log().Info(ctx, "更新算法记录成功", g.Map{
			"algorithmId": req.Params.AlgorithmId,
			"version":     req.Params.AlgorithmVersion,
		})
	}

	return nil
}

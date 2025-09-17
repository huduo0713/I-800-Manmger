package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"demo/internal/dao"
)

// AlgorithmDeleteService 算法删除服务
type AlgorithmDeleteService struct {
	downloadPath string // 下载基础路径 "/usr/runtime"
}

// NewAlgorithmDeleteService 创建算法删除服务实例
func NewAlgorithmDeleteService() *AlgorithmDeleteService {
	ctx := gctx.New()

	// 从配置文件读取下载路径，支持跨平台
	downloadPath := g.Cfg().MustGet(ctx, "algorithm.downloadPath").String()

	// 如果配置中没有设置路径，使用默认值
	if downloadPath == "" {
		// 根据操作系统设置不同的默认路径
		if runtime.GOOS == "windows" {
			downloadPath = "./runtime"
		} else {
			downloadPath = "/usr/runtime"
		}
	}

	g.Log().Info(ctx, "算法删除服务初始化", g.Map{
		"downloadPath": downloadPath,
		"platform":     runtime.GOOS,
	})

	return &AlgorithmDeleteService{
		downloadPath: downloadPath,
	}
}

// CheckAlgorithmExists 检查算法是否存在
func (s *AlgorithmDeleteService) CheckAlgorithmExists(algorithmId string) (bool, *map[string]interface{}, error) {
	ctx := gctx.New()

	// 查询数据库中是否存在该算法
	existing, err := dao.Algorithm.Ctx(ctx).Where(dao.Algorithm.Columns().AlgorithmId, algorithmId).One()
	if err != nil {
		return false, nil, fmt.Errorf("查询数据库失败: %v", err)
	}

	if existing.IsEmpty() {
		return false, nil, nil
	}

	// 转换为map以便获取字段值
	record := existing.Map()
	return true, &record, nil
}

// DeleteAlgorithmFiles 删除算法文件目录
func (s *AlgorithmDeleteService) DeleteAlgorithmFiles(algorithmId string) error {
	ctx := gctx.New()

	// 构建算法目录路径
	algorithmDir := filepath.Join(s.downloadPath, "algorithm", algorithmId)

	// 检查目录是否存在
	if _, err := os.Stat(algorithmDir); os.IsNotExist(err) {
		g.Log().Warning(ctx, "算法目录不存在", g.Map{
			"algorithmId": algorithmId,
			"directory":   algorithmDir,
		})
		return nil // 目录不存在也算删除成功
	}

	// 删除整个算法目录
	err := os.RemoveAll(algorithmDir)
	if err != nil {
		return fmt.Errorf("删除算法目录失败: %v", err)
	}

	g.Log().Info(ctx, "算法文件删除成功", g.Map{
		"algorithmId": algorithmId,
		"directory":   algorithmDir,
	})

	return nil
}

// DeleteAlgorithmFromDatabase 从数据库中删除算法记录
func (s *AlgorithmDeleteService) DeleteAlgorithmFromDatabase(algorithmId string) error {
	ctx := gctx.New()

	// 删除数据库记录
	_, err := dao.Algorithm.Ctx(ctx).
		Where(dao.Algorithm.Columns().AlgorithmId, algorithmId).
		Delete()
	if err != nil {
		return fmt.Errorf("删除算法数据库记录失败: %v", err)
	}

	g.Log().Info(ctx, "算法数据库记录删除成功", g.Map{
		"algorithmId": algorithmId,
	})

	return nil
}

// DeleteAlgorithm 完整的算法删除流程
func (s *AlgorithmDeleteService) DeleteAlgorithm(algorithmId string) error {
	ctx := gctx.New()

	// 1. 检查算法是否存在
	exists, algorithmInfo, err := s.CheckAlgorithmExists(algorithmId)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("算法不存在: %s", algorithmId)
	}

	g.Log().Info(ctx, "开始删除算法", g.Map{
		"algorithmId": algorithmId,
		"version":     (*algorithmInfo)["algorithm_version"],
	})

	// 2. 删除文件目录
	err = s.DeleteAlgorithmFiles(algorithmId)
	if err != nil {
		g.Log().Error(ctx, "删除算法文件失败", g.Map{
			"algorithmId": algorithmId,
			"error":       err,
		})
		return err
	}

	// 3. 删除数据库记录
	err = s.DeleteAlgorithmFromDatabase(algorithmId)
	if err != nil {
		g.Log().Error(ctx, "删除算法数据库记录失败", g.Map{
			"algorithmId": algorithmId,
			"error":       err,
		})
		return err
	}

	g.Log().Info(ctx, "算法删除完成", g.Map{
		"algorithmId": algorithmId,
	})

	return nil
}

// isDirEmpty 检查目录是否为空
func (s *AlgorithmDeleteService) isDirEmpty(dirPath string) (bool, error) {
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

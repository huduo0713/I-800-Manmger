package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gogf/gf/v2/frame/g"
	"gopkg.in/yaml.v3"

	"demo/internal/dao"
	"demo/internal/model/entity"
)

// AlgorithmConfigService 算法启停配置服务
type AlgorithmConfigService struct{}

// NewAlgorithmConfigService 创建算法启停配置服务实例
func NewAlgorithmConfigService() *AlgorithmConfigService {
	return &AlgorithmConfigService{}
}

// UpdateAlgorithmRunStatus 更新算法运行状态
func (s *AlgorithmConfigService) UpdateAlgorithmRunStatus(ctx context.Context, algorithmId string, runStatus int) error {
	// 1. 参数验证
	if algorithmId == "" {
		return fmt.Errorf("算法ID不能为空")
	}

	if runStatus != 0 && runStatus != 1 {
		return fmt.Errorf("运行状态值无效，只能是0(关闭)或1(开启)")
	}

	// 2. 从数据库查询算法是否存在
	algorithm, err := s.queryAlgorithmFromDB(ctx, algorithmId)
	if err != nil {
		return fmt.Errorf("查询算法失败: %v", err)
	}

	if algorithm == nil {
		return fmt.Errorf("算法不存在，algorithmId: %s", algorithmId)
	}

	g.Log().Info(ctx, "找到算法记录", g.Map{
		"algorithmId":        algorithm.AlgorithmId,
		"algorithmName":      algorithm.AlgorithmName,
		"algorithmVersionId": algorithm.AlgorithmVersionId,
	})

	// 3. 构建配置文件路径
	configPath := s.buildConfigPath(algorithm.AlgorithmId, algorithm.AlgorithmVersionId)

	// 4. 更新配置文件中的runStatus
	err = s.updateConfigFile(ctx, configPath, runStatus)
	if err != nil {
		return fmt.Errorf("更新配置文件失败: %v", err)
	}

	g.Log().Info(ctx, "算法运行状态更新成功", g.Map{
		"algorithmId": algorithmId,
		"runStatus":   runStatus,
		"configPath":  configPath,
	})

	return nil
}

// queryAlgorithmFromDB 从数据库查询算法记录
func (s *AlgorithmConfigService) queryAlgorithmFromDB(ctx context.Context, algorithmId string) (*entity.Algorithm, error) {
	var algorithm entity.Algorithm

	// 根据algorithmId查询算法记录
	err := dao.Algorithm.Ctx(ctx).Where("algorithm_id = ?", algorithmId).Scan(&algorithm)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %v", err)
	}

	// 检查是否找到记录
	if algorithm.Id == 0 {
		return nil, nil // 未找到记录
	}

	return &algorithm, nil
}

// BuildConfigPath 构建配置文件路径 (公开方法用于测试)
func (s *AlgorithmConfigService) BuildConfigPath(algorithmId, algorithmVersionId string) string {
	return s.buildConfigPath(algorithmId, algorithmVersionId)
}

// buildConfigPath 构建配置文件路径
func (s *AlgorithmConfigService) buildConfigPath(algorithmId, algorithmVersionId string) string {
	var basePath string

	// 根据操作系统选择基础路径
	if runtime.GOOS == "windows" {
		// Windows环境：使用当前工作目录下的runtime目录
		basePath = "runtime/algorithm"
	} else {
		// Linux/Unix环境：使用/usr/runtime/algorithm
		basePath = "/usr/runtime/algorithm"
	}

	// 构建完整路径: basePath/algorithmId/algorithmVersionId/config.yaml
	configPath := filepath.Join(basePath, algorithmId, algorithmVersionId, "config.yaml")

	return configPath
}

// updateConfigFile 更新配置文件中的runStatus
func (s *AlgorithmConfigService) updateConfigFile(ctx context.Context, configPath string, newRunStatus int) error {
	// 1. 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 2. 读取现有配置文件
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 3. 解析现有YAML配置
	var config AlgorithmConfig
	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return fmt.Errorf("解析YAML配置失败: %v", err)
	}

	// 4. 记录原有状态用于日志
	oldRunStatus := config.Algo.RunStatus

	// 5. 更新runStatus
	config.Algo.RunStatus = newRunStatus

	// 6. 重新序列化为YAML
	newData, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("序列化YAML配置失败: %v", err)
	}

	// 7. 写入文件
	err = os.WriteFile(configPath, newData, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	g.Log().Info(ctx, "配置文件更新成功", g.Map{
		"configPath":   configPath,
		"oldRunStatus": oldRunStatus,
		"newRunStatus": newRunStatus,
	})

	return nil
}

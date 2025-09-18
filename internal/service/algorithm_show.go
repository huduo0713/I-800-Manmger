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

// AlgorithmShowService 算法查询服务
type AlgorithmShowService struct{}

// NewAlgorithmShowService 创建算法查询服务实例
func NewAlgorithmShowService() *AlgorithmShowService {
	return &AlgorithmShowService{}
}

// GetAlgorithmList 获取算法列表及其运行状态
func (s *AlgorithmShowService) GetAlgorithmList(ctx context.Context) ([]AlgorithmShowResponseData, error) {
	// 1. 从数据库查询所有算法记录
	algorithms, err := s.queryAlgorithmsFromDB(ctx)
	if err != nil {
		g.Log().Error(ctx, "查询算法数据失败", g.Map{
			"error": err,
		})
		return nil, fmt.Errorf("数据库查询失败: %v", err)
	}

	// 2. 构建响应数据列表
	responseData := make([]AlgorithmShowResponseData, 0, len(algorithms))

	for _, algorithm := range algorithms {
		// 3. 读取每个算法的运行状态
		runStatus := s.readAlgorithmRunStatus(ctx, algorithm.AlgorithmId, algorithm.AlgorithmVersionId)

		responseData = append(responseData, AlgorithmShowResponseData{
			AlgorithmName:    algorithm.AlgorithmName,
			AlgorithmId:      algorithm.AlgorithmId,
			AlgorithmVersion: algorithm.AlgorithmVersion,
			RunStatus:        runStatus,
		})
	}

	g.Log().Info(ctx, "算法列表查询完成", g.Map{
		"count": len(responseData),
	})

	return responseData, nil
}

// queryAlgorithmsFromDB 从数据库查询算法记录
func (s *AlgorithmShowService) queryAlgorithmsFromDB(ctx context.Context) ([]entity.Algorithm, error) {
	var algorithms []entity.Algorithm

	// 查询所有算法记录
	err := dao.Algorithm.Ctx(ctx).Scan(&algorithms)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %v", err)
	}

	return algorithms, nil
}

// readAlgorithmRunStatus 读取算法运行状态
func (s *AlgorithmShowService) readAlgorithmRunStatus(ctx context.Context, algorithmId, algorithmVersionId string) int {
	// 构建config.yaml文件路径
	configPath := s.buildConfigPath(algorithmId, algorithmVersionId)

	// 读取并解析config.yaml文件
	runStatus, err := s.parseConfigFile(ctx, configPath)
	if err != nil {
		g.Log().Warning(ctx, "读取算法配置文件失败，使用默认状态", g.Map{
			"algorithmId":        algorithmId,
			"algorithmVersionId": algorithmVersionId,
			"configPath":         configPath,
			"error":              err,
		})
		// 默认返回停止状态
		return 0
	}

	g.Log().Debug(ctx, "读取算法运行状态成功", g.Map{
		"algorithmId":        algorithmId,
		"algorithmVersionId": algorithmVersionId,
		"runStatus":          runStatus,
	})

	return runStatus
}

// buildConfigPath 构建配置文件路径
func (s *AlgorithmShowService) buildConfigPath(algorithmId, algorithmVersionId string) string {
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

// parseConfigFile 解析config.yaml配置文件
func (s *AlgorithmShowService) parseConfigFile(ctx context.Context, configPath string) (int, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取文件内容
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return 0, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML内容
	var config AlgorithmConfig
	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return 0, fmt.Errorf("解析YAML配置失败: %v", err)
	}

	// 验证runStatus值的有效性
	runStatus := config.Algo.RunStatus
	if runStatus != 0 && runStatus != 1 {
		g.Log().Warning(ctx, "配置文件中runStatus值无效，使用默认值0", g.Map{
			"configPath": configPath,
			"runStatus":  runStatus,
		})
		return 0, nil
	}

	return runStatus, nil
}

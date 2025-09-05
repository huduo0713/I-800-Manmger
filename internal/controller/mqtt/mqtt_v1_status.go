package mqtt

import (
	"context"
	v1 "demo/api/mqtt/v1"
	"demo/internal/service"
)

// GetStatus 获取MQTT连接状态
func (c *ControllerV1) GetStatus(ctx context.Context, req *v1.GetStatusReq) (res *v1.GetStatusRes, err error) {
	// 获取MQTT服务实例
	mqttSvc := service.Mqtt()

	// 获取状态信息
	status := mqttSvc.GetStatus()

	// 处理服务器信息
	broker := "unknown"
	if servers, ok := status["servers"].([]string); ok && len(servers) > 0 {
		broker = servers[0]
	}

	// 处理客户端ID
	clientId := "unknown"
	if id, ok := status["client_id"].(string); ok {
		clientId = id
	}

	return &v1.GetStatusRes{
		Connected: status["connected"].(bool),
		ClientId:  clientId,
		Broker:    broker,
	}, nil
}

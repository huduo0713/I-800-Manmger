package mqtt

import (
	"context"
	v1 "demo/api/mqtt/v1"
	"demo/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Publish 发布MQTT消息
func (c *ControllerV1) Publish(ctx context.Context, req *v1.PublishReq) (res *v1.PublishRes, err error) {
	// 获取MQTT服务实例
	mqttSvc := service.Mqtt()

	// 检查连接状态
	if !mqttSvc.IsConnected() {
		return nil, gerror.New("MQTT client is not connected")
	}

	// 设置默认值
	qos := byte(0)
	if req.Qos != nil {
		qos = byte(*req.Qos)
	}

	retained := false
	if req.Retained != nil {
		retained = *req.Retained
	}

	// 发布消息
	err = mqttSvc.Publish(req.Topic, qos, retained, req.Payload)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to publish message to topic: %s", req.Topic)
	}

	return &v1.PublishRes{
		Success: true,
		Message: "Message published successfully",
	}, nil
}

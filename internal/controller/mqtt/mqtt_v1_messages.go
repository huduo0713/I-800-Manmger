package mqtt

import (
	"context"
	v1 "demo/api/mqtt/v1"
	"demo/internal/service"
)

// GetMessages 获取接收到的MQTT消息列表
func (c *ControllerV1) GetMessages(ctx context.Context, req *v1.GetMessagesReq) (res *v1.GetMessagesRes, err error) {
	// 获取MQTT服务实例
	mqttSvc := service.Mqtt()

	// 设置默认值
	limit := 20
	if req.Limit != nil {
		limit = *req.Limit
	}

	// 获取消息列表
	messages := mqttSvc.GetMessages(req.Topic, limit)

	// 转换为API响应格式
	var apiMessages []v1.MqttMessage
	for _, msg := range messages {
		apiMessages = append(apiMessages, v1.MqttMessage{
			Topic:     msg.Topic,
			Payload:   msg.Payload,
			Qos:       msg.Qos,
			Retained:  msg.Retained,
			Timestamp: msg.CreatedAt,
		})
	}

	return &v1.GetMessagesRes{
		Messages: apiMessages,
		Total:    len(apiMessages),
	}, nil
}

package mqtt

import (
	"context"
	v1 "demo/api/mqtt/v1"
	"demo/internal/service"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// Subscribe 订阅MQTT主题
func (c *ControllerV1) Subscribe(ctx context.Context, req *v1.SubscribeReq) (res *v1.SubscribeRes, err error) {
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

	// 定义消息处理回调函数
	callback := func(client mqtt.Client, msg mqtt.Message) {
		g.Log().Infof(gctx.New(), "Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
	}

	// 订阅主题
	err = mqttSvc.Subscribe(req.Topic, qos, callback)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to subscribe to topic: %s", req.Topic)
	}

	return &v1.SubscribeRes{
		Success: true,
		Message: "Subscribed to topic successfully",
	}, nil
}

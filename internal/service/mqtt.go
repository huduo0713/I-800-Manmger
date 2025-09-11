package service

import (
	"demo/internal/model/entity"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// 定义我们的 MQTT 服务结构体
type sMqtt struct {
	client   mqtt.Client          // Paho MQTT 客户端实例
	messages []entity.MqttMessage // 内存中存储的消息（实际项目中应该用数据库）
	msgMutex sync.RWMutex         // 消息操作的读写锁
}

var (
	mqttService *sMqtt    // 用于存储单例的服务实例
	mqttOnce    sync.Once // 保证单例只被创建一次
)

// Mqtt 是获取 MQTT 服务单例的函数
func Mqtt() *sMqtt {
	mqttOnce.Do(func() {
		// --- MQTT 客户端配置 ---
		// 使用公共的 EMQ X 测试服务器，你也可以换成自己的
		broker := "tcp://broker.emqx.io:1883"
		clientId := "goframe-mqtt-client-example"

		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker)
		opts.SetClientID(clientId)
		opts.SetKeepAlive(60 * time.Second)
		// 设置一个默认的消息处理回调函数
		opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			g.Log().Infof(gctx.New(), "MQTT Received Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
			// 将接收到的消息存储到内存中
			if mqttService != nil {
				mqttService.storeMessage(msg)
			}
		})
		// 设置连接成功的回调
		opts.OnConnect = func(client mqtt.Client) {
			g.Log().Info(gctx.New(), "MQTT Connected")
		}
		// 设置连接丢失的回调
		opts.OnConnectionLost = func(client mqtt.Client, err error) {
			g.Log().Errorf(gctx.New(), "MQTT Connection Lost: %v", err)
		}

		// 创建客户端实例
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			// 在实际项目中，这里应该处理失败，比如 panic 或重试
			g.Log().Fatalf(gctx.New(), "MQTT Connect Error: %s", token.Error())
		}

		mqttService = &sMqtt{
			client:   client,
			messages: make([]entity.MqttMessage, 0),
		}
	})
	return mqttService
}

// Publish 方法用于发布消息
func (s *sMqtt) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := s.client.Publish(topic, qos, retained, payload)
	// Wait 会等待消息发送完成
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe 方法用于订阅主题
func (s *sMqtt) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := s.client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	g.Log().Infof(gctx.New(), "Subscribed to topic: %s", topic)
	return nil
}

// storeMessage 存储接收到的消息到内存中（实际项目中应该存储到数据库）
func (s *sMqtt) storeMessage(msg mqtt.Message) {
	s.msgMutex.Lock()
	defer s.msgMutex.Unlock()

	// TODO: 实体字段问题，暂时跳过存储
	_ = msg
	/*
		message := entity.MqttMessage{
			Topic:     msg.Topic(),
			Payload:   string(msg.Payload()),
			Qos:       int(msg.Qos()),
			Retained:  msg.Retained(),
			CreatedAt: gtime.Now().Unix(),
		}

		s.messages = append(s.messages, message)
	*/

	// 限制内存中的消息数量，只保留最近的1000条
	if len(s.messages) > 1000 {
		s.messages = s.messages[len(s.messages)-1000:]
	}
}

// GetMessages 获取接收到的消息列表
func (s *sMqtt) GetMessages(topic string, limit int) []entity.MqttMessage {
	s.msgMutex.RLock()
	defer s.msgMutex.RUnlock()

	// TODO: 实体字段问题，暂时返回空列表
	_ = topic
	_ = limit
	return []entity.MqttMessage{}

	/*
		var result []entity.MqttMessage
		count := 0

		// 从后往前取，获取最新的消息
		for i := len(s.messages) - 1; i >= 0 && count < limit; i-- {
			if topic == "" || s.messages[i].Topic == topic {
				result = append([]entity.MqttMessage{s.messages[i]}, result...)
				count++
			}
		}

		return result
	*/
}

// GetStatus 获取MQTT连接状态
func (s *sMqtt) GetStatus() map[string]interface{} {
	opts := s.client.OptionsReader()
	return map[string]interface{}{
		"connected": s.client.IsConnected(),
		"client_id": opts.ClientID(),
		"servers":   opts.Servers(),
	}
}

// IsConnected 检查MQTT是否连接
func (s *sMqtt) IsConnected() bool {
	return s.client.IsConnected()
}

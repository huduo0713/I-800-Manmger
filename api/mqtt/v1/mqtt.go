package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PublishReq 发布MQTT消息请求
type PublishReq struct {
	g.Meta   `path:"/mqtt/publish" method:"post" tags:"MQTT" summary:"Publish MQTT message"`
	Topic    string      `v:"required" dc:"MQTT topic"`
	Payload  interface{} `v:"required" dc:"Message payload"`
	Qos      *int        `v:"in:0,1,2" dc:"Quality of Service level (0,1,2)" default:"0"`
	Retained *bool       `dc:"Retained message flag" default:"false"`
}

type PublishRes struct {
	Success bool   `json:"success" dc:"Publish result"`
	Message string `json:"message" dc:"Result message"`
}

// SubscribeReq 订阅MQTT主题请求
type SubscribeReq struct {
	g.Meta `path:"/mqtt/subscribe" method:"post" tags:"MQTT" summary:"Subscribe to MQTT topic"`
	Topic  string `v:"required" dc:"MQTT topic to subscribe"`
	Qos    *int   `v:"in:0,1,2" dc:"Quality of Service level (0,1,2)" default:"0"`
}

type SubscribeRes struct {
	Success bool   `json:"success" dc:"Subscribe result"`
	Message string `json:"message" dc:"Result message"`
}

// GetStatusReq 获取MQTT连接状态请求
type GetStatusReq struct {
	g.Meta `path:"/mqtt/status" method:"get" tags:"MQTT" summary:"Get MQTT connection status"`
}

type GetStatusRes struct {
	Connected bool   `json:"connected" dc:"Connection status"`
	ClientId  string `json:"client_id" dc:"MQTT client ID"`
	Broker    string `json:"broker" dc:"MQTT broker address"`
}

// GetMessagesReq 获取接收到的消息列表
type GetMessagesReq struct {
	g.Meta `path:"/mqtt/messages" method:"get" tags:"MQTT" summary:"Get received MQTT messages"`
	Topic  string `v:"" dc:"Filter by topic (optional)"`
	Limit  *int   `v:"min:1,max:100" dc:"Message limit" default:"20"`
}

type GetMessagesRes struct {
	Messages []MqttMessage `json:"messages" dc:"MQTT messages list"`
	Total    int           `json:"total" dc:"Total message count"`
}

// MqttMessage MQTT消息结构
type MqttMessage struct {
	Topic     string `json:"topic" dc:"Message topic"`
	Payload   string `json:"payload" dc:"Message payload"`
	Qos       int    `json:"qos" dc:"Quality of Service"`
	Retained  bool   `json:"retained" dc:"Retained flag"`
	Timestamp int64  `json:"timestamp" dc:"Receive timestamp"`
}

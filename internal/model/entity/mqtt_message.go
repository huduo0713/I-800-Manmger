// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MqttMessage is the golang structure for table mqtt_message.
type MqttMessage struct {
	Id        uint   `json:"id"        orm:"id"        description:"message id"`         // message id
	Topic     string `json:"topic"     orm:"topic"     description:"MQTT topic"`         // MQTT topic
	Payload   string `json:"payload"   orm:"payload"   description:"message payload"`    // message payload
	Qos       int    `json:"qos"       orm:"qos"       description:"quality of service"` // quality of service
	Retained  bool   `json:"retained"  orm:"retained"  description:"retained flag"`      // retained flag
	CreatedAt int64  `json:"created_at" orm:"created_at" description:"create time"`      // create time
}

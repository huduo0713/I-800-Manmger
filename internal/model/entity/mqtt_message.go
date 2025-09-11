// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MqttMessage is the golang structure for table mqtt_message.
type MqttMessage struct {
	Id        int         `json:"id"        orm:"id"         description:""` //
	Topic     string      `json:"topic"     orm:"topic"      description:""` //
	Payload   string      `json:"payload"   orm:"payload"    description:""` //
	Qos       int         `json:"qos"       orm:"qos"        description:""` //
	Retained  int         `json:"retained"  orm:"retained"   description:""` //
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""` //
}

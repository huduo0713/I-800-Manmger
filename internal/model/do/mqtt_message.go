// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MqttMessage is the golang structure of table mqtt_message for DAO operations like Where/Data.
type MqttMessage struct {
	g.Meta    `orm:"table:mqtt_message, do:true"`
	Id        interface{} //
	Topic     interface{} //
	Payload   interface{} //
	Qos       interface{} //
	Retained  interface{} //
	CreatedAt *gtime.Time //
}

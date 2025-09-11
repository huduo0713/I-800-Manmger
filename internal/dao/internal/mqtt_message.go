// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MqttMessageDao is the data access object for the table mqtt_message.
type MqttMessageDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MqttMessageColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MqttMessageColumns defines and stores column names for the table mqtt_message.
type MqttMessageColumns struct {
	Id        string //
	Topic     string //
	Payload   string //
	Qos       string //
	Retained  string //
	CreatedAt string //
}

// mqttMessageColumns holds the columns for the table mqtt_message.
var mqttMessageColumns = MqttMessageColumns{
	Id:        "id",
	Topic:     "topic",
	Payload:   "payload",
	Qos:       "qos",
	Retained:  "retained",
	CreatedAt: "created_at",
}

// NewMqttMessageDao creates and returns a new DAO object for table data access.
func NewMqttMessageDao(handlers ...gdb.ModelHandler) *MqttMessageDao {
	return &MqttMessageDao{
		group:    "default",
		table:    "mqtt_message",
		columns:  mqttMessageColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MqttMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MqttMessageDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MqttMessageDao) Columns() MqttMessageColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MqttMessageDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MqttMessageDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *MqttMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

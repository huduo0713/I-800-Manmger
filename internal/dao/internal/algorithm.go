// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AlgorithmDao is the data access object for the table algorithm.
type AlgorithmDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AlgorithmColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AlgorithmColumns defines and stores column names for the table algorithm.
type AlgorithmColumns struct {
	Id                 string //
	AlgorithmId        string //
	AlgorithmName      string //
	AlgorithmVersion   string //
	AlgorithmVersionId string //
	AlgorithmDataUrl   string //
	FileSize           string //
	Md5                string //
	LocalPath          string //
}

// algorithmColumns holds the columns for the table algorithm.
var algorithmColumns = AlgorithmColumns{
	Id:                 "id",
	AlgorithmId:        "algorithm_id",
	AlgorithmName:      "algorithm_name",
	AlgorithmVersion:   "algorithm_version",
	AlgorithmVersionId: "algorithm_version_id",
	AlgorithmDataUrl:   "algorithm_data_url",
	FileSize:           "file_size",
	Md5:                "md5",
	LocalPath:          "local_path",
}

// NewAlgorithmDao creates and returns a new DAO object for table data access.
func NewAlgorithmDao(handlers ...gdb.ModelHandler) *AlgorithmDao {
	return &AlgorithmDao{
		group:    "default",
		table:    "algorithm",
		columns:  algorithmColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AlgorithmDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AlgorithmDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AlgorithmDao) Columns() AlgorithmColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AlgorithmDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AlgorithmDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AlgorithmDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

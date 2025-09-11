// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package algorithm

import (
	"context"

	"demo/api/algorithm/v1"
)

type IAlgorithmV1 interface {
	Add(ctx context.Context, req *v1.AddReq) (res *v1.AddRes, err error)
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
	GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
}

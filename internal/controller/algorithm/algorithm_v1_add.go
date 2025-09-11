package algorithm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"demo/api/algorithm/v1"
)

func (c *ControllerV1) Add(ctx context.Context, req *v1.AddReq) (res *v1.AddRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}

package v1

import (
	"demo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// AddReq 添加算法请求 (对应算法下发payload)
type AddReq struct {
	g.Meta             `path:"/algorithm" method:"post" tags:"Algorithm" summary:"Add algorithm from payload"`
	CmdId              string `json:"cmdId" v:"required" dc:"Command ID"`
	Version            string `json:"version" v:"required" dc:"Protocol version"`
	Method             string `json:"method" v:"required" dc:"Method name"`
	Timestamp          string `json:"timestamp" v:"required" dc:"Timestamp"`
	AlgorithmId        string `json:"algorithmId" v:"required" dc:"Algorithm unique ID"`
	AlgorithmName      string `json:"algorithmName" v:"required" dc:"Algorithm name"`
	AlgorithmVersion   string `json:"algorithmVersion" v:"required" dc:"Algorithm version"`
	AlgorithmVersionId string `json:"algorithmVersionId" v:"required" dc:"Algorithm version ID"`
	AlgorithmDataUrl   string `json:"algorithmDataUrl" v:"required|url" dc:"Algorithm download URL"`
	FileSize           int64  `json:"fileSize" v:"required|min:1" dc:"File size in bytes"`
	Md5                string `json:"md5" v:"required|length:32,32" dc:"MD5 checksum"`
}

type AddRes struct {
	Id      int64  `json:"id" dc:"Algorithm record ID"`
	Success bool   `json:"success" dc:"Operation result"`
	Message string `json:"message" dc:"Result message"`
}

// GetListReq 获取算法列表请求
type GetListReq struct {
	g.Meta   `path:"/algorithm" method:"get" tags:"Algorithm" summary:"Get algorithm list"`
	Name     string `v:"" dc:"Algorithm name filter"`
	Page     *int   `v:"min:1" dc:"Page number" default:"1"`
	PageSize *int   `v:"min:1,max:100" dc:"Page size" default:"20"`
}

type GetListRes struct {
	List  []entity.Algorithm `json:"list" dc:"Algorithm list"`
	Total int                `json:"total" dc:"Total count"`
	Page  int                `json:"page" dc:"Current page"`
}

// GetOneReq 获取单个算法信息请求
type GetOneReq struct {
	g.Meta      `path:"/algorithm/{id}" method:"get" tags:"Algorithm" summary:"Get algorithm by ID"`
	Id          int64  `v:"required" dc:"Algorithm record ID"`
	AlgorithmId string `v:"" dc:"Algorithm unique ID (alternative)"`
}

type GetOneRes struct {
	*entity.Algorithm
}

// UpdateReq 更新算法信息请求
type UpdateReq struct {
	g.Meta    `path:"/algorithm/{id}" method:"put" tags:"Algorithm" summary:"Update algorithm info"`
	Id        int64  `v:"required" dc:"Algorithm record ID"`
	LocalPath string `v:"" dc:"Local storage path"`
}

type UpdateRes struct {
	Success bool   `json:"success" dc:"Update result"`
	Message string `json:"message" dc:"Result message"`
}

// DeleteReq 删除算法请求
type DeleteReq struct {
	g.Meta `path:"/algorithm/{id}" method:"delete" tags:"Algorithm" summary:"Delete algorithm"`
	Id     int64 `v:"required" dc:"Algorithm record ID"`
}

type DeleteRes struct {
	Success bool   `json:"success" dc:"Delete result"`
	Message string `json:"message" dc:"Result message"`
}

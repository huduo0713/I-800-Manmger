package service

import "fmt"

// 通用请求结构体定义

// AlgorithmAddRequest 算法添加请求结构
type AlgorithmAddRequest struct {
	CmdId     string `json:"cmdId"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Params    struct {
		AlgorithmId        string  `json:"algorithmId"`
		AlgorithmName      string  `json:"algorithmName"`
		AlgorithmVersion   string  `json:"algorithmVersion"`
		AlgorithmVersionId string  `json:"algorithmVersionId"`
		AlgorithmDataUrl   string  `json:"algorithmDataUrl"`
		FileSize           float64 `json:"fileSize"` // 改为float64以兼容JSON中的浮点数格式
		LastModifyTime     string  `json:"lastModifyTime"`
		Md5                string  `json:"md5"`
	} `json:"params"`
}

// AlgorithmDeleteRequest 算法删除请求结构
type AlgorithmDeleteRequest struct {
	CmdId     string `json:"cmdId"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Params    struct {
		AlgorithmId string `json:"algorithmId"`
	} `json:"params"`
}

// AlgorithmShowRequest 算法查询请求结构
type AlgorithmShowRequest struct {
	CmdId     string      `json:"cmdId"`
	Version   string      `json:"version"`
	Method    string      `json:"method"`
	Timestamp string      `json:"timestamp"`
	Params    interface{} `json:"params"` // 查询请求params为null
}

// AlgorithmConfigRequest 算法启停配置请求结构
type AlgorithmConfigRequest struct {
	CmdId     string `json:"cmdId"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Params    struct {
		AlgorithmId string `json:"algorithmId"` // 算法标识
		RunStatus   int    `json:"runStatus"`   // 设置为0标识关闭；设置为1标识开启
	} `json:"params"`
}

// AlgorithmShowResponseData 算法查询响应数据结构
type AlgorithmShowResponseData struct {
	AlgorithmName    string `json:"algorithmName"`    // 算法名称
	AlgorithmId      string `json:"algorithmId"`      // 算法标识
	AlgorithmVersion string `json:"algorithmVersion"` // 算法版本
	RunStatus        int    `json:"runStatus"`        // 算法运行状态，1-运行；0-停止
}

// AlgorithmConfig config.yaml文件结构
type AlgorithmConfig struct {
	Algo struct {
		RunStatus int `yaml:"runStatus"` // 1表示开启，0表示关闭
	} `yaml:"algo"`
}

// AlgorithmReply 算法操作响应结构
type AlgorithmReply struct {
	CmdId     string      `json:"cmdId"`
	Version   string      `json:"version"`
	Method    string      `json:"method"`
	Timestamp string      `json:"timestamp"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

// 错误码定义
const (
	CodeSuccess           = 0
	CodeDownloadFailed    = 1001
	CodeMd5CheckFailed    = 1002
	CodeFileSystemError   = 1003
	CodeDatabaseError     = 1004
	CodeInvalidParams     = 1005
	CodeAlgorithmNotFound = 1006
	CodeVersionExists     = 1007 // 算法版本已存在
)

// AlgorithmVersionExistsError 算法版本已存在错误类型
type AlgorithmVersionExistsError struct {
	AlgorithmId string
	Version     string
	LocalPath   string
}

func (e *AlgorithmVersionExistsError) Error() string {
	return fmt.Sprintf("算法版本已存在，algorithmId: %s, version: %s", e.AlgorithmId, e.Version)
}

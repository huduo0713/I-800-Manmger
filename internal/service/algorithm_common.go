package service

// 通用请求结构体定义

// AlgorithmAddRequest 算法添加请求结构
type AlgorithmAddRequest struct {
	CmdId     string `json:"cmdId"`
	Version   string `json:"version"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	Params    struct {
		AlgorithmId        string `json:"algorithmId"`
		AlgorithmName      string `json:"algorithmName"`
		AlgorithmVersion   string `json:"algorithmVersion"`
		AlgorithmVersionId string `json:"algorithmVersionId"`
		AlgorithmDataUrl   string `json:"algorithmDataUrl"`
		FileSize           int64  `json:"fileSize"`
		LastModifyTime     string `json:"lastModifyTime"`
		Md5                string `json:"md5"`
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
)

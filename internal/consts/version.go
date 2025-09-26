package consts

// 版本信息常量
const (
	AppName    = "Edge Device Manager"
	AppVersion = "v1.3.0"
	Author     = "GoFrame Team"
)

// 编译时注入的变量 (通过 go build -ldflags 设置)
var (
	BuildTime = "unknown"
	GitCommit = "unknown"
	GitBranch = "unknown"
	GoVersion = "unknown"
)

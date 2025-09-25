package service

import (
	"net"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// PortChecker 端口检测服务
type PortChecker struct{}

// NewPortChecker 创建端口检测服务
func NewPortChecker() *PortChecker {
	return &PortChecker{}
}

// IsPortListening 检测指定端口是否正在监听
// 方法1：尝试连接本地端口（推荐，开销最小）
func (p *PortChecker) IsPortListening(port int) int {
	ctx := gctx.New()

	// 尝试连接本地端口
	address := g.NewVar(port).String()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+address, 100*time.Millisecond)
	if err != nil {
		g.Log().Debugf(ctx, "端口 %d 未监听: %v", port, err)
		return 0
	}

	conn.Close()
	g.Log().Debugf(ctx, "端口 %d 正在监听", port)
	return 1
}

// IsPortListeningFast 快速检测端口状态（备选方案）
// 方法2：使用更短的超时时间
func (p *PortChecker) IsPortListeningFast(port int) int {
	address := g.NewVar(port).String()

	// 使用极短超时时间，减少等待开销
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+address, 50*time.Millisecond)
	if err != nil {
		return 0
	}

	conn.Close()
	return 1
}

// GetRuntimeStatus 获取runtime状态（从配置读取端口号）
func (p *PortChecker) GetRuntimeStatus() int {
	ctx := gctx.New()
	// 从配置文件读取runtime端口号，默认为1231
	runtimePort := g.Cfg().MustGet(ctx, "device.runtime.port", 1231).Int()
	return p.IsPortListening(runtimePort)
}

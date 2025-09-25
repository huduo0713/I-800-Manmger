package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name string `json:"name"` // 网卡名称
	MAC  string `json:"mac"`  // MAC地址 (格式: AA-BB-CC-DD-EE-FF)
	IP   string `json:"ip"`   // IP地址
}

// NetworkDetectionService 网络检测服务
type NetworkDetectionService struct {
	ctx context.Context
}

// NewNetworkDetectionService 创建网络检测服务
func NewNetworkDetectionService() *NetworkDetectionService {
	return &NetworkDetectionService{
		ctx: gctx.New(),
	}
}

// DetectAvailableNetwork 检测可用的网络接口
func (s *NetworkDetectionService) DetectAvailableNetwork() (*NetworkInterface, error) {
	ctx := s.ctx

	// 获取配置
	mode := g.Cfg().MustGet(ctx, "device.network.mode", "auto").String()
	brokerUrl := g.Cfg().MustGet(ctx, "mqtt.broker").String()

	g.Log().Infof(ctx, "🌐 开始网卡检测, 模式: %s", mode)

	switch mode {
	case "manual":
		return s.detectManualNetwork(ctx)
	case "config":
		return s.detectConfigPriorityNetwork(ctx, brokerUrl)
	default: // "auto"
		return s.detectAutoNetwork(ctx, brokerUrl)
	}
}

// detectManualNetwork 手动指定网络接口
func (s *NetworkDetectionService) detectManualNetwork(ctx context.Context) (*NetworkInterface, error) {
	iface := g.Cfg().MustGet(ctx, "device.network.interface").String()
	mac := g.Cfg().MustGet(ctx, "device.network.mac").String()
	ip := g.Cfg().MustGet(ctx, "device.network.ip").String()

	if iface == "" || mac == "" || ip == "" {
		return nil, fmt.Errorf("手动模式下必须指定 interface, mac, ip")
	}

	g.Log().Infof(ctx, "✅ 使用手动配置网卡: %s, MAC: %s, IP: %s", iface, mac, ip)

	return &NetworkInterface{
		Name: iface,
		MAC:  mac,
		IP:   ip,
	}, nil
}

// detectConfigPriorityNetwork 配置优先级网络检测
func (s *NetworkDetectionService) detectConfigPriorityNetwork(ctx context.Context, brokerUrl string) (*NetworkInterface, error) {
	priorities := g.Cfg().MustGet(ctx, "device.network.priority").Strings()
	if len(priorities) == 0 {
		priorities = []string{"eth0", "ens", "enp", "wlan", "wlp"}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %v", err)
	}

	// 按优先级顺序检测
	for _, priority := range priorities {
		for _, iface := range interfaces {
			if s.matchesPattern(iface.Name, priority) && s.isValidInterface(iface) {
				if netIface, err := s.testInterface(ctx, iface, brokerUrl); err == nil {
					g.Log().Infof(ctx, "✅ 选择网卡: %s (匹配优先级: %s), MAC: %s, IP: %s",
						netIface.Name, priority, netIface.MAC, netIface.IP)
					return netIface, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("未找到可用的网络接口")
}

// detectAutoNetwork 自动检测网络接口
func (s *NetworkDetectionService) detectAutoNetwork(ctx context.Context, brokerUrl string) (*NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口列表失败: %v", err)
	}

	g.Log().Infof(ctx, "🔍 开始自动检测可用网卡，共找到 %d 个网络接口", len(interfaces))

	// 遍历所有网络接口
	for _, iface := range interfaces {
		if !s.isValidInterface(iface) {
			g.Log().Debugf(ctx, "⏭️ 跳过无效网卡: %s (flags: %v)", iface.Name, iface.Flags)
			continue
		}

		if netIface, err := s.testInterface(ctx, iface, brokerUrl); err == nil {
			g.Log().Infof(ctx, "✅ 自动选择网卡: %s, MAC: %s, IP: %s", netIface.Name, netIface.MAC, netIface.IP)
			return netIface, nil
		} else {
			g.Log().Debugf(ctx, "❌ 网卡不可用: %s, 原因: %v", iface.Name, err)
		}
	}

	return nil, fmt.Errorf("未找到任何可用的网络接口")
}

// isValidInterface 检查网络接口是否有效
func (s *NetworkDetectionService) isValidInterface(iface net.Interface) bool {
	// 必须是UP状态
	if iface.Flags&net.FlagUp == 0 {
		return false
	}

	// 跳过回环接口
	if iface.Flags&net.FlagLoopback != 0 {
		return false
	}

	// 跳过虚拟网卡
	name := strings.ToLower(iface.Name)
	virtualPatterns := []string{
		"docker", "veth", "br-", "virbr", "vmnet", "vbox", "tap", "tun",
		"lo", "dummy", "bond", "team", "vlan", "sit", "gre", "ipip",
	}

	for _, pattern := range virtualPatterns {
		if strings.Contains(name, pattern) {
			return false
		}
	}

	// 检查是否有MAC地址
	if len(iface.HardwareAddr) == 0 {
		return false
	}

	return true
}

// testInterface 测试网络接口是否可用
func (s *NetworkDetectionService) testInterface(ctx context.Context, iface net.Interface, brokerUrl string) (*NetworkInterface, error) {
	// 获取IP地址
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("获取IP地址失败: %v", err)
	}

	var ipv4 string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				ipv4 = ip4.String()
				break
			}
		}
	}

	if ipv4 == "" {
		return nil, fmt.Errorf("未找到有效的IPv4地址")
	}

	// 格式化MAC地址为大写并用连字符分隔
	mac := strings.ToUpper(iface.HardwareAddr.String())
	mac = strings.ReplaceAll(mac, ":", "-")

	// 测试连通性
	if err := s.testConnectivity(ctx, ipv4, brokerUrl); err != nil {
		return nil, fmt.Errorf("连通性测试失败: %v", err)
	}

	return &NetworkInterface{
		Name: iface.Name,
		MAC:  mac,
		IP:   ipv4,
	}, nil
}

// testConnectivity 测试网络连通性
func (s *NetworkDetectionService) testConnectivity(ctx context.Context, localIP, brokerUrl string) error {
	// 解析broker地址
	host, port, err := net.SplitHostPort(strings.TrimPrefix(brokerUrl, "tcp://"))
	if err != nil {
		return fmt.Errorf("解析broker地址失败: %v", err)
	}

	// 获取超时配置
	timeout := g.Cfg().MustGet(ctx, "device.network.connectivity.timeout", 3).Duration() * time.Second
	retries := g.Cfg().MustGet(ctx, "device.network.connectivity.retries", 2).Int()

	// 多次尝试连接测试
	for i := 0; i <= retries; i++ {
		if err := s.dialWithLocalIP(localIP, net.JoinHostPort(host, port), timeout); err == nil {
			g.Log().Debugf(ctx, "✅ 连通性测试成功: %s -> %s", localIP, brokerUrl)
			return nil
		}

		if i < retries {
			time.Sleep(100 * time.Millisecond) // 短暂等待后重试
		}
	}

	return fmt.Errorf("连接测试失败，重试 %d 次后仍无法连接", retries+1)
}

// dialWithLocalIP 使用指定本地IP进行连接测试
func (s *NetworkDetectionService) dialWithLocalIP(localIP, target string, timeout time.Duration) error {
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: net.ParseIP(localIP)},
		Timeout:   timeout,
	}

	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}

// matchesPattern 检查接口名是否匹配模式
func (s *NetworkDetectionService) matchesPattern(ifaceName, pattern string) bool {
	return strings.HasPrefix(strings.ToLower(ifaceName), strings.ToLower(pattern))
}

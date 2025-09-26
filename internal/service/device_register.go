package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/guid"
)

// DeviceRegisterService 设备注册服务
type DeviceRegisterService struct {
	ctx           context.Context
	mqttClient    mqtt.Client
	networkIface  *NetworkInterface
	deviceId      string
	retryCount    int
	maxRetries    int
	retryInterval time.Duration
	portChecker   *PortChecker // 端口检测器
}

// DeviceRegisterRequest 设备注册请求结构
type DeviceRegisterRequest struct {
	CmdId     string                    `json:"cmdId"`
	Version   string                    `json:"version"`
	Method    string                    `json:"method"`
	Timestamp string                    `json:"timestamp"`
	Data      DeviceRegisterRequestData `json:"data"`
}

// DeviceRegisterRequestData 设备注册请求数据
type DeviceRegisterRequestData struct {
	DeviceModule     string `json:"deviceModule"`     // 设备模块名
	DeviceId         string `json:"deviceId"`         // 设备ID (MAC地址)
	HeartBeat        int    `json:"heartBeat"`        // 心跳周期(秒)
	IP               string `json:"IP"`               // 设备IP地址
	RuntimeStatus    int    `json:"runtimeStatus"`    // Runtime进程状态 (1-正常运行，0-停止)
	OpcuaServerPort  int    `json:"opcuaServerPort"`  // OPC UA服务器端口
}

// NewDeviceRegisterService 创建设备注册服务
func NewDeviceRegisterService(mqttClient mqtt.Client, networkIface *NetworkInterface) *DeviceRegisterService {
	ctx := gctx.New()

	// 从配置获取重试参数
	maxRetries := g.Cfg().MustGet(ctx, "mqtt.register.maxRetries", 3).Int()
	retryInterval := g.Cfg().MustGet(ctx, "mqtt.register.retryInterval", 30).Duration() * time.Second

	return &DeviceRegisterService{
		ctx:           ctx,
		mqttClient:    mqttClient,
		networkIface:  networkIface,
		deviceId:      networkIface.MAC, // deviceId就是MAC地址
		maxRetries:    maxRetries,
		retryInterval: retryInterval,
		portChecker:   NewPortChecker(), // 初始化端口检测器
	}
}

// Register 执行设备注册
func (s *DeviceRegisterService) Register() error {
	ctx := s.ctx

	// 检查是否启用设备注册
	if !g.Cfg().MustGet(ctx, "mqtt.register.enable", true).Bool() {
		g.Log().Info(ctx, "⏭️ 设备注册功能已禁用")
		return nil
	}

	g.Log().Infof(ctx, "🏷️ 开始设备注册: DeviceId=%s, IP=%s", s.deviceId, s.networkIface.IP)

	// 构建注册消息
	registerMsg, err := s.buildRegisterMessage()
	if err != nil {
		return fmt.Errorf("构建注册消息失败: %v", err)
	}

	// 获取注册主题
	topic, err := s.getRegisterTopic()
	if err != nil {
		return fmt.Errorf("获取注册主题失败: %v", err)
	}

	// 发送注册消息
	return s.publishRegisterMessage(topic, registerMsg)
}

// RegisterWithRetry 带重试的设备注册
func (s *DeviceRegisterService) RegisterWithRetry() {
	ctx := s.ctx

	for s.retryCount <= s.maxRetries {
		if err := s.Register(); err != nil {
			s.retryCount++
			g.Log().Errorf(ctx, "❌ 设备注册失败 (第%d次尝试): %v", s.retryCount, err)

			if s.retryCount <= s.maxRetries {
				g.Log().Infof(ctx, "⏳ %d秒后进行第%d次注册重试", int(s.retryInterval.Seconds()), s.retryCount+1)
				time.Sleep(s.retryInterval)
			} else {
				g.Log().Errorf(ctx, "🚫 设备注册失败，已达到最大重试次数(%d)，停止重试", s.maxRetries)
				return
			}
		} else {
			g.Log().Infof(ctx, "✅ 设备注册成功: DeviceId=%s", s.deviceId)
			s.retryCount = 0 // 重置重试计数
			return
		}
	}
}

// buildRegisterMessage 构建设备注册消息
func (s *DeviceRegisterService) buildRegisterMessage() (string, error) {
	ctx := s.ctx

	// 获取配置信息
	deviceModule := g.Cfg().MustGet(ctx, "device.module", "I-800-RK").String()
	heartBeat := g.Cfg().MustGet(ctx, "mqtt.keepAlive", 60).Int()
	opcuaServerPort := g.Cfg().MustGet(ctx, "device.opcua.serverPort", 4840).Int()

	// 检测 runtime 状态（1231端口监听状态）
	runtimeStatus := s.portChecker.GetRuntimeStatus()
	g.Log().Debugf(ctx, "🔍 Runtime状态检测: 1231端口监听 = %d", runtimeStatus)

	// 构建注册请求
	request := DeviceRegisterRequest{
		CmdId:     guid.S(),
		Version:   "1.0",
		Method:    "event.register",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Data: DeviceRegisterRequestData{
			DeviceModule:    deviceModule,
			DeviceId:        s.deviceId,
			HeartBeat:       heartBeat,
			IP:              s.networkIface.IP,
			RuntimeStatus:   runtimeStatus,
			OpcuaServerPort: opcuaServerPort,
		},
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化注册消息失败: %v", err)
	}

	g.Log().Debugf(ctx, "📋 设备注册消息: %s", string(jsonData))
	return string(jsonData), nil
}

// getRegisterTopic 获取设备注册主题
func (s *DeviceRegisterService) getRegisterTopic() (string, error) {
	ctx := s.ctx

	// 获取主题模板
	topicTemplate := g.Cfg().MustGet(ctx, "mqtt.topics.device.register").String()
	if topicTemplate == "" {
		return "", fmt.Errorf("未配置设备注册主题")
	}

	// 替换占位符
	topic := strings.ReplaceAll(topicTemplate, "{deviceId}", s.deviceId)

	g.Log().Debugf(ctx, "📡 设备注册主题: %s", topic)
	return topic, nil
}

// publishRegisterMessage 发布设备注册消息
func (s *DeviceRegisterService) publishRegisterMessage(topic, message string) error {
	ctx := s.ctx

	// 检查MQTT客户端连接状态
	if s.mqttClient == nil || !s.mqttClient.IsConnected() {
		return fmt.Errorf("MQTT客户端未连接")
	}

	// 发布消息
	token := s.mqttClient.Publish(topic, 1, false, message)

	// 等待发布完成
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("发布注册消息失败: %v", token.Error())
	}

	g.Log().Infof(ctx, "📤 设备注册消息发送成功")
	g.Log().Infof(ctx, "   📡 主题: %s", topic)
	g.Log().Infof(ctx, "   🏷️ 设备ID: %s", s.deviceId)
	g.Log().Infof(ctx, "   🌐 IP地址: %s", s.networkIface.IP)
	g.Log().Infof(ctx, "   💻 网卡: %s (%s)", s.networkIface.Name, s.networkIface.MAC)
	
	// 获取OPC UA端口信息并显示
	opcuaPort := g.Cfg().MustGet(s.ctx, "device.opcua.serverPort", 4840).Int()
	g.Log().Infof(ctx, "   🏭 OPC UA: %s:%d", s.networkIface.IP, opcuaPort)

	return nil
}

// UpdateNetworkInterface 更新网络接口信息
func (s *DeviceRegisterService) UpdateNetworkInterface(networkIface *NetworkInterface) {
	s.networkIface = networkIface
	s.deviceId = networkIface.MAC
	s.retryCount = 0 // 重置重试计数
}

// GetDeviceId 获取设备ID (MAC地址)
func (s *DeviceRegisterService) GetDeviceId() string {
	return s.deviceId
}

// GetNetworkInterface 获取网络接口信息
func (s *DeviceRegisterService) GetNetworkInterface() *NetworkInterface {
	return s.networkIface
}

package service

import (
	"context"
	"demo/internal/model/entity"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

// 定义我们的 MQTT 服务结构体
type sMqtt struct {
	client             mqtt.Client          // Paho MQTT 客户端实例
	messages           []entity.MqttMessage // 内存中存储的消息（实际项目中应该用数据库）
	msgMutex           sync.RWMutex         // 消息操作的读写锁
	subscriptions      map[string]byte      // 跟踪已订阅的主题和QoS（当前连接状态）
	savedSubscriptions map[string]byte      // 保存的订阅记录（用于重连恢复）
	subMutex           sync.RWMutex         // 订阅操作的读写锁
	deviceId           string               // 设备ID（MAC地址），用于重连时恢复订阅

	// 🌐 网络检测与设备注册相关
	networkInterface *NetworkInterface        // 当前使用的网络接口信息
	registerService  *DeviceRegisterService   // 设备注册服务
	netDetectService *NetworkDetectionService // 网络检测服务
	isFirstConnect   bool                     // 标记是否为首次连接
}

var (
	mqttService *sMqtt    // 用于存储单例的服务实例
	mqttOnce    sync.Once // 保证单例只被创建一次
)

// Mqtt 是获取 MQTT 服务单例的函数
func Mqtt() *sMqtt {
	mqttOnce.Do(func() {
		ctx := gctx.New()

		// 🌐 网络接口检测
		g.Log().Info(ctx, "🌐 开始网络接口检测...")
		netDetectService := NewNetworkDetectionService()
		networkInterface, err := netDetectService.DetectAvailableNetwork()
		if err != nil {
			g.Log().Errorf(ctx, "❌ 网络接口检测失败: %v", err)
			g.Log().Warning(ctx, "⚠️ 将使用默认配置继续启动")
			// 使用默认网络配置
			networkInterface = &NetworkInterface{
				Name: "default",
				MAC:  "00-00-00-00-00-00",
				IP:   "127.0.0.1",
			}
		}

		// 📝 更新设备ID为检测到的MAC地址
		deviceId := networkInterface.MAC
		g.Log().Infof(ctx, "🏷️ 设备ID: %s", deviceId)

		// --- MQTT 客户端配置 ---
		// 从配置文件读取MQTT服务器配置
		broker := g.Cfg().MustGet(ctx, "mqtt.broker", "tcp://127.0.0.1:1883").String()

		// 🆔 动态生成ClientID: edge-{MAC地址}
		clientId := fmt.Sprintf("edge-%s", deviceId)
		g.Log().Infof(ctx, "🔗 MQTT客户端ID: %s", clientId)

		keepAlive := g.Cfg().MustGet(ctx, "mqtt.keepAlive", 60).Int()
		pingTimeout := g.Cfg().MustGet(ctx, "mqtt.pingTimeout", 10).Int()
		connectTimeout := g.Cfg().MustGet(ctx, "mqtt.connectTimeout", 30).Int()
		autoReconnect := g.Cfg().MustGet(ctx, "mqtt.autoReconnect", true).Bool()
		maxReconnectInterval := g.Cfg().MustGet(ctx, "mqtt.maxReconnectInterval", 60).Int()
		connectRetryInterval := g.Cfg().MustGet(ctx, "mqtt.connectRetryInterval", 5).Int()
		connectRetry := g.Cfg().MustGet(ctx, "mqtt.connectRetry", true).Bool()
		cleanSession := g.Cfg().MustGet(ctx, "mqtt.cleanSession", false).Bool()

		g.Log().Info(ctx, "🔗 MQTT服务配置", g.Map{
			"broker":               broker,
			"clientId":             clientId,
			"keepAlive":            keepAlive,
			"pingTimeout":          pingTimeout,
			"connectTimeout":       connectTimeout,
			"autoReconnect":        autoReconnect,
			"maxReconnectInterval": maxReconnectInterval,
			"connectRetryInterval": connectRetryInterval,
			"connectRetry":         connectRetry,
			"cleanSession":         cleanSession,
		})

		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker)
		opts.SetClientID(clientId)

		// 🔄 可靠性配置
		opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)                       // 心跳间隔
		opts.SetPingTimeout(time.Duration(pingTimeout) * time.Second)                   // ping超时时间
		opts.SetConnectTimeout(time.Duration(connectTimeout) * time.Second)             // 连接超时时间
		opts.SetAutoReconnect(autoReconnect)                                            // 启用自动重连
		opts.SetMaxReconnectInterval(time.Duration(maxReconnectInterval) * time.Second) // 最大重连间隔
		opts.SetConnectRetryInterval(time.Duration(connectRetryInterval) * time.Second) // 重连间隔
		opts.SetConnectRetry(connectRetry)                                              // 启用连接重试

		// 🔐 会话配置
		opts.SetCleanSession(cleanSession) // 持久会话，重连后恢复订阅

		// 设置一个默认的消息处理回调函数
		opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			g.Log().Infof(gctx.New(), "MQTT Received Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
			// 将接收到的消息存储到内存中
			if mqttService != nil {
				mqttService.storeMessage(msg)
			}
		})

		// 📡 连接状态回调 - 这个回调会在mqttService创建后被设置

		// ❌ 连接丢失回调
		opts.OnConnectionLost = func(client mqtt.Client, err error) {
			g.Log().Error(gctx.New(), "🔴 MQTT连接丢失，准备自动重连", g.Map{
				"error":     err.Error(),
				"broker":    broker,
				"clientId":  clientId,
				"time":      time.Now().Format("2006-01-02 15:04:05"),
				"autoRetry": "启用",
			})

			// 🔧 连接丢失时保存订阅记录并清空当前状态
			if mqttService != nil {
				mqttService.subMutex.Lock()
				// 保存当前订阅记录用于重连恢复
				for topic, qos := range mqttService.subscriptions {
					mqttService.savedSubscriptions[topic] = qos
				}
				subscriptionCount := len(mqttService.subscriptions)
				// 清空当前订阅状态，因为连接已丢失
				mqttService.subscriptions = make(map[string]byte)
				mqttService.subMutex.Unlock()

				g.Log().Info(gctx.New(), "💾 已保存订阅记录用于重连恢复", g.Map{
					"count":  subscriptionCount,
					"reason": "连接丢失，将在重连后恢复订阅",
				})
			}
		}

		// 🔄 重连回调
		opts.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {
			g.Log().Info(gctx.New(), "🟡 MQTT正在尝试重连", g.Map{
				"broker":   broker,
				"clientId": clientId,
				"time":     time.Now().Format("2006-01-02 15:04:05"),
			})
		}

		// 先创建mqttService结构体（不包含client）

		mqttService = &sMqtt{
			messages:           make([]entity.MqttMessage, 0),
			subscriptions:      make(map[string]byte),
			savedSubscriptions: make(map[string]byte),
			deviceId:           deviceId,
			networkInterface:   networkInterface,
			netDetectService:   netDetectService,
			isFirstConnect:     true, // 初始化为true，表示首次连接
		}

		// � 设置正确的连接回调（mqttService初始化后）
		opts.OnConnect = func(client mqtt.Client) {
			isReconnect := !mqttService.isFirstConnect
			if mqttService.isFirstConnect {
				mqttService.isFirstConnect = false // 首次连接后设置为false
			}

			connectionType := "首次连接"
			if isReconnect {
				connectionType = "重连"
			}

			g.Log().Info(gctx.New(), "🟢 MQTT连接成功", g.Map{
				"broker":         broker,
				"clientId":       clientId,
				"deviceId":       deviceId,
				"connectionType": connectionType,
				"time":           time.Now().Format("2006-01-02 15:04:05"),
			})

			// 只有在重连时才恢复订阅
			if isReconnect {
				g.Log().Info(gctx.New(), "🔄 检测到重连，开始恢复订阅...")
				go mqttService.reconnectSubscriptions()
			}

			// 🏷️ 设备注册：连接成功后自动发送设备注册消息
			g.Log().Debug(gctx.New(), "🏷️ 准备执行设备注册...")
			go mqttService.handleDeviceRegistration(client, isReconnect)
		}

		// 创建客户端实例（在回调设置之后）
		client := mqtt.NewClient(opts)
		mqttService.client = client

		// 📁 确保算法文件夹存在
		ensureAlgorithmDir(ctx)

		// �🔄 异步连接MQTT，避免阻塞主程序启动
		go func() {
			g.Log().Info(ctx, "🔄 开始连接MQTT broker...")

			connectToken := client.Connect()

			// 设置连接超时检查
			go func() {
				timeout := time.Duration(connectTimeout) * time.Second
				time.Sleep(timeout)
				if !connectToken.WaitTimeout(100 * time.Millisecond) {
					g.Log().Error(ctx, "⚠️ MQTT连接超时", g.Map{
						"broker":  broker,
						"timeout": fmt.Sprintf("%d秒", connectTimeout),
						"action":  "将继续尝试连接，程序其他服务正常运行",
					})
				}
			}()

			// 等待连接结果
			if connectToken.Wait() && connectToken.Error() != nil {
				g.Log().Error(ctx, "❌ MQTT初始连接失败", g.Map{
					"broker": broker,
					"error":  connectToken.Error().Error(),
					"action": "自动重连机制已启用，将在后台持续尝试连接",
				})
			}
		}()

		// 🏥 启动健康检查协程
		healthCheckEnable := g.Cfg().MustGet(ctx, "mqtt.healthCheck.enable", true).Bool()
		if healthCheckEnable {
			go mqttService.startHealthCheck()
		}
	})
	return mqttService
}

// Publish 方法用于发布消息
func (s *sMqtt) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := s.client.Publish(topic, qos, retained, payload)
	// Wait 会等待消息发送完成
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe 方法用于订阅主题
func (s *sMqtt) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := s.client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// 记录订阅状态，用于重连时恢复
	s.subMutex.Lock()
	s.subscriptions[topic] = qos      // 当前连接状态
	s.savedSubscriptions[topic] = qos // 保存用于重连恢复
	s.subMutex.Unlock()

	g.Log().Infof(gctx.New(), "✅ 订阅主题成功: %s (QoS: %d)", topic, qos)
	return nil
}

// storeMessage 存储接收到的消息到内存中（实际项目中应该存储到数据库）
func (s *sMqtt) storeMessage(msg mqtt.Message) {
	s.msgMutex.Lock()
	defer s.msgMutex.Unlock()

	// TODO: 实体字段问题，暂时跳过存储
	_ = msg
	/*
		message := entity.MqttMessage{
			Topic:     msg.Topic(),
			Payload:   string(msg.Payload()),
			Qos:       int(msg.Qos()),
			Retained:  msg.Retained(),
			CreatedAt: gtime.Now().Unix(),
		}

		s.messages = append(s.messages, message)
	*/

	// 限制内存中的消息数量，只保留最近的1000条
	if len(s.messages) > 1000 {
		s.messages = s.messages[len(s.messages)-1000:]
	}
}

// GetMessages 获取接收到的消息列表
func (s *sMqtt) GetMessages(topic string, limit int) []entity.MqttMessage {
	s.msgMutex.RLock()
	defer s.msgMutex.RUnlock()

	// TODO: 实体字段问题，暂时返回空列表
	_ = topic
	_ = limit
	return []entity.MqttMessage{}

	/*
		var result []entity.MqttMessage
		count := 0

		// 从后往前取，获取最新的消息
		for i := len(s.messages) - 1; i >= 0 && count < limit; i-- {
			if topic == "" || s.messages[i].Topic == topic {
				result = append([]entity.MqttMessage{s.messages[i]}, result...)
				count++
			}
		}

		return result
	*/
}

// GetStatus 获取MQTT连接状态
func (s *sMqtt) GetStatus() map[string]interface{} {
	opts := s.client.OptionsReader()

	s.subMutex.RLock()
	subscriptionCount := len(s.subscriptions)
	subscriptionList := make([]string, 0, len(s.subscriptions))
	for topic := range s.subscriptions {
		subscriptionList = append(subscriptionList, topic)
	}
	s.subMutex.RUnlock()

	return map[string]interface{}{
		"connected":          s.client.IsConnected(),
		"client_id":          opts.ClientID(),
		"servers":            opts.Servers(),
		"keep_alive":         opts.KeepAlive(),
		"auto_reconnect":     opts.AutoReconnect(),
		"clean_session":      opts.CleanSession(),
		"device_id":          s.deviceId,
		"subscription_count": subscriptionCount,
		"subscriptions":      subscriptionList,
	}
}

// ForceReconnect 强制重新连接MQTT
func (s *sMqtt) ForceReconnect() error {
	ctx := gctx.New()

	if s.client.IsConnected() {
		g.Log().Info(ctx, "🔄 断开当前MQTT连接以进行重连")
		s.client.Disconnect(250)
	}

	g.Log().Info(ctx, "🔄 开始强制重连MQTT")

	// 等待一小段时间
	time.Sleep(1 * time.Second)

	token := s.client.Connect()
	if token.Wait() && token.Error() != nil {
		g.Log().Error(ctx, "❌ 强制重连失败", g.Map{
			"error": token.Error().Error(),
		})
		return token.Error()
	}

	g.Log().Info(ctx, "✅ 强制重连成功")
	return nil
}

// GetConnectionQuality 获取连接质量信息
func (s *sMqtt) GetConnectionQuality() map[string]interface{} {
	return map[string]interface{}{
		"is_connected":      s.client.IsConnected(),
		"last_ping_time":    time.Now().Format("2006-01-02 15:04:05"),
		"connection_stable": s.client.IsConnected(),
	}
}

// startHealthCheck 启动MQTT健康检查
func (s *sMqtt) startHealthCheck() {
	ctx := gctx.New()

	// 从配置文件读取健康检查间隔
	healthCheckInterval := g.Cfg().MustGet(ctx, "mqtt.healthCheck.interval", 30).Int()
	ticker := time.NewTicker(time.Duration(healthCheckInterval) * time.Second)
	defer ticker.Stop()

	g.Log().Info(ctx, "🏥 MQTT健康检查服务已启动", g.Map{
		"interval": fmt.Sprintf("%d秒", healthCheckInterval),
	})

	for range ticker.C {
		if !s.client.IsConnected() {
			g.Log().Warning(ctx, "⚠️ MQTT健康检查: 连接已断开", g.Map{
				"time":       time.Now().Format("2006-01-02 15:04:05"),
				"auto_retry": "MQTT客户端会自动尝试重连",
			})
		} else {
			// 连接正常时进行简单的状态记录
			s.subMutex.RLock()
			subCount := len(s.subscriptions)
			s.subMutex.RUnlock()

			g.Log().Debug(ctx, "💚 MQTT健康检查: 连接正常", g.Map{
				"time":          time.Now().Format("2006-01-02 15:04:05"),
				"subscriptions": subCount,
				"device_id":     s.deviceId,
			})
		}
	}
}

// Disconnect 优雅断开MQTT连接
func (s *sMqtt) Disconnect() {
	ctx := gctx.New()

	if s.client.IsConnected() {
		g.Log().Info(ctx, "🔌 正在断开MQTT连接")
		s.client.Disconnect(250) // 等待250毫秒完成当前操作
		g.Log().Info(ctx, "✅ MQTT连接已断开")
	} else {
		g.Log().Info(ctx, "ℹ️ MQTT连接已经断开")
	}
}

// IsConnected 检查MQTT是否连接
func (s *sMqtt) IsConnected() bool {
	return s.client.IsConnected()
}

// reconnectSubscriptions 重连后恢复所有订阅
func (s *sMqtt) reconnectSubscriptions() {
	ctx := gctx.New()

	s.subMutex.RLock()
	subscriptions := make(map[string]byte)
	// 从保存的订阅记录中读取，而不是当前的空记录
	for topic, qos := range s.savedSubscriptions {
		subscriptions[topic] = qos
	}
	s.subMutex.RUnlock()

	if len(subscriptions) == 0 {
		g.Log().Info(ctx, "🔄 重连完成，无需恢复订阅")
		return
	}

	g.Log().Info(ctx, "🔄 开始恢复订阅", g.Map{
		"count": len(subscriptions),
	})

	// 等待一小段时间确保连接稳定
	time.Sleep(2 * time.Second)

	successCount := 0
	for topic, qos := range subscriptions {
		// 重新订阅算法消息
		if strings.Contains(topic, "/request") && s.deviceId != "" {
			g.Log().Info(ctx, "🔄 重新订阅算法消息", g.Map{
				"topic":    topic,
				"deviceId": s.deviceId,
			})

			err := s.StartAlgorithmMessageListener(s.deviceId)
			if err != nil {
				g.Log().Error(ctx, "❌ 恢复算法订阅失败", g.Map{
					"topic": topic,
					"error": err.Error(),
				})
			} else {
				successCount++
				g.Log().Info(ctx, "✅ 恢复算法订阅成功", g.Map{
					"topic": topic,
					"qos":   qos,
				})
			}
		} else {
			// 其他普通订阅恢复
			g.Log().Info(ctx, "🔄 重新订阅普通主题", g.Map{
				"topic": topic,
				"qos":   qos,
			})

			token := s.client.Subscribe(topic, qos, nil)
			if token.Wait() && token.Error() != nil {
				g.Log().Error(ctx, "❌ 恢复订阅失败", g.Map{
					"topic": topic,
					"error": token.Error().Error(),
				})
			} else {
				// 手动添加到订阅记录中
				s.subMutex.Lock()
				s.subscriptions[topic] = qos
				s.subMutex.Unlock()

				successCount++
				g.Log().Info(ctx, "✅ 恢复订阅成功", g.Map{
					"topic": topic,
					"qos":   qos,
				})
			}
		}
	}

	g.Log().Info(ctx, "🔄 订阅恢复完成", g.Map{
		"total":   len(subscriptions),
		"success": successCount,
		"failed":  len(subscriptions) - successCount,
	})
}

// StartAlgorithmMessageListener 启动算法相关消息监听
func (s *sMqtt) StartAlgorithmMessageListener(deviceId string) error {
	ctx := gctx.New()

	// 记录deviceId供重连时使用
	s.deviceId = deviceId

	// 从配置文件读取算法请求主题模板
	requestTopicTemplate := g.Cfg().MustGet(ctx, "mqtt.topics.algorithm.request", "/sys/i800/{deviceId}/request").String()

	// 替换deviceId占位符
	topic := strings.Replace(requestTopicTemplate, "{deviceId}", deviceId, -1)

	// 检查是否已经订阅过该主题，避免重复订阅
	s.subMutex.RLock()
	_, alreadySubscribed := s.subscriptions[topic]
	s.subMutex.RUnlock()

	if alreadySubscribed {
		g.Log().Info(ctx, "⚠️ 主题已订阅，跳过重复订阅", g.Map{
			"topic": topic,
		})
		return nil
	}

	g.Log().Info(ctx, "🎯 启动算法监听服务", g.Map{
		"requestTopic": topic,
		"deviceId":     deviceId,
	})

	return s.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		g.Log().Info(ctx, "📨 收到算法处理请求", g.Map{
			"topic":   msg.Topic(),
			"payload": string(msg.Payload()),
		})

		go s.handleAlgorithmMessage(msg, deviceId)
	})
}

// handleAlgorithmMessage 处理算法相关MQTT消息
func (s *sMqtt) handleAlgorithmMessage(msg mqtt.Message, deviceId string) {
	ctx := gctx.New()

	// 先解析基本结构，确定方法类型
	var baseRequest struct {
		CmdId     string          `json:"cmdId"`
		Version   string          `json:"version"`
		Method    string          `json:"method"`
		Timestamp string          `json:"timestamp"`
		Params    json.RawMessage `json:"params"`
	}

	if err := json.Unmarshal(msg.Payload(), &baseRequest); err != nil {
		g.Log().Error(ctx, "解析MQTT消息失败", g.Map{
			"error":   err,
			"payload": string(msg.Payload()),
		})
		return
	}

	// 根据方法类型处理
	switch baseRequest.Method {
	case "algorithm.add":
		var request AlgorithmAddRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.add消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
			// 发送错误响应给客户端，避免客户端一直等待
			reply := AlgorithmReply{
				CmdId:     baseRequest.CmdId,
				Version:   baseRequest.Version,
				Method:    baseRequest.Method,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Code:      CodeInvalidParams,
				Message:   "JSON解析失败: " + err.Error(),
				Data:      nil,
			}
			s.sendAlgorithmReply(&reply, deviceId)
			return
		}
		s.handleAlgorithmAdd(&request, deviceId)
	case "algorithm.delete":
		var request AlgorithmDeleteRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.delete消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
			// 发送错误响应给客户端
			reply := AlgorithmReply{
				CmdId:     baseRequest.CmdId,
				Version:   baseRequest.Version,
				Method:    baseRequest.Method,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Code:      CodeInvalidParams,
				Message:   "JSON解析失败: " + err.Error(),
				Data:      nil,
			}
			s.sendAlgorithmReply(&reply, deviceId)
			return
		}
		s.handleAlgorithmDelete(&request, deviceId)
	case "algorithm.show":
		var request AlgorithmShowRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.show消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
			// 发送错误响应给客户端
			reply := AlgorithmReply{
				CmdId:     baseRequest.CmdId,
				Version:   baseRequest.Version,
				Method:    baseRequest.Method,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Code:      CodeInvalidParams,
				Message:   "JSON解析失败: " + err.Error(),
				Data:      nil,
			}
			s.sendAlgorithmReply(&reply, deviceId)
			return
		}
		s.handleAlgorithmShow(&request, deviceId)
	case "algorithm.config":
		var request AlgorithmConfigRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.config消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
			// 发送错误响应给客户端
			reply := AlgorithmReply{
				CmdId:     baseRequest.CmdId,
				Version:   baseRequest.Version,
				Method:    baseRequest.Method,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Code:      CodeInvalidParams,
				Message:   "JSON解析失败: " + err.Error(),
				Data:      nil,
			}
			s.sendAlgorithmReply(&reply, deviceId)
			return
		}
		s.handleAlgorithmConfig(&request, deviceId)
	default:
		g.Log().Warning(ctx, "未知的算法操作方法", g.Map{
			"method": baseRequest.Method,
		})
	}
}

// handleAlgorithmAdd 处理算法添加请求
func (s *sMqtt) handleAlgorithmAdd(req *AlgorithmAddRequest, deviceId string) {
	ctx := gctx.New()

	// 创建响应结构
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "success",
	}

	// 参数验证
	if req.Params.AlgorithmId == "" || req.Params.AlgorithmDataUrl == "" || req.Params.Md5 == "" {
		reply.Code = CodeInvalidParams
		reply.Message = "必要参数缺失"
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 创建算法下载服务
	downloadService := NewAlgorithmDownloadService()

	// 下载算法文件
	localPath, err := downloadService.DownloadAlgorithmFile(
		req.Params.AlgorithmId,
		req.Params.AlgorithmVersionId,
		req.Params.AlgorithmDataUrl,
		req.Params.Md5,
	)
	if err != nil {
		g.Log().Error(ctx, "下载算法文件失败", g.Map{
			"error":       err,
			"algorithmId": req.Params.AlgorithmId,
			"url":         req.Params.AlgorithmDataUrl,
		})

		// 根据错误类型设置错误码
		if strings.Contains(err.Error(), "MD5校验失败") {
			reply.Code = CodeMd5CheckFailed
		} else if strings.Contains(err.Error(), "下载") {
			reply.Code = CodeDownloadFailed
		} else {
			reply.Code = CodeFileSystemError
		}
		reply.Message = err.Error()
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 同步到数据库
	err = downloadService.SyncAlgorithmToDatabase(req, localPath)
	if err != nil {
		// 检查是否是版本已存在错误
		if versionExistsErr, ok := err.(*AlgorithmVersionExistsError); ok {
			// 版本已存在，返回特定信息
			reply.Code = CodeVersionExists
			reply.Message = "算法版本已存在，忽略本次下发"
			reply.Data = map[string]interface{}{
				"algorithmId": versionExistsErr.AlgorithmId,
				"version":     versionExistsErr.Version,
				"localPath":   versionExistsErr.LocalPath,
				"status":      "ignored",
			}
			g.Log().Info(ctx, "算法版本已存在，返回忽略响应", g.Map{
				"algorithmId": versionExistsErr.AlgorithmId,
				"version":     versionExistsErr.Version,
			})
		} else {
			// 其他数据库错误
			g.Log().Error(ctx, "同步算法到数据库失败", g.Map{
				"error":       err,
				"algorithmId": req.Params.AlgorithmId,
				"localPath":   localPath,
			})
			reply.Code = CodeDatabaseError
			reply.Message = err.Error()
		}
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 成功响应
	reply.Message = "success"
	reply.Data = map[string]interface{}{
		"localPath":   localPath,
		"algorithmId": req.Params.AlgorithmId,
		"version":     req.Params.AlgorithmVersion,
	}

	g.Log().Info(ctx, "算法添加完成", g.Map{
		"algorithmId": req.Params.AlgorithmId,
		"version":     req.Params.AlgorithmVersion,
		"localPath":   localPath,
	})

	s.sendAlgorithmReply(&reply, deviceId)
}

// handleAlgorithmDelete 处理算法删除请求
func (s *sMqtt) handleAlgorithmDelete(req *AlgorithmDeleteRequest, deviceId string) {
	ctx := gctx.New()

	// 创建响应结构
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "success",
	}

	// 参数验证
	algorithmId := req.Params.AlgorithmId
	if algorithmId == "" {
		reply.Code = CodeInvalidParams
		reply.Message = "算法ID不能为空"
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	g.Log().Info(ctx, "开始处理算法删除请求", g.Map{
		"algorithmId": algorithmId,
		"deviceId":    deviceId,
	})

	// 创建算法删除服务实例
	deleteService := NewAlgorithmDeleteService()

	// 执行算法删除
	err := deleteService.DeleteAlgorithm(algorithmId)
	if err != nil {
		// 特殊处理：算法不存在时按照接口协议返回成功，但message警告
		if strings.Contains(err.Error(), "算法不存在") {
			g.Log().Warning(ctx, "算法不存在，按协议返回成功状态", g.Map{
				"algorithmId": algorithmId,
				"warning":     err.Error(),
			})
			// 算法不存在时，code=0（成功），message警告
			reply.Message = fmt.Sprintf("警告：算法不存在: %s", algorithmId)
			reply.Data = map[string]interface{}{
				"algorithmId": algorithmId,
			}
			s.sendAlgorithmReply(&reply, deviceId)
			return
		}

		// 其他错误正常处理
		g.Log().Error(ctx, "删除算法失败", g.Map{
			"algorithmId": algorithmId,
			"error":       err,
		})

		if strings.Contains(err.Error(), "数据库") {
			reply.Code = CodeDatabaseError
		} else {
			reply.Code = CodeFileSystemError
		}
		reply.Message = err.Error()
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 成功响应
	reply.Message = "success"
	reply.Data = map[string]interface{}{
		"algorithmId": algorithmId,
	}

	g.Log().Info(ctx, "算法删除完成", g.Map{
		"algorithmId": algorithmId,
	})

	s.sendAlgorithmReply(&reply, deviceId)
}

// handleAlgorithmShow 处理算法展示请求
func (s *sMqtt) handleAlgorithmShow(req *AlgorithmShowRequest, deviceId string) {
	ctx := gctx.New()

	// 创建响应结构
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "success",
	}

	// 使用算法查询服务获取算法列表
	showService := NewAlgorithmShowService()
	algorithmList, err := showService.GetAlgorithmList(ctx)
	if err != nil {
		reply.Code = CodeDatabaseError
		reply.Message = "查询算法列表失败"
		g.Log().Error(ctx, "algorithm.show处理失败", g.Map{
			"deviceId": deviceId,
			"error":    err,
		})
	} else {
		reply.Data = algorithmList
		g.Log().Info(ctx, "algorithm.show处理成功", g.Map{
			"deviceId": deviceId,
			"count":    len(algorithmList),
		})
	}

	s.sendAlgorithmReply(&reply, deviceId)
}

// handleAlgorithmConfig 处理算法配置请求
func (s *sMqtt) handleAlgorithmConfig(req *AlgorithmConfigRequest, deviceId string) {
	ctx := gctx.New()

	// 创建响应结构
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "success",
	}

	// 参数验证
	if req.Params.AlgorithmId == "" {
		reply.Code = CodeInvalidParams
		reply.Message = "算法ID不能为空"
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	if req.Params.RunStatus != 0 && req.Params.RunStatus != 1 {
		reply.Code = CodeInvalidParams
		reply.Message = "运行状态值无效，只能是0(关闭)或1(开启)"
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 使用算法配置服务更新运行状态
	configService := NewAlgorithmConfigService()
	err := configService.UpdateAlgorithmRunStatus(ctx, req.Params.AlgorithmId, req.Params.RunStatus)
	if err != nil {
		reply.Code = CodeAlgorithmNotFound
		reply.Message = err.Error()
		g.Log().Error(ctx, "algorithm.config处理失败", g.Map{
			"deviceId":    deviceId,
			"algorithmId": req.Params.AlgorithmId,
			"runStatus":   req.Params.RunStatus,
			"error":       err,
		})
	} else {
		g.Log().Info(ctx, "algorithm.config处理成功", g.Map{
			"deviceId":    deviceId,
			"algorithmId": req.Params.AlgorithmId,
			"runStatus":   req.Params.RunStatus,
		})
	}

	s.sendAlgorithmReply(&reply, deviceId)
}

// sendAlgorithmReply 发送算法处理结果响应
func (s *sMqtt) sendAlgorithmReply(reply *AlgorithmReply, deviceId string) {
	ctx := gctx.New()

	// 从配置文件读取算法响应主题模板
	replyTopicTemplate := g.Cfg().MustGet(ctx, "mqtt.topics.algorithm.reply", "/sys/i800/{deviceId}/reply").String()

	// 替换deviceId占位符
	replyTopic := strings.Replace(replyTopicTemplate, "{deviceId}", deviceId, -1)

	// 序列化响应消息
	replyJson, err := json.Marshal(reply)
	if err != nil {
		g.Log().Error(ctx, "序列化响应消息失败", g.Map{
			"error": err,
			"reply": reply,
		})
		return
	}

	// 发送响应
	err = s.Publish(replyTopic, 0, false, replyJson)
	if err != nil {
		g.Log().Error(ctx, "发送算法响应失败", g.Map{
			"error": err,
			"topic": replyTopic,
			"reply": string(replyJson),
		})
	} else {
		g.Log().Info(ctx, "算法响应发送成功", g.Map{
			"topic":     replyTopic,
			"cmdId":     reply.CmdId,
			"version":   reply.Version,
			"method":    reply.Method,
			"timestamp": reply.Timestamp,
			"code":      reply.Code,
			"message":   reply.Message,
			"reply":     string(replyJson),
		})
	}
}

// ================================== 设备注册相关方法 ==================================

// handleDeviceRegistration 处理设备注册
func (s *sMqtt) handleDeviceRegistration(client mqtt.Client, isReconnect bool) {
	ctx := gctx.New()
	g.Log().Debug(ctx, "🏷️ 进入设备注册处理方法, isReconnect=%t", isReconnect)

	// 检查是否启用设备注册
	enableRegister := g.Cfg().MustGet(ctx, "mqtt.register.enable", true).Bool()
	if !enableRegister {
		g.Log().Info(ctx, "⏭️ 设备注册功能已禁用")
		return
	}

	// 检查连接类型的注册开关
	if isReconnect {
		if !g.Cfg().MustGet(ctx, "mqtt.register.onReconnect", true).Bool() {
			g.Log().Info(ctx, "⏭️ 重连时设备注册已禁用")
			return
		}
	} else {
		if !g.Cfg().MustGet(ctx, "mqtt.register.onConnect", true).Bool() {
			g.Log().Info(ctx, "⏭️ 首次连接设备注册已禁用")
			return
		}
	}

	// 创建或更新设备注册服务
	if s.registerService == nil {
		s.registerService = NewDeviceRegisterService(client, s.networkInterface)
	} else {
		s.registerService.UpdateNetworkInterface(s.networkInterface)
	}

	// 执行设备注册
	registerType := "首次连接"
	if isReconnect {
		registerType = "重连"
	}

	g.Log().Infof(ctx, "🏷️ 开始设备注册 (%s): DeviceId=%s", registerType, s.deviceId)

	// 异步执行注册，避免阻塞
	go func() {
		err := s.registerService.Register()
		if err != nil {
			g.Log().Errorf(ctx, "❌ 设备注册失败 (%s): %v", registerType, err)

			// 启动重试机制
			g.Log().Info(ctx, "🔄 启动设备注册重试机制")
			go s.registerService.RegisterWithRetry()
		} else {
			g.Log().Infof(ctx, "✅ 设备注册成功 (%s): DeviceId=%s", registerType, s.deviceId)
		}
	}()
}

// GetDeviceInfo 获取设备信息
func (s *sMqtt) GetDeviceInfo() (deviceId string, networkInterface *NetworkInterface) {
	return s.deviceId, s.networkInterface
}

// GetDeviceId 获取设备ID（MAC地址）
func (s *sMqtt) GetDeviceId() string {
	return s.deviceId
}

// UpdateNetworkInterface 更新网络接口信息
func (s *sMqtt) UpdateNetworkInterface() error {
	ctx := gctx.New()

	g.Log().Info(ctx, "🔄 重新检测网络接口...")

	// 重新检测网络接口
	newInterface, err := s.netDetectService.DetectAvailableNetwork()
	if err != nil {
		return fmt.Errorf("网络接口重新检测失败: %v", err)
	}

	// 检查是否有变化
	oldDeviceId := s.deviceId
	newDeviceId := newInterface.MAC

	if oldDeviceId != newDeviceId {
		g.Log().Infof(ctx, "⚠️ 检测到网络接口变化: %s -> %s", oldDeviceId, newDeviceId)

		// 更新内部状态
		s.deviceId = newDeviceId
		s.networkInterface = newInterface

		// 更新设备注册服务
		if s.registerService != nil {
			s.registerService.UpdateNetworkInterface(newInterface)
		}

		// 重新发送设备注册
		if s.client != nil && s.client.IsConnected() {
			go s.handleDeviceRegistration(s.client, false)
		}
	} else {
		g.Log().Info(ctx, "✅ 网络接口未发生变化")
	}

	return nil
}

// getAlgorithmDownloadPath 获取算法下载路径
func getAlgorithmDownloadPath() string {
	ctx := gctx.New()

	// 从配置文件读取下载路径，支持跨平台
	downloadPath := g.Cfg().MustGet(ctx, "algorithm.downloadPath").String()

	// 如果配置文件未设置，使用默认路径
	if downloadPath == "" {
		if runtime.GOOS == "windows" {
			// Windows环境：使用当前工作目录下的runtime/algorithm文件夹
			downloadPath = "./runtime/algorithm"
		} else {
			// Linux/Unix环境：使用/usr/runtime/algorithm
			downloadPath = "/usr/runtime/algorithm"
		}
	}

	return downloadPath
}

// ensureAlgorithmDir 确保算法文件夹存在
func ensureAlgorithmDir(ctx context.Context) {
	// 获取算法下载路径
	downloadPath := getAlgorithmDownloadPath()

	// 检查目录是否存在
	if gfile.IsDir(downloadPath) {
		g.Log().Info(ctx, "📁 算法文件夹检查完成", g.Map{
			"path":   downloadPath,
			"status": "exists",
		})
	} else {
		// 目录不存在，创建目录
		if err := gfile.Mkdir(downloadPath); err != nil {
			g.Log().Error(ctx, "❌ 创建算法文件夹失败", g.Map{
				"path":  downloadPath,
				"error": err,
			})
		} else {
			g.Log().Info(ctx, "📁 算法文件夹检查完成", g.Map{
				"path":   downloadPath,
				"status": "created",
				"note":   "文件夹不存在，已自动创建",
			})
		}
	}
}

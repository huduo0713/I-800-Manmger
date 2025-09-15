package service

import (
	"demo/internal/model/entity"
	"encoding/json"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// 定义我们的 MQTT 服务结构体
type sMqtt struct {
	client   mqtt.Client          // Paho MQTT 客户端实例
	messages []entity.MqttMessage // 内存中存储的消息（实际项目中应该用数据库）
	msgMutex sync.RWMutex         // 消息操作的读写锁
}

var (
	mqttService *sMqtt    // 用于存储单例的服务实例
	mqttOnce    sync.Once // 保证单例只被创建一次
)

// Mqtt 是获取 MQTT 服务单例的函数
func Mqtt() *sMqtt {
	mqttOnce.Do(func() {
		ctx := gctx.New()

		// --- MQTT 客户端配置 ---
		// 从配置文件读取MQTT服务器配置
		broker := g.Cfg().MustGet(ctx, "mqtt.broker", "tcp://127.0.0.1:1883").String()
		clientId := g.Cfg().MustGet(ctx, "mqtt.clientId", "goframe-edge-device").String()
		keepAlive := g.Cfg().MustGet(ctx, "mqtt.keepAlive", 60).Int()

		g.Log().Info(ctx, "MQTT服务配置", g.Map{
			"broker":    broker,
			"clientId":  clientId,
			"keepAlive": keepAlive,
		})

		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker)
		opts.SetClientID(clientId)
		opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)
		// 设置一个默认的消息处理回调函数
		opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			g.Log().Infof(gctx.New(), "MQTT Received Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
			// 将接收到的消息存储到内存中
			if mqttService != nil {
				mqttService.storeMessage(msg)
			}
		})
		// 设置连接成功的回调
		opts.OnConnect = func(client mqtt.Client) {
			g.Log().Info(gctx.New(), "MQTT Connected")
		}
		// 设置连接丢失的回调
		opts.OnConnectionLost = func(client mqtt.Client, err error) {
			g.Log().Errorf(gctx.New(), "MQTT Connection Lost: %v", err)
		}

		// 创建客户端实例
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			// 在实际项目中，这里应该处理失败，比如 panic 或重试
			g.Log().Fatalf(gctx.New(), "MQTT Connect Error: %s", token.Error())
		}

		mqttService = &sMqtt{
			client:   client,
			messages: make([]entity.MqttMessage, 0),
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
	g.Log().Infof(gctx.New(), "Subscribed to topic: %s", topic)
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
	return map[string]interface{}{
		"connected": s.client.IsConnected(),
		"client_id": opts.ClientID(),
		"servers":   opts.Servers(),
	}
}

// IsConnected 检查MQTT是否连接
func (s *sMqtt) IsConnected() bool {
	return s.client.IsConnected()
}

// StartAlgorithmMessageListener 启动算法相关消息监听
func (s *sMqtt) StartAlgorithmMessageListener(deviceId string) error {
	ctx := gctx.New()

	// 从配置文件读取算法请求主题模板
	requestTopicTemplate := g.Cfg().MustGet(ctx, "mqtt.topics.algorithm.request", "/sys/i800/{deviceId}/request").String()

	// 替换deviceId占位符
	topic := strings.Replace(requestTopicTemplate, "{deviceId}", deviceId, -1)

	g.Log().Info(ctx, "算法监听服务配置", g.Map{
		"requestTopic": topic,
		"deviceId":     deviceId,
	})

	return s.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		g.Log().Info(ctx, "收到算法处理请求", g.Map{
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
			return
		}
		s.handleAlgorithmDeleteCorrect(&request, deviceId)
	case "algorithm.show":
		var request AlgorithmAddRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.show消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
			return
		}
		s.handleAlgorithmShow(&request, deviceId)
	case "algorithm.config":
		var request AlgorithmAddRequest
		if err := json.Unmarshal(msg.Payload(), &request); err != nil {
			g.Log().Error(ctx, "解析algorithm.config消息失败", g.Map{
				"error":   err,
				"payload": string(msg.Payload()),
			})
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
		req.Params.Md5,
		req.Params.AlgorithmDataUrl,
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
		g.Log().Error(ctx, "同步算法到数据库失败", g.Map{
			"error":       err,
			"algorithmId": req.Params.AlgorithmId,
			"localPath":   localPath,
		})

		reply.Code = CodeDatabaseError
		reply.Message = err.Error()
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 成功响应
	reply.Code = CodeSuccess
	reply.Message = "算法添加成功"
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

// handleAlgorithmDeleteCorrect 处理算法删除请求（使用正确的结构体）
func (s *sMqtt) handleAlgorithmDeleteCorrect(req *AlgorithmDeleteRequest, deviceId string) {
	ctx := gctx.New()

	// 创建响应结构
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
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
		g.Log().Error(ctx, "删除算法失败", g.Map{
			"algorithmId": algorithmId,
			"error":       err,
		})

		// 根据错误类型设置错误码
		if strings.Contains(err.Error(), "算法不存在") {
			reply.Code = CodeAlgorithmNotFound
		} else if strings.Contains(err.Error(), "数据库") {
			reply.Code = CodeDatabaseError
		} else {
			reply.Code = CodeFileSystemError
		}
		reply.Message = err.Error()
		s.sendAlgorithmReply(&reply, deviceId)
		return
	}

	// 成功响应
	reply.Code = CodeSuccess
	reply.Message = "算法删除成功"
	reply.Data = map[string]interface{}{
		"algorithmId": algorithmId,
	}

	g.Log().Info(ctx, "算法删除完成", g.Map{
		"algorithmId": algorithmId,
	})

	s.sendAlgorithmReply(&reply, deviceId)
}

// handleAlgorithmShow 处理算法展示请求 (占位符实现)
func (s *sMqtt) handleAlgorithmShow(req *AlgorithmAddRequest, deviceId string) {
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "algorithm.show 功能待实现",
	}
	s.sendAlgorithmReply(&reply, deviceId)
}

// handleAlgorithmConfig 处理算法配置请求 (占位符实现)
func (s *sMqtt) handleAlgorithmConfig(req *AlgorithmAddRequest, deviceId string) {
	reply := AlgorithmReply{
		CmdId:     req.CmdId,
		Version:   req.Version,
		Method:    req.Method,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Code:      CodeSuccess,
		Message:   "algorithm.config 功能待实现",
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
			"topic": replyTopic,
			"cmdId": reply.CmdId,
			"code":  reply.Code,
		})
	}
}

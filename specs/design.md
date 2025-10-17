# I-800 边缘设备算法管理系统 - 设计文档

> **项目名称**: I-800 Algorithm Management System
> **创建日期**: 2025-09-15
> **当前版本**: v2.5.0
> **最后更新**: 2025-10-09
> **维护人员**: Development Team

## 📋 文档修订历史

| 版本 | 日期 | 修订内容 | 修订人 |
|------|------|---------|--------|
| v2.5.0 | 2025-10-09 | MQTT日志增强和算法目录管理设计 | Team |
| v2.4.0 | 2025-09-26 | OPC UA集成设计 | Team |
| v2.3.0 | 2025-09-25 | 设备注册系统设计 | Team |
| v2.2.0 | 2025-09-24 | MQTT可靠性架构设计 | Team |
| v2.1.0 | 2025-09-23 | 算法管理服务化设计 | Team |
| v2.0.0 | 2025-09-20 | GoFrame架构设计 | Team |

---

## 1. 系统架构设计

### 1.1 总体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      云端管理平台                              │
│         (MQTT Broker + 算法存储 + OPC UA Client)             │
└──────────────────────┬──────────────────────────────────────┘
                       │ MQTT (QoS 0/1)
                       │ Topic: /sys/i800/{deviceId}/*
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                  I-800 边缘设备 (Go应用)                       │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐   │
│  │            HTTP API Layer (GoFrame)                  │   │
│  │  • RESTful API (8000端口)                            │   │
│  │  • Swagger文档                                        │   │
│  │  • 用户管理、算法查询                                  │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓                                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         MQTT Service Layer (Eclipse Paho)            │   │
│  │  • 设备注册服务 (DeviceRegisterService)              │   │
│  │  • 网络检测服务 (NetworkDetectionService)            │   │
│  │  • 算法下发处理 (handleAlgorithmAdd)                 │   │
│  │  • 算法删除处理 (handleAlgorithmDelete)              │   │
│  │  • 算法查询处理 (handleAlgorithmShow)                │   │
│  │  • 算法配置处理 (handleAlgorithmConfig)              │   │
│  │  • 连接可靠性管理 (健康检查、重连、订阅状态)          │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓                                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Business Service Layer                      │   │
│  │  • AlgorithmDownloadService (算法下载)               │   │
│  │  • AlgorithmDeleteService (算法删除)                 │   │
│  │  • AlgorithmShowService (算法查询)                   │   │
│  │  • AlgorithmConfigService (算法配置)                 │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓                                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Data Access Layer (DAO)                  │   │
│  │  • Algorithm DAO (算法数据操作)                       │   │
│  │  • User DAO (用户数据操作)                            │   │
│  └──────────────────────────────────────────────────────┘   │
│                            ↓                                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           SQLite Database + File System               │   │
│  │  • algorithm表 (复合唯一约束)                         │   │
│  │  • /usr/runtime/algorithm/ (Linux)                    │   │
│  │  • ./runtime/algorithm/ (Windows)                     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              OPC UA Server (可选集成)                         │
│         • 设备数据访问 (端口4840)                             │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 分层职责

| 层次 | 职责 | 主要组件 |
|------|------|----------|
| HTTP API层 | 提供RESTful接口 | Controller、Router |
| MQTT服务层 | 处理MQTT消息、设备注册、连接管理 | MQTT Service、Handlers |
| 业务逻辑层 | 实现具体业务逻辑 | Algorithm Services |
| 数据访问层 | 数据库操作封装 | DAO、Model |
| 存储层 | 数据持久化和文件存储 | SQLite、File System |

---

## 2. 核心模块设计

### 2.1 MQTT服务模块

#### 2.1.1 核心结构

```go
type sMqtt struct {
    client             mqtt.Client          // MQTT客户端
    messages           []entity.MqttMessage // 消息缓存
    msgMutex           sync.RWMutex         // 消息锁
    subscriptions      map[string]byte      // 当前订阅状态
    savedSubscriptions map[string]byte      // 保存的订阅（重连用）
    subMutex           sync.RWMutex         // 订阅锁
    deviceId           string               // 设备ID (MAC地址)

    // 网络与注册相关
    networkInterface   *NetworkInterface        // 网络接口信息
    registerService    *DeviceRegisterService   // 注册服务
    netDetectService   *NetworkDetectionService // 网络检测服务
    isFirstConnect     bool                     // 首次连接标记
}
```

#### 2.1.2 连接可靠性设计

**异步启动机制**:
```go
// 主程序启动
func Main() {
    // HTTP服务同步启动
    startHTTPServer()

    // MQTT服务异步启动（不阻塞）
    go startMQTTService()
}

// MQTT服务初始化
func Mqtt() *sMqtt {
    mqttOnce.Do(func() {
        // 配置MQTT客户端
        opts := mqtt.NewClientOptions()
        opts.SetAutoReconnect(true)
        opts.SetConnectRetry(true)

        // 设置回调
        opts.OnConnect = onConnectCallback
        opts.OnConnectionLost = onConnectionLostCallback
        opts.OnReconnecting = onReconnectingCallback

        // 异步连接
        go func() {
            connectToken := client.Connect()
            connectToken.Wait()
        }()
    })
}
```

**双状态订阅管理**:
```
连接建立 → 订阅Topic → subscriptions[topic]=qos
                     → savedSubscriptions[topic]=qos

连接丢失 → 清空subscriptions
        → 保留savedSubscriptions

重连成功 → 从savedSubscriptions恢复订阅
        → 更新subscriptions
```

**健康检查机制**:
```go
func startHealthCheck() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        if !client.IsConnected() {
            log.Warning("连接已断开")
            // 自动重连机制会处理
        } else {
            log.Debug("连接正常", subCount)
        }
    }
}
```

#### 2.1.3 消息处理流程

```
MQTT消息到达
    ↓
解析基本结构 (cmdId, version, method)
    ↓
根据method路由
    ├─ algorithm.add    → handleAlgorithmAdd()
    ├─ algorithm.delete → handleAlgorithmDelete()
    ├─ algorithm.show   → handleAlgorithmShow()
    └─ algorithm.config → handleAlgorithmConfig()
    ↓
业务处理
    ↓
构造Reply (完整字段)
    ├─ cmdId
    ├─ version
    ├─ method
    ├─ timestamp
    ├─ code (初始化为0-成功)
    ├─ message (初始化为"success")
    └─ data
    ↓
发送回复到 /sys/i800/{deviceId}/reply
    ↓
记录详细日志 (包含完整JSON)
```

---

### 2.2 网络检测与设备注册模块

#### 2.2.1 网络检测服务

```go
type NetworkDetectionService struct {
    mode             string   // auto/config/manual
    configInterface  string   // 配置指定的网卡
    manualMAC        string   // 手动指定的MAC
    manualIP         string   // 手动指定的IP
}

// 检测流程
func DetectAvailableNetwork() (*NetworkInterface, error) {
    switch mode {
    case "auto":
        return autoDetect()      // 自动检测最优网卡
    case "config":
        return configDetect()    // 使用配置指定的网卡
    case "manual":
        return manualDetect()    // 使用手动指定的MAC/IP
    }
}

// 自动检测逻辑
func autoDetect() (*NetworkInterface, error) {
    interfaces := net.Interfaces()

    for _, iface := range interfaces {
        // 1. 过滤条件
        if isVirtualInterface(iface.Name) {
            continue  // 跳过虚拟网卡
        }
        if iface.Flags&net.FlagUp == 0 {
            continue  // 跳过未启动的网卡
        }

        // 2. 获取IP地址
        addrs := iface.Addrs()
        ip := extractIP(addrs)

        // 3. 连通性测试
        if pingTest(ip) {
            return &NetworkInterface{
                Name: iface.Name,
                MAC:  iface.HardwareAddr.String(),
                IP:   ip,
            }, nil
        }
    }

    return nil, errors.New("无可用网络接口")
}
```

#### 2.2.2 设备注册服务

```go
type DeviceRegisterService struct {
    mqttClient       mqtt.Client
    networkInterface *NetworkInterface
    deviceModule     string    // 设备型号
    heartBeat        int       // 心跳周期
    opcuaPort        int       // OPC UA端口
}

// 注册消息结构
type DeviceRegisterRequest struct {
    CmdId     string
    Version   string
    Method    string    // "event.register"
    Timestamp string
    Data      struct {
        DeviceModule    string
        DeviceId        string
        HeartBeat       int
        IP              string
        RuntimeStatus   int
        OpcuaServerPort int
    }
}

// 注册流程
func Register() error {
    // 1. 检测Runtime状态
    runtimeStatus := detectRuntimeStatus()  // 检查1231端口

    // 2. 构造注册消息
    request := buildRegisterRequest(runtimeStatus)

    // 3. 发送到Topic
    topic := fmt.Sprintf("/sys/i800/%s/event/register", deviceId)
    return mqttClient.Publish(topic, 0, false, request)
}
```

---

### 2.3 算法管理服务模块

#### 2.3.1 算法下载服务

```go
type AlgorithmDownloadService struct {
    downloadPath string    // 下载根目录
}

// 下载流程
func DownloadAlgorithmFile(algorithmId, versionId, url, md5 string) (string, error) {
    // 1. 创建目标目录
    targetDir := filepath.Join(downloadPath, algorithmId, versionId)
    os.MkdirAll(targetDir, 0755)

    // 2. 下载文件（实时MD5计算）
    targetFile := filepath.Join(targetDir, versionId)
    hash := md5.New()
    resp, _ := http.Get(url)

    file, _ := os.Create(targetFile)
    writer := io.MultiWriter(file, hash)
    io.Copy(writer, resp.Body)

    // 3. MD5校验
    calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
    if calculatedMD5 != md5 {
        os.Remove(targetFile)
        return "", fmt.Errorf("MD5校验失败")
    }

    // 4. 解压ZIP（如果是ZIP格式）
    if isZipFile(targetFile) {
        extractAlgorithmFile(targetFile, targetDir)
        os.Remove(targetFile)  // 删除ZIP文件
    }

    return targetDir, nil
}

// ZIP解压
func extractAlgorithmFile(zipPath, targetDir string) error {
    reader, _ := zip.OpenReader(zipPath)
    defer reader.Close()

    for _, file := range reader.File {
        extractPath := filepath.Join(targetDir, file.Name)

        if file.FileInfo().IsDir() {
            os.MkdirAll(extractPath, file.Mode())
        } else {
            fileReader, _ := file.Open()
            targetFile, _ := os.OpenFile(extractPath,
                os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
            io.Copy(targetFile, fileReader)
        }
    }
    return nil
}

// 数据库同步
func SyncAlgorithmToDatabase(req *Request, localPath string) error {
    // 1. 检查版本是否已存在
    existing, _ := dao.Algorithm.Ctx(ctx).
        Where("algorithm_id = ? AND algorithm_version_id = ?",
            req.AlgorithmId, req.AlgorithmVersionId).
        One()

    if !existing.IsEmpty() {
        return &AlgorithmVersionExistsError{...}
    }

    // 2. 插入新记录
    _, err := dao.Algorithm.Ctx(ctx).Insert(do.Algorithm{
        AlgorithmId:        req.AlgorithmId,
        AlgorithmName:      req.AlgorithmName,
        AlgorithmVersion:   req.AlgorithmVersion,
        AlgorithmVersionId: req.AlgorithmVersionId,
        LocalPath:          localPath,
        Md5:                req.Md5,
        FileSize:           int64(req.FileSize),
    })

    return err
}
```

#### 2.3.2 算法删除服务

```go
type AlgorithmDeleteService struct {
    downloadPath string
}

// 删除流程
func DeleteAlgorithm(algorithmId string) error {
    // 1. 检查算法是否存在
    exists, algorithmInfo, _ := CheckAlgorithmExists(algorithmId)
    if !exists {
        return fmt.Errorf("算法不存在: %s", algorithmId)
    }

    // 2. 删除文件
    algorithmDir := filepath.Join(downloadPath, "algorithm", algorithmId)
    err := os.RemoveAll(algorithmDir)
    if err != nil {
        return fmt.Errorf("删除文件失败: %v", err)
    }

    // 3. 删除数据库记录
    _, err = dao.Algorithm.Ctx(ctx).
        Where("algorithm_id = ?", algorithmId).
        Delete()

    return err
}
```

#### 2.3.3 算法查询服务

```go
type AlgorithmShowService struct{}

// 查询流程
func GetAlgorithmList(ctx context.Context) ([]AlgorithmShowResponseData, error) {
    // 1. 查询数据库
    algorithms, _ := dao.Algorithm.Ctx(ctx).All()

    // 2. 读取运行状态
    responseData := make([]AlgorithmShowResponseData, 0)
    for _, algo := range algorithms {
        runStatus := readAlgorithmRunStatus(
            algo.AlgorithmId,
            algo.AlgorithmVersionId,
        )

        responseData = append(responseData, AlgorithmShowResponseData{
            AlgorithmName:    algo.AlgorithmName,
            AlgorithmId:      algo.AlgorithmId,
            AlgorithmVersion: algo.AlgorithmVersion,
            RunStatus:        runStatus,
        })
    }

    return responseData, nil
}

// 读取运行状态
func readAlgorithmRunStatus(algorithmId, versionId string) int {
    configPath := filepath.Join(
        downloadPath, algorithmId, versionId, "config.yaml",
    )

    var config AlgorithmConfig
    yamlFile, _ := ioutil.ReadFile(configPath)
    yaml.Unmarshal(yamlFile, &config)

    return config.Algo.RunStatus  // 0或1
}
```

#### 2.3.4 算法配置服务

```go
type AlgorithmConfigService struct{}

// 更新运行状态
func UpdateAlgorithmRunStatus(ctx context.Context, algorithmId string, runStatus int) error {
    // 1. 查询算法信息
    algo, _ := dao.Algorithm.Ctx(ctx).
        Where("algorithm_id = ?", algorithmId).
        One()

    if algo.IsEmpty() {
        return fmt.Errorf("算法不存在: %s", algorithmId)
    }

    // 2. 读取config.yaml
    configPath := filepath.Join(
        downloadPath, algorithmId, algo.AlgorithmVersionId, "config.yaml",
    )

    var config AlgorithmConfig
    yamlFile, _ := ioutil.ReadFile(configPath)
    yaml.Unmarshal(yamlFile, &config)

    // 3. 更新runStatus
    config.Algo.RunStatus = runStatus

    // 4. 写回文件
    newYaml, _ := yaml.Marshal(&config)
    return ioutil.WriteFile(configPath, newYaml, 0644)
}
```

---

### 2.4 算法目录管理设计

#### 2.4.1 目录结构

```
Linux生产环境:
/usr/runtime/algorithm/
└── {algorithmId}/
    └── {algorithmVersionId}/
        ├── config.yaml         # 算法配置
        ├── algorithm.so        # 算法库文件
        └── resources/          # 资源文件

Windows开发环境:
./runtime/algorithm/
└── {algorithmId}/
    └── {algorithmVersionId}/
        ├── config.yaml
        ├── algorithm.dll
        └── resources/
```

#### 2.4.2 自动创建机制

```go
// MQTT服务启动时调用
func ensureAlgorithmDir(ctx context.Context) {
    // 1. 获取算法下载路径
    downloadPath := getAlgorithmDownloadPath()

    // 2. 检查目录是否存在
    if gfile.IsDir(downloadPath) {
        g.Log().Info(ctx, "📁 算法文件夹检查完成", g.Map{
            "path":   downloadPath,
            "status": "exists",
        })
    } else {
        // 3. 创建目录（包含父目录）
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

// 跨平台路径获取
func getAlgorithmDownloadPath() string {
    // 1. 优先使用配置文件
    downloadPath := g.Cfg().MustGet(ctx, "algorithm.downloadPath").String()

    // 2. 使用默认路径
    if downloadPath == "" {
        if runtime.GOOS == "windows" {
            downloadPath = "./runtime/algorithm"
        } else {
            downloadPath = "/usr/runtime/algorithm"
        }
    }

    return downloadPath
}
```

---

### 2.5 日志增强设计

#### 2.5.1 MQTT回复日志

```go
func sendAlgorithmReply(reply *AlgorithmReply, deviceId string) {
    // 序列化
    replyJson, _ := json.Marshal(reply)

    // 发送
    err := s.Publish(replyTopic, 0, false, replyJson)

    if err != nil {
        // 失败日志
        g.Log().Error(ctx, "发送算法响应失败", g.Map{
            "error": err,
            "topic": replyTopic,
            "reply": string(replyJson),
        })
    } else {
        // 成功日志（完整字段）
        g.Log().Info(ctx, "算法响应发送成功", g.Map{
            "topic":     replyTopic,
            "cmdId":     reply.CmdId,
            "version":   reply.Version,
            "method":    reply.Method,
            "timestamp": reply.Timestamp,
            "code":      reply.Code,
            "message":   reply.Message,
            "reply":     string(replyJson),  // 完整JSON
        })
    }
}
```

#### 2.5.2 日志级别设计

```yaml
logger:
  level: "info"     # debug/info/warning/error

  # debug: 详细调试信息（开发环境）
  # info:  一般信息日志（生产环境）
  # warning: 警告信息
  # error: 错误信息
```

---

## 3. 数据库设计

### 3.1 algorithm表

```sql
CREATE TABLE algorithm (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    algorithm_id        TEXT NOT NULL,
    algorithm_name      TEXT NOT NULL,
    algorithm_version   TEXT NOT NULL,
    algorithm_version_id TEXT NOT NULL,
    algorithm_data_url  TEXT NOT NULL,
    file_size           INTEGER NOT NULL,
    md5                 TEXT NOT NULL,
    local_path          TEXT NOT NULL,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- 复合唯一约束
    UNIQUE(algorithm_id, algorithm_version_id)
);

-- 更新时间触发器
CREATE TRIGGER update_algorithm_timestamp
AFTER UPDATE ON algorithm
FOR EACH ROW
BEGIN
    UPDATE algorithm SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;

-- 索引优化
CREATE INDEX idx_algorithm_id ON algorithm(algorithm_id);
CREATE INDEX idx_algorithm_version ON algorithm(algorithm_version_id);
```

### 3.2 user表（预留）

```sql
CREATE TABLE user (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    email      TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 4. 配置设计

### 4.1 配置文件结构

```yaml
# manifest/config/config.yaml

# 服务器配置
server:
  address: ":8000"
  swaggerPath: "/swagger"

# 日志配置
logger:
  level: "info"
  stdout: true
  path: "logs"
  file: "app.log"
  rotateSize: "20MB"
  rotateBackupLimit: 10
  rotateBackupExpire: "30d"
  rotateBackupCompress: 9

# 数据库配置
database:
  default:
    type: "sqlite"
    path: "./data/sqlite.db"
    debug: false

# MQTT配置
mqtt:
  broker: "tcp://127.0.0.1:1883"
  keepAlive: 60
  pingTimeout: 10
  connectTimeout: 30
  autoReconnect: true
  maxReconnectInterval: 60
  connectRetryInterval: 5
  connectRetry: true
  cleanSession: false

  # 主题配置
  topics:
    algorithm:
      request: "/sys/i800/{deviceId}/request"
      reply: "/sys/i800/{deviceId}/reply"

  # 健康检查
  healthCheck:
    enable: true
    interval: 30

  # 设备注册配置
  register:
    enable: true
    onConnect: true
    onReconnect: true
    retryInterval: 30
    maxRetries: 10

# 设备配置
device:
  module: "I-800-RK"
  heartBeat: 10

  # 网络检测配置
  network:
    mode: "auto"              # auto/config/manual
    interface: ""             # 指定网卡名称(config模式)
    manualMAC: ""             # 手动MAC(manual模式)
    manualIP: ""              # 手动IP(manual模式)
    pingTarget: "8.8.8.8"     # Ping测试目标
    pingTimeout: 3

  # Runtime配置
  runtime:
    port: 1231
    checkTimeout: 5

  # OPC UA配置
  opcua:
    serverPort: 4840

# 算法配置
algorithm:
  downloadPath: ""            # 留空使用默认路径
  maxConcurrentDownloads: 3
  maxFileSize: 524288000      # 500MB
```

---

## 5. 错误处理设计

### 5.1 错误码定义

```go
const (
    CodeSuccess           = 0
    CodeDownloadFailed    = 1001
    CodeMd5CheckFailed    = 1002
    CodeFileSystemError   = 1003
    CodeDatabaseError     = 1004
    CodeInvalidParams     = 1005
    CodeAlgorithmNotFound = 1006
    CodeVersionExists     = 1007
)
```

### 5.2 错误响应格式

```json
{
  "cmdId": "uuid-1234",
  "version": "1.0",
  "method": "algorithm.add",
  "timestamp": "2025-10-09 10:00:00",
  "code": 1002,
  "message": "MD5校验失败: 期望5d4ee33d，实际a1b2c3d4",
  "data": null
}
```

### 5.3 特殊错误处理

**算法删除不存在的情况**:
```go
// 按照接口协议，算法不存在时返回成功
if strings.Contains(err.Error(), "算法不存在") {
    reply.Code = CodeSuccess  // code=0
    reply.Message = fmt.Sprintf("警告：算法不存在: %s", algorithmId)
    reply.Data = map[string]interface{}{
        "algorithmId": algorithmId,
    }
    return
}
```

---

## 6. 部署设计

### 6.1 SysV Init脚本

```bash
#!/bin/sh
# /etc/init.d/S99app_service

APP_NAME="app_mng"
APP_PATH="/root/I-800-manmger"
PID_FILE="/var/run/$APP_NAME.pid"
LOG_FILE="/var/log/$APP_NAME.log"

start() {
    if [ -f "$PID_FILE" ]; then
        echo "$APP_NAME is already running"
        return 1
    fi

    cd "$APP_PATH" || exit 1
    nohup ./"$APP_NAME" >> "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "$APP_NAME started"
}

stop() {
    if [ ! -f "$PID_FILE" ]; then
        echo "$APP_NAME is not running"
        return 1
    fi

    PID=$(cat "$PID_FILE")
    kill "$PID"
    rm -f "$PID_FILE"
    echo "$APP_NAME stopped"
}

case "$1" in
    start)   start   ;;
    stop)    stop    ;;
    restart) stop; sleep 1; start ;;
    status)  status  ;;
    *)       echo "Usage: $0 {start|stop|restart|status}" ;;
esac
```

### 6.2 Docker部署（可选）

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app_mng .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app_mng .
COPY manifest/config/config.yaml manifest/config/
EXPOSE 8000
CMD ["./app_mng"]
```

---

## 7. 性能优化设计

### 7.1 并发下载控制

```go
type DownloadManager struct {
    semaphore chan struct{}  // 限制并发数
}

func (dm *DownloadManager) Download(url string) error {
    dm.semaphore <- struct{}{}  // 获取信号量
    defer func() { <-dm.semaphore }()  // 释放信号量

    // 执行下载
    return downloadFile(url)
}

// 初始化
manager := &DownloadManager{
    semaphore: make(chan struct{}, 3),  // 最多3个并发
}
```

### 7.2 数据库连接池

```yaml
database:
  default:
    maxIdleConnNum: 10    # 最大空闲连接数
    maxOpenConnNum: 20    # 最大打开连接数
    maxConnLifetime: "30s"  # 连接最大生命周期
```

---

## 8. 安全设计

### 8.1 MQTT认证（可选）

```yaml
mqtt:
  broker: "tcp://127.0.0.1:1883"
  username: "edge-device"
  password: "secure-password"
```

### 8.2 文件权限

```go
// 创建文件时设置权限
os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)

// 创建目录时设置权限
os.MkdirAll(dir, 0755)
```

---

## 9. 监控设计

### 9.1 关键指标

| 指标 | 描述 | 阈值 |
|------|------|------|
| mqtt_connected | MQTT连接状态 | 1=连接，0=断开 |
| algorithm_download_success_rate | 算法下载成功率 | >95% |
| md5_check_success_rate | MD5校验通过率 | 100% |
| api_response_time_p95 | API响应时间P95 | <200ms |
| disk_usage | 磁盘使用率 | <80% |

### 9.2 日志监控

```bash
# 监控关键日志
tail -f /var/log/app_mng.log | grep -E "ERROR|WARN|MQTT连接"

# 统计错误日志
grep "ERROR" /var/log/app_mng.log | wc -l
```

---

**文档结束**

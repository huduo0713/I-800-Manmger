# I-800 边缘设备算法管理系统

> 基于 **GoFrame v2.9.3** 构建的边缘设备算法管理系统，支持 MQTT 协议的云端算法下发、文件管理、自动解压和数据同步功能。

## ⭐ 核心功能

- 🔗 **MQTT算法下发**: 支持云端通过MQTT协议进行算法下发、启停等管理
- � **MQTT可靠性增强**: 异步连接启动、双状态订阅管理、自动重连、健康检查系统
- �📁 **智能文件管理**: 自动下载、MD5校验、版本管理、ZIP自动解压
- 🗃️ **数据同步**: SQLite数据库存储算法信息，复合唯一约束
- 🔄 **跨平台支持**: Windows开发、Linux部署自适应
- 📊 **完整API**: HTTP RESTful API (用户管理、算法查询)
- 🎯 **版本控制**: 基于算法ID+版本的智能去重机制
- 📝 **滚动日志**: 20MB文件滚动、自动压缩、持久化存储
- 🛡️ **复合校验**: MD5完整性 + 版本唯一性 + 数据库约束 + 文件结构校验

## � 重要版本历史

### 🎯 v2.3.0 - 智能设备注册系统 (2025-09-25)
- 🆕 **智能网卡检测**: 自动检测可用网卡，过滤虚拟网卡，连通性验证
- 🆕 **多模式支持**: auto/config/manual三种检测模式，适配不同环境
- 🆕 **动态设备ID**: 基于MAC地址生成唯一设备标识，避免配置冲突
- 🆕 **自动注册**: MQTT连接成功后自动发送设备注册消息
- 🆕 **ClientID优化**: edge-{MAC地址} 格式，符合设备命名规范
- 🛠️ **连接状态修复**: 正确区分首次连接和重连，修复连接类型显示问题
- ⚙️ **配置增强**: 新增device.network和mqtt.register配置节点

### 🛡️ v2.2.0 - MQTT可靠性增强 (2025-09-24)
- 🆕 **异步MQTT启动**: 解决连接阻塞主程序启动问题
- 🆕 **双状态订阅管理**: 彻底解决重连后"假订阅"现象
- 🆕 **智能健康检查**: 30秒间隔状态检测，自动重连机制
- 🆕 **生产级可靠性**: 连接失败自动重试，网络异常自动恢复
- 🛠️ **订阅状态追踪**: savedSubscriptions机制确保重连后正确恢复订阅
- 📊 **详细日志记录**: 连接、订阅、重连全过程监控

### 🎯 v2.1.0 - 企业级算法管理 (2025-09-23)
- 🆕 **复合校验机制**: MD5完整性 + 版本唯一性 + 数据库约束 + 文件结构校验
- 🆕 **智能去重**: 基于算法ID+版本的复合唯一约束，防止重复下载
- 🆕 **ZIP自动解压**: 支持ZIP文件自动解压和目录结构创建
- 🆕 **跨平台路径**: Windows开发/Linux部署路径自适应
- 🛠️ **增强数据库**: SQLite触发器、约束优化，数据完整性保障
- 📝 **滚动日志**: 20MB文件滚动、自动压缩、持久化存储

### 🚀 v2.0.0 - GoFrame架构重构 (2025-09-20)
- 🔄 **架构升级**: 从原生Go迁移到GoFrame v2.9.3企业级框架
- 🆕 **标准化结构**: 采用GoFrame标准项目结构，提升可维护性
- 🆕 **HTTP API**: 完整的RESTful API接口，支持用户管理和算法查询
- 🆕 **配置管理**: 基于GoFrame的配置系统，支持多环境配置
- 🆕 **ORM集成**: 使用GoFrame ORM，简化数据库操作
- 📊 **Swagger文档**: 自动生成API文档，提升开发效率

### 🎉 v1.0.0 - 基础MQTT算法管理 (2025-09-15)
- 🆕 **MQTT协议支持**: 基于Eclipse Paho实现云端算法下发
- 🆕 **文件下载管理**: HTTP文件下载、MD5校验、版本管理
- 🆕 **SQLite数据库**: 算法信息持久化存储
- 🆕 **算法生命周期**: 下发、启动、停止、删除完整流程
- 🛠️ **基础日志**: 文件日志记录，支持调试和监控
- 🔧 **配置文件**: YAML配置文件，支持MQTT和数据库参数配置

## �🚀 快速开始

```bash
# 快速启动项目
go run .

# 编译和运行项目
bash script/run.sh

# 同步目标文件到目标设备I-800
bash script/rsync.sh

# 目标设备I-800上部署自启动服务脚本并启动
bash script/deploy.sh install

# 访问API文档
http://127.0.0.1:8000/swagger/
```

## 🏗️ 系统架构

### 技术栈
- **框架**: GoFrame v2.9.3
- **数据库**: SQLite3 (嵌入式) - 增强约束和触发器
- **通信协议**: MQTT (Eclipse Paho v1.5.0) + HTTP - 生产级可靠性增强
- **MQTT可靠性**: 异步连接、双状态订阅管理、自动重连、健康检查
- **文件处理**: 自动ZIP解压、MD5实时校验
- **日志系统**: 滚动文件存储、自动压缩清理
- **部署**: Docker + Kubernetes

### 分层架构
```
MQTT协议层 → Service业务层 → DAO数据层 → SQLite数据库
HTTP API层 → Controller控制层 → Service业务层 ↗
```

### 🔄 复合校验机制

```
📥 MQTT算法下发请求
    ↓
🔍 版本存在性检查（基于algorithmId + algorithmVersion）
    ↓
📥 HTTP文件下载（流式处理 + 实时MD5计算）
    ↓
🔐 MD5完整性验证（与预期值对比）
    ↓
📦 ZIP文件自动解压（提取动态库和配置文件）
    ↓
💾 数据库记录存储（复合唯一约束保护）
    ↓
🏠 版本化目录结构（algorithm/{algorithmId}/{algorithmVersionId}/）
    ↓
📝 操作日志记录（滚动文件持久化）
    ↓
🎉 算法部署就绪
```

## 📁 项目结构详解

### 💼 核心业务层
```
internal/
├── service/              ⭐ 业务逻辑层
│   ├── mqtt.go              # MQTT通信服务
│   ├── algorithm_common.go  # 算法通用数据结构和错误码
│   ├── algorithm_add.go     # 算法添加/下发核心逻辑（新增ZIP解压、版本控制）
│   └── algorithm_delete.go  # 算法删除核心逻辑
├── controller/          📡 HTTP API控制器
│   ├── algorithm/       # 算法管理API
│   └── user/           # 用户管理API
├── dao/                💾 数据访问层 (自动生成)
├── model/              📋 数据模型层
│   ├── entity/         # 数据库表实体（增强约束）
│   └── do/             # 数据操作对象
└── cmd/                🚀 启动命令层
```

### 🔧 配置和部署
```
manifest/
├── config/config.yaml   # 主配置文件（新增滚动日志配置）
├── deploy/             # K8s部署配置
└── docker/             # Docker容器化

data/
└── init.sql            # 数据库初始化脚本（增强约束和触发器）

logs/                   # 📝 新增：滚动日志目录
├── app-YYYYMMDD.log    # 当前日志文件（20MB滚动）
└── app-*.log.gz        # 历史日志文件（自动压缩）

runtime/                # 🗂️ 算法文件存储目录
└── algorithm/          # 新的目录结构
    └── {algorithmId}/  # 算法ID目录
        └── {versionId}/ # 版本ID目录
            ├── algorithm.zip     # 原始压缩包
            ├── algorithm.so/.dll # 解压后的动态库
            └── config.json       # 解压后的配置文件

hack/
├── config.yaml         # GoFrame CLI配置
└── *.mk               # 开发脚本
```

### 📊 业务服务架构
```
internal/service/
├── algorithm_common.go     # 🔧 公共组件
│   ├── AlgorithmAddRequest
│   ├── AlgorithmDeleteRequest  
│   ├── AlgorithmReply
│   └── 错误码常量 (CodeSuccess, CodeDownloadFailed等)
│
├── algorithm_add.go        # ⬇️ 算法添加服务（重构增强）
│   ├── AlgorithmDownloadService
│   ├── DownloadAlgorithmFile() - 下载+实时MD5校验
│   ├── extractAlgorithmFile() - 🆕 ZIP自动解压功能
│   ├── SyncAlgorithmToDatabase() - 版本去重逻辑
│   └── cleanupDirectoryOnFailure() - 故障安全清理
│
├── algorithm_delete.go     # 🗑️ 算法删除服务
│   ├── AlgorithmDeleteService
│   ├── CheckAlgorithmExists()
│   ├── DeleteAlgorithmFiles()
│   ├── DeleteAlgorithmFromDatabase()
│   └── DeleteAlgorithm()
│
└── mqtt.go                # 📡 MQTT通信服务
    ├── handleAlgorithmAdd() → NewAlgorithmDownloadService()
    └── handleAlgorithmDeleteCorrect() → NewAlgorithmDeleteService()
```

🎯 **架构优势**
- 🧩 **单一职责** - 每个文件专注特定功能
- 🔄 **易于扩展** - 添加新算法操作只需新增文件
- 🛠️ **便于维护** - 修改特定功能不影响其他模块
- 📦 **代码复用** - 公共组件统一管理，避免重复
- 🧪 **便于测试** - 每个服务可以独立进行单元测试
- 🛡️ **故障隔离** - 完善的错误处理和自动清理机制
```
manifest/
├── config/config.yaml   # 主配置文件
├── deploy/             # K8s部署配置
└── docker/             # Docker容器化

data/
└── init.sql            # 数据库初始化脚本

hack/
├── config.yaml         # GoFrame CLI配置
└── *.mk               # 开发脚本
```

## ⚙️ 配置说明

### 主配置文件 (`manifest/config/config.yaml`)

```yaml
# 边缘设备配置
device:
  id: "edge-device-001"
  
# MQTT服务配置 - 生产级可靠性增强
mqtt:
  broker: "tcp://127.0.0.1:1883"  # 你的MQTT服务器地址
  clientId: "goframe-edge-device"
  
  # 连接可靠性参数
  keepAlive: 60              # 心跳间隔(秒) - 检测连接状态
  pingTimeout: 10            # Ping超时(秒) - 心跳响应超时判定  
  connectTimeout: 30         # 连接超时(秒) - 初始连接超时
  
  # 自动重连参数  
  autoReconnect: true        # 启用自动重连
  maxReconnectInterval: 60   # 最大重连间隔(秒)
  connectRetryInterval: 5    # 重连尝试间隔(秒)
  
  # 会话管理
  cleanSession: false        # 保持会话状态，支持离线消息
  
  # 健康检查
  healthCheck:
    enable: true            # 启用健康检查
    interval: 30            # 检查间隔(秒)

# 算法下载配置  
algorithm:
  downloadPath: ""  # 留空自动适配平台路径
  # Windows: ./runtime/algorithm/
  # Linux: /usr/runtime/algorithm/

# 📝 滚动日志配置（新增）
logger:
  path: "logs"                     # 日志文件存储目录
  file: "app-{Ymd}.log"           # 日志文件命名格式（按天）
  rotateSize: "20MB"              # 单个文件大小限制
  rotateBackupLimit: 5            # 保留历史文件数量
  rotateBackupExpire: "30d"       # 历史文件过期时间
  rotateBackupCompress: true      # 压缩历史文件
```

### 跨平台路径自适应

- **开发环境(Windows)**: `./runtime/algorithm/{algorithmId}/{algorithmVersionId}/`
- **生产环境(Linux)**: `/usr/runtime/algorithm/{algorithmId}/{algorithmVersionId}/`
- **自定义路径**: 在配置文件中指定 `algorithm.downloadPath`

### 📦 算法文件组织结构（新版）

```
runtime/
└── algorithm/                    # 算法根目录
    └── {algorithmId}/           # 按算法ID分组
        └── {algorithmVersionId}/ # 按版本ID分组
            ├── algorithm.zip     # 🔒 原始压缩包（MD5已校验）
            ├── algorithm.so      # 🚀 自动解压的动态库
            ├── algorithm.dll     # 🚀 Windows动态库
            ├── config.json       # ⚙️ 算法配置文件
            └── *.so / *.dll      # 📚 其他解压文件
```

**关键改进**：
- ✅ **版本隔离**：不同版本独立目录，避免冲突
- ✅ **自动解压**：ZIP文件下载后自动提取内容
- ✅ **路径安全**：防止ZIP炸弹和路径遍历攻击
- ✅ **完整保留**：原始ZIP包和解压文件并存

## � MQTT可靠性增强系统 【v2.2.0 - 生产级稳定性升级】

### 🎯 核心问题解决

在实际设备部署过程中，MQTT连接的稳定性直接影响算法下发的成功率。我们在生产环境中发现了两个关键问题并进行了系统性的解决：

#### ⚠️ 问题1: MQTT连接阻塞启动
**现象**: 程序启动时卡在 `🔗 MQTT服务配置`，长时间无响应
**根因**: MQTT连接采用同步阻塞方式，当网络环境不稳定时会导致整个启动流程卡死
**影响**: 设备无法正常启动，影响生产部署

#### ⚠️ 问题2: "假订阅"现象  
**现象**: MQTT重连后显示订阅成功，但实际收不到消息
**根因**: 连接丢失后 `alreadySubscribed` 标志位未正确重置，导致重连时跳过订阅操作
**影响**: 设备看似正常但无法接收云端指令，业务功能完全失效

### 🛡️ 解决方案架构

#### 1️⃣ **异步MQTT连接启动**
```go
// 修改前 - 阻塞启动
func startMQTTAlgorithmService() {
    service.Mqtt().StartAlgorithmMessageListener() // 阻塞等待连接
}

// 修改后 - 异步非阻塞
func startMQTTAlgorithmService() {
    go func() {
        time.Sleep(2 * time.Second) // 给其他服务启动时间
        service.Mqtt().StartAlgorithmMessageListener()
    }()
}
```

**优势**:
- ✅ 启动流程不再阻塞，确保HTTP服务等其他功能正常启动
- ✅ MQTT连接失败不影响设备基本功能
- ✅ 支持后台重试连接，提高系统可用性

#### 2️⃣ **双状态订阅管理系统**
```go
type sMqtt struct {
    // 当前激活订阅状态 (连接丢失时清空)
    subscriptions map[string]bool
    
    // 保存的订阅意图 (连接丢失时保留，用于恢复)
    savedSubscriptions map[string]bool
    
    // 其他字段...
}
```

**关键机制**:
- **连接丢失时**: `subscriptions` 清空，`savedSubscriptions` 保留
- **重连成功后**: 从 `savedSubscriptions` 恢复所有订阅
- **订阅操作时**: 同时更新两个状态表

#### 3️⃣ **MQTT参数优化配置**
```yaml
mqtt:
  broker: "tcp://your-mqtt-server:1883"
  clientId: "edge-device-{deviceId}"
  
  # 连接可靠性参数
  keepAlive: 60              # 心跳间隔(秒) - 检测连接状态
  pingTimeout: 10            # Ping超时(秒) - 心跳响应超时判定  
  connectTimeout: 30         # 连接超时(秒) - 初始连接超时
  
  # 自动重连参数  
  autoReconnect: true        # 启用自动重连
  maxReconnectInterval: 60   # 最大重连间隔(秒)
  connectRetryInterval: 5    # 重连尝试间隔(秒)
  
  # 会话管理
  cleanSession: false        # 保持会话状态，支持离线消息
  
  # 健康检查
  healthCheck:
    enable: true            # 启用健康检查
    interval: 30            # 检查间隔(秒)
```

**参数说明**:
- `keepAlive`: 心跳间隔，用于检测连接是否存活
- `pingTimeout`: 心跳响应超时时间，超时认为连接断开
- `connectTimeout`: 初始连接建立的超时时间
- `autoReconnect`: 自动重连开关，连接断开后自动尝试重连
- `maxReconnectInterval`: 重连最大间隔，避免频繁重连消耗资源
- `connectRetryInterval`: 重连尝试间隔，平衡及时性与资源消耗
- `cleanSession`: 会话清理标志，false时服务器保留订阅和离线消息

#### 4️⃣ **健康检查与诊断系统**
```go
// 定期健康检查
func (s *sMqtt) startHealthCheck() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        for range ticker.C {
            s.checkConnectionHealth()
        }
    }()
}

func (s *sMqtt) checkConnectionHealth() {
    if s.client == nil || !s.client.IsConnected() {
        g.Log().Warning(ctx, "MQTT连接异常，尝试重连")
        s.reconnect()
    }
}
```

### 📊 完整架构流程

```
设备启动
    ↓
异步启动MQTT服务 (非阻塞)
    ↓
HTTP服务正常启动 ✅
    ↓
后台MQTT连接尝试
    ↓
┌─────────────────────────────────────┐
│            连接状态监控              │
│  ┌─────────┐    ┌──────────────┐   │
│  │健康检查  │────│连接丢失检测   │   │
│  │30s间隔  │    │心跳超时判定   │   │
│  └─────────┘    └──────────────┘   │
└─────────────────────────────────────┘
    ↓ (连接丢失)
保存当前订阅状态到 savedSubscriptions
清空 subscriptions (标记为未订阅状态)
    ↓
自动重连机制启动
    ↓ (重连成功)
从 savedSubscriptions 恢复所有订阅
更新 subscriptions 状态
    ↓
✅ 消息正常收发恢复
```

### 🔧 关键代码实现

#### OnConnectionLost 回调处理
```go
func (s *sMqtt) OnConnectionLost(client mqtt.Client, err error) {
    g.Log().Warningf(ctx, "MQTT连接丢失: %v", err)
    
    // 保存当前订阅状态用于重连后恢复
    s.mu.Lock()
    s.savedSubscriptions = make(map[string]bool)
    for topic, subscribed := range s.subscriptions {
        s.savedSubscriptions[topic] = subscribed
    }
    // 清空当前订阅状态
    s.subscriptions = make(map[string]bool)
    s.mu.Unlock()
}
```

#### 重连订阅恢复
```go
func (s *sMqtt) reconnectSubscriptions() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // 从保存的订阅状态中恢复
    for topic, _ := range s.savedSubscriptions {
        if token := s.client.Subscribe(topic, 1, s.handleMessage); token.Wait() && token.Error() != nil {
            g.Log().Errorf(ctx, "重连后订阅失败: topic=%s, error=%v", topic, token.Error())
        } else {
            s.subscriptions[topic] = true
            g.Log().Infof(ctx, "重连后订阅成功: %s", topic)
        }
    }
}
```

### 🧪 测试验证方法

#### 1. 启动阻塞问题测试
```bash
# 模拟MQTT服务器不可达
# 修改配置文件中的broker地址为不存在的服务器
broker: "tcp://192.168.1.999:1883"

# 启动程序
go run .

# 验证: 应该能看到HTTP服务正常启动，而不是卡在MQTT连接
```

#### 2. 假订阅问题测试  
```bash
# 1. 正常启动程序并连接MQTT
go run .

# 2. 使用MQTTx发送测试消息，验证能正常接收
# 主题: /sys/i800/{deviceId}/request

# 3. 强制断开MQTT Broker连接 (kill broker进程或断网)

# 4. 观察日志，等待自动重连成功

# 5. 再次发送测试消息
# 验证: 应该能正常接收到消息，不会出现"假订阅"现象
```

### 📈 性能与稳定性提升

#### 性能指标
- **启动时间**: 从不确定(可能无限阻塞) → 稳定2-3秒
- **故障恢复**: 从手动重启 → 自动重连(5秒间隔)
- **消息可靠性**: 从连接丢失后失效 → 自动恢复订阅

#### 稳定性改进
- ✅ **启动解耦**: MQTT连接失败不影响其他服务
- ✅ **状态同步**: 订阅状态与实际连接状态保持一致
- ✅ **故障隔离**: 单个订阅失败不影响其他订阅
- ✅ **监控诊断**: 详细的连接状态日志和健康检查

### 🛠️ 故障排查指南

#### 常见问题诊断

**问题**: MQTT重连后无法接收消息
```bash
# 检查日志中的关键信息
grep "订阅成功\|重连后订阅\|savedSubscriptions" logs/app-*.log

# 预期看到:
# [INFO] 重连后订阅成功: /sys/i800/{deviceId}/request
# [DEBUG] savedSubscriptions恢复: map[/sys/i800/{deviceId}/request:true]
```

**问题**: MQTT连接频繁断开重连
```yaml
# 调整连接参数
mqtt:
  keepAlive: 120          # 增加心跳间隔
  pingTimeout: 20         # 增加ping超时
  maxReconnectInterval: 120 # 增加最大重连间隔
```

**问题**: 程序启动卡住
```bash
# 检查是否使用了异步MQTT启动
grep "异步启动MQTT" logs/app-*.log

# 如果没有看到相关日志，检查cmd.go中的启动代码
```

## �🔄 MQTT算法下发执行流程详解

### 📋 整体架构流程

```
程序启动 
   ↓
cmd/cmd.go 调用 startMQTTAlgorithmService() (异步非阻塞)
   ↓
获取 service.Mqtt() 单例实例 (mqtt.go)
   ↓
调用 StartAlgorithmMessageListener() 订阅配置化主题
   ↓ (并行进行健康检查与自动重连)
接收MQTT消息 → handleAlgorithmMessage() 
   ↓
解析JSON → 根据method调用对应处理器
   ↓
handleAlgorithmAdd() → NewAlgorithmDownloadService()
   ↓                            ↓
算法下载服务 (algorithm_add.go) 执行完整处理流程
   ↓
sendAlgorithmReply() 发送MQTT响应到配置化回复主题
```

### 🎯 核心处理步骤

#### 1. **MQTT消息接收与解析**
```go
// 位置: internal/service/mqtt.go
// 功能: 监听配置化主题，接收算法下发请求
订阅主题: /sys/i800/{deviceId}/request (可配置)
消息格式: JSON格式的AlgorithmAddRequest结构体
```

#### 2. **参数验证**
```go
// 验证必要参数
- algorithmId: 算法唯一标识
- algorithmDataUrl: 文件下载URL  
- md5: 文件完整性校验值
// 缺少任何参数返回错误码 1005
```

#### 3. **文件下载与校验** ⭐
```go
// 位置: internal/service/algorithm_add.go
// 功能: DownloadAlgorithmFile() - 重构增强版本

步骤1: 创建版本化目录 {downloadPath}/algorithm/{algorithmId}/{algorithmVersionId}/
步骤2: HTTP GET 下载文件到本地
步骤3: 实时计算MD5哈希值（边下载边计算）
步骤4: 校验MD5是否匹配
步骤5: 🆕 自动解压ZIP文件到同级目录
步骤6: 故障时自动清理不完整文件和目录

解压功能详解:
- 支持标准ZIP格式算法包
- 自动提取动态库文件 (.so/.dll)
- 自动提取配置文件 (.json/.yaml/.xml等)
- 路径安全检查，防止ZIP炸弹攻击
- 解压失败时完整回滚清理

错误处理:
- 下载失败 → 错误码 1001
- MD5校验失败 → 错误码 1002  
- 文件系统错误 → 错误码 1003
- 🆕 解压失败 → 错误码 1007
```

#### 4. **数据库同步** 🗃️
```go
// 位置: internal/service/algorithm_add.go  
// 功能: SyncAlgorithmToDatabase() - 增强版本管理

🆕 版本去重机制:
- 基于 (algorithmId + algorithmVersion) 复合唯一约束
- 相同版本存在时直接忽略，避免重复下载
- 支持同一算法的多个版本并存
- 自动记录版本变更历史

数据库增强约束:
- MD5字段：强制32位十六进制格式 CHECK(length(md5) = 32)
- 复合唯一：UNIQUE(algorithm_id, algorithm_version)
- 自动时间戳：created_at, updated_at 触发器维护
- 存储路径：记录解压后的完整目录路径

处理逻辑:
- 🆕 新版本: INSERT新记录到数据库
- 🔄 已存在版本: 直接返回成功，记录跳过日志
- 📝 操作日志: 记录所有版本检查和数据库操作

错误处理:
- 数据库操作失败 → 错误码 1004
- 版本查询失败 → 记录错误日志但不中断流程
```

#### 5. **MQTT响应发送**
```go
// 发送到配置化回复主题
回复主题: /sys/i800/{deviceId}/reply (可配置)

成功响应 (code: 0):
{
  "cmdId": "原请求ID",
  "code": 0,
  "message": "算法添加成功",
  "data": {
    "localPath": "本地存储路径",
    "algorithmId": "算法ID",
    "version": "算法版本"
  }
}

失败响应 (code: 非0):
{
  "cmdId": "原请求ID", 
  "code": 1001-1005,
  "message": "具体错误信息"
}
```

### 🔧 服务间协作机制

```
mqtt.go (MQTT通信层)
    ↓ 调用
algorithm_add.go (业务逻辑层)
    ↓ 使用
dao层 (数据访问层)
    ↓ 操作
SQLite数据库
```

**关键设计原则**:
- ✅ **职责分离**: MQTT通信与业务逻辑分离
- ✅ **错误隔离**: 每个步骤独立错误处理
- ✅ **事务安全**: 先文件后数据库，避免数据不一致
- ✅ **配置化**: 主题、路径、服务器等均可配置

### 📊 配置化主题系统

```yaml
# manifest/config/config.yaml
mqtt:
  topics:
    algorithm:
      request: "/sys/i800/{deviceId}/request"   # 算法下发请求主题
      reply: "/sys/i800/{deviceId}/reply"       # 算法下发响应主题
      # 支持扩展其他业务主题
```

**动态替换机制**:
- `{deviceId}` 占位符在运行时自动替换
- 支持灵活的主题规划和企业标准对接

## 🧪 MQTT测试指南

### 1. 启动项目
```bash
go run .
```
看到如下日志表示启动成功：
```
[INFO] MQTT算法处理服务启动成功 {"deviceId":"edge-device-001","topic":"/sys/i800/edge-device-001/request"}
[INFO] swagger ui is serving at address: http://127.0.0.1:8000/swagger/
```

### 2. 使用MQTTx客户端测试

**连接配置**:
- **Host**: `127.0.0.1` (或你的MQTT服务器)
- **Port**: `1883`
- **Client ID**: 任意唯一ID

**发布算法下发请求**:
- **主题**: `/sys/i800/edge-device-001/request`
- **消息**（更新字段）:
```json
{
  "cmdId": "test-001",
  "version": "1.0",
  "method": "algorithm.add",
  "timestamp": "2025-09-18 17:30:00",
  "params": {
    "algorithmId": "algo-test-001",
    "algorithmName": "测试算法",
    "algorithmVersion": "v1.0.0",
    "algorithmVersionId": "version-001",
    "algorithmDataUrl": "https://httpbin.org/base64/UEsDBBQDAAAAAL11XlEAAAAAAAAAAAAAAAAJAAAAYWxnb3JpdGhtL1BLAwQUAwAAAAC9dV5RAAAAAAAAAAAAAAAAABEAAABhbGdvcml0aG0vY29uZmlnLmpzb25QSwMEFAMAAAAAvXVeUQAAAAAAAAAAAAAAAA8AAABhbGdvcml0aG0vbGliLnNvUEsBAj8AFAMAAAAAvXVeUQAAAAAAAAAAAAAAAAkAJAAAAAAAAAAQAAAAAAAAAGFsZ29yaXRobS8KACAAAAAAAQAYAK6zKzrcyDgBrrMrOtzIOAGusys63Mg4AQ==",
    "fileSize": 245,
    "lastModifyTime": "2025-09-18 17:30:00",
    "md5": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d"
  }
}
```
> 💡 **示例说明**: 上述URL是一个包含算法目录结构的ZIP文件的base64编码，解压后包含 `algorithm/config.json` 和 `algorithm/lib.so` 文件

**订阅响应结果**:
- **主题**: `/sys/i800/edge-device-001/reply`
- **成功响应示例**（更新路径结构）:
```json
{
  "cmdId": "test-001",
  "version": "1.0",
  "method": "algorithm.add", 
  "timestamp": "2025-09-18 17:30:01",
  "code": 0,
  "message": "算法添加成功",
  "data": {
    "algorithmId": "algo-test-001",
    "version": "v1.0.0",
    "localPath": "./runtime/algorithm/algo-test-001/version-001/"
  }
}
```

### 3. 验证测试结果

✅ **成功标志**:
1. **版本化目录**: 项目根目录下 `runtime/algorithm/algo-test-001/version-001/`
2. **文件完整性**: 
   - 原始ZIP文件: `algorithm.zip`
   - 自动解压文件: `config.json`, `lib.so` 等
3. **数据库记录**: SQLite数据库中 `algorithm` 表新增记录，包含版本信息
4. **MQTT回复**: 收到 `code: 0` 的成功响应
5. **📝 日志记录**: `logs/` 目录下生成详细的操作日志

🔍 **版本去重测试**:
- 发送相同 `algorithmId` + `algorithmVersion` 的请求
- 系统应返回成功但跳过下载（检查日志确认）

## 🔍 错误码详细说明

| 错误码 | 错误类型 | 触发条件 | 排查方法 | 对应处理阶段 |
|-------|---------|---------|---------|-------------|
| 0 | 成功 | 算法下发完整流程成功 | - | ✅ 全流程完成 |
| 1001 | 下载失败 | HTTP下载超时/网络异常/URL无效 | 检查网络连接、URL可访问性、服务器响应状态 | ❌ 文件下载阶段 |
| 1002 | MD5校验失败 | 下载文件MD5与请求不符 | 确认源文件完整性、重新计算MD5值、检查传输过程 | ❌ 文件校验阶段 |
| 1003 | 文件系统错误 | 磁盘空间不足/目录权限不够/路径创建失败 | 检查磁盘空间、目录读写权限、路径格式 | ❌ 文件系统操作 |
| 1004 | 数据库错误 | SQLite连接失败/SQL执行异常/表结构问题 | 检查数据库文件权限、表结构、SQL语法 | ❌ 数据库同步阶段 |
| 1005 | 参数错误 | algorithmId/algorithmDataUrl/md5缺失/runStatus无效 | 检查MQTT消息格式和必要参数完整性 | ❌ 参数验证阶段 |
| 1006 | 算法不存在 | 删除/配置不存在的算法ID | 检查算法ID是否正确、是否已被删除 | ❌ 算法删除/配置阶段 |
| 🆕 1007 | 版本已存在/ZIP解压失败 | 算法版本重复下发/压缩文件格式错误/解压权限不足 | **版本冲突**: 忽略重复下发。**解压错误**: 检查ZIP文件完整性、解压目录权限、磁盘空间 | ❌ 版本检查/文件解压阶段 |

### 🔄 错误处理流程（增强版）

```
MQTT请求接收
    ↓
参数验证 (错误码1005)
    ↓
🆕 版本存在性检查 (基于algorithmId + algorithmVersion)
    ↓ (版本已存在) → 直接返回成功，记录跳过日志
创建版本化目录结构 (algorithm/{algorithmId}/{algorithmVersionId}/)
    ↓
文件下载 (错误码1001) ❌→ 清理空目录
    ↓  
MD5校验 (错误码1002) ❌→ 删除文件 + 清理空目录
    ↓
🆕 ZIP自动解压 (错误码1007) ❌→ 删除文件 + 清理空目录
    ↓
文件系统操作 (错误码1003) ❌→ 清理空目录
    ↓
数据库同步 (错误码1004)
    ↓
📝 滚动日志记录 (所有操作持久化记录)
    ↓
响应发送 (错误码0-成功)

## 🔄 版本管理重要修复 【v2.1.0关键改进】

### 🐛 关键Bug修复

#### 1. JSON兼容性问题修复 🔧
**问题**: 客户端发送整数fileSize时，Go结构体int64字段解析失败导致程序卡死
```go
// ❌ 修复前：兼容性问题
type AlgorithmAddRequest struct {
    FileSize int64 `json:"fileSize"`    // 无法解析客户端的浮点数JSON
}

// ✅ 修复后：全面兼容
type AlgorithmAddRequest struct {
    FileSize float64 `json:"fileSize"`  // 兼容整数和浮点数JSON
}

// 安全转换使用
func processRequest(req *AlgorithmAddRequest) {
    fileSize := int64(req.FileSize)  // 安全转换为int64使用
}
```

#### 2. 版本冲突处理优化 ⚠️
**问题**: 算法版本已存在时返回"添加成功"误导用户
```go
// 新增专用错误类型
type AlgorithmVersionExistsError struct {
    AlgorithmId string
    Version     string
}

func (e AlgorithmVersionExistsError) Error() string {
    return fmt.Sprintf("算法版本已存在: %s-%s", e.AlgorithmId, e.Version)
}
```

**改进效果**:
- ✅ 版本重复时明确返回错误码 `1007`
- ✅ 错误信息包含具体的algorithmId和version
- ✅ 客户端可根据错误码判断是否为版本冲突

#### 3. 版本更新文件删除Bug修复 🚨
**问题**: 下发新版本时，删除旧版本逻辑错误地将刚下载的新文件也删除
```go
// ❌ 修复前：严重Bug
if oldVersionId != "" {
    gfile.Remove(algorithmDir)  // 错误地删除了新文件
}

// ✅ 修复后：精确删除
if oldVersionId != "" && oldVersionId != req.AlgorithmVersionId {
    oldDir := gfile.Join(baseAlgorithmDir, req.AlgorithmId, oldVersionId)
    if gfile.Exists(oldDir) {
        gfile.Remove(oldDir)  // 只删除真正的旧版本目录
        g.Log().Infof(ctx, "已删除旧版本目录: %s", oldDir)
    }
}
```

### 📊 优化后的版本管理流程

```
算法下发请求 → algorithm.add
    ↓
🔍 数据库版本存在性检查
    ↓
❓ (algorithmId + algorithmVersion) 是否已存在？
    ├── ✅ 存在 → 返回错误码1007 + 明确错误信息 ❌
    └── ❌ 不存在 ↓
📥 HTTP下载文件 (fileSize支持float64格式)
    ↓
✅ MD5文件完整性校验
    ↓
📂 ZIP解压到新版本目录
    ↓
🗑️ 智能清理旧版本 (仅在新版本成功部署后执行)
    ↓
💾 数据库记录更新
    ↓
✅ 返回成功响应
```

### 🧪 修复验证测试

**测试1: JSON兼容性**
```json
{
  "params": {
    "fileSize": 1048576    // 整数格式 ✅ 现在可正常解析
  }
}
```

**测试2: 版本冲突处理**
```json
// 重复版本请求
{
  "params": {
    "algorithmId": "algo-001",
    "algorithmVersion": "v1.0.0"  // 已存在版本
  }
}

// 期望响应
{
  "code": 1007,
  "message": "算法版本已存在: algo-001-v1.0.0",
  "data": null
}
```

**测试3: 版本更新文件保护**
- 下发 `algo-001 v2.0.0` (新版本)
- 验证 `v1.0.0` 目录被删除 ✅
- 验证 `v2.0.0` 目录保留完整 ✅

简化：
📥 MQTT算法下发请求
    ↓
🔍 版本存在性检查（数据库查询）
    ↓
📥 HTTP文件下载（流式处理）
    ↓
🔐 实时MD5计算（边下载边校验）
    ↓
✅ MD5值对比验证
    ↓
📦 ZIP文件解压校验
    ↓
💾 数据库记录存储（复合约束）
    ↓
🏠 文件系统组织（版本化目录）
    ↓
🎉 校验完成，算法就绪

🧹 智能清理机制增强:
- cleanupDirectoryOnFailure() 支持版本化目录清理
- 避免失败后留下 runtime/algorithm/{algorithmId}/{versionId}/ 空文件夹  
- 只删除确认为空的目录，保护现有算法版本
- 🔧 **Windows文件句柄优化**: 确保HTTP响应体和ZIP文件正确关闭
- 🔄 **重试机制**: 带延迟重试的目录删除，提高清理成功率
- ⏱️ **延迟清理**: 等待文件句柄释放后再执行清理操作
- 🛡️ **路径安全**: ZIP解压时防止路径遍历攻击
```

### 🎯 **版本管理机制**（重构升级）

**新版本去重逻辑**:
```
算法下发请求 → 版本存在性检查 (algorithmId + algorithmVersion)
    ↓
如果版本已存在:
  ✅ 直接返回成功响应
  📝 记录日志: "算法版本已存在，忽略下发"
  🚫 跳过下载、解压、数据库操作
  💡 避免重复处理，节省网络和存储资源

如果版本不存在:
  📥 执行完整下发流程
  📦 下载 → MD5校验 → ZIP解压 → 数据库记录
  🗂️ 新版本与现有版本并存，互不干扰
  
支持场景:
- ✅ 同一算法的多个版本并存 (v1.0, v1.1, v2.0)
- ✅ 版本升级时保留历史版本
- ✅ 重复下发请求的幂等性处理
- ✅ 网络异常重试时的去重保护
```
  │
  ├── 清理旧算法目录 (/usr/runtime/algorithmId/)
  ├── 下载新算法文件 
  ├── 更新数据库记录 (version、MD5、path等)
  └── 记录覆盖操作日志
  
支持场景:
- 版本升级: v1.0 → v2.0
- 相同版本覆盖: v1.0 → v1.0 (代码修正)
- 文件内容变更: 相同版本但MD5不同
```

**数据库更新策略**:
- 🆕 **新算法**: `INSERT` 新记录
- 🔄 **存在算法**: `UPDATE` 所有字段 (版本、MD5、路径、文件大小等)
- 🧹 **旧文件清理**: 删除整个algorithmId目录，确保无残留

**重要说明**: 
- ⚠️ **如果文件下载成功但数据库写入失败**: 文件已存储到本地，但数据库无记录
- ✅ **错误恢复**: 重新发送相同请求，系统会检查现有文件并尝试重新同步数据库
- 🧹 **智能清理**: 下载失败时自动清理创建的空目录，避免磁盘垃圾
  - MD5校验失败 → 删除文件 + 清理MD5目录 + 清理空算法目录
  - 网络下载失败 → 清理完整目录结构
  - 文件创建失败 → 清理空目录

## 🌟 架构特点

### 1️⃣ **MQTT边缘通信架构**
- ✅ **模块化设计** - 算法管理分为 `algorithm_add.go` (添加) / `algorithm_delete.go` (删除) / `algorithm_common.go` (公共)
- ✅ **内部通信** - MQTT协议，不暴露HTTP接口
- ✅ **支持方法** - algorithm.add (下发) / algorithm.delete (删除) / algorithm.show (查询) / algorithm.config (配置)
- 🔄 **易扩展** - 后期可添加Controller层提供查询API

### 2️⃣ **跨平台文件管理**
- 🖥️ **Windows开发**: 使用 `./runtime/` 相对路径
- 🐧 **Linux部署**: 自动切换到 `/usr/runtime/` 系统路径
- ⚙️ **配置化**: 支持自定义下载路径

### 3️⃣ **完整的业务流程**
```
MQTT请求 → 参数验证 → 文件下载 → MD5校验 → 数据库同步 → MQTT响应
```

**流程特点**:
- 🔐 **安全可靠**: 每步都有错误处理和回滚机制
- ⚡ **高性能**: 异步处理，不阻塞MQTT主线程  
- 📊 **可观测**: 详细日志记录每个处理步骤
- 🎯 **配置化**: 主题、路径、服务器地址均可配置
- 🔄 **幂等性**: 相同请求多次执行结果一致

### 4️⃣ **MQTT单例架构设计**
```go
// 设计模式: 单例 + 工厂
service.Mqtt() → 全局MQTT客户端实例
    ↓
StartAlgorithmMessageListener() → 订阅配置化主题
    ↓  
handleAlgorithmMessage() → 消息路由分发
    ↓
NewAlgorithmDownloadService() → 创建专用下载服务
```

**优势**:
- 💾 **资源节约**: 单个MQTT连接服务多个业务
- 🔀 **解耦设计**: MQTT通信与业务逻辑分离
- 📈 **易于扩展**: 新增业务只需添加消息处理器

## 🐧 Linux部署指南

### 1. 构建Linux版本
```bash
# 交叉编译
GOOS=linux GOARCH=amd64 go build -o i800-manager-linux .
```

### 2. 部署到Linux服务器
```bash
# 创建部署目录
sudo mkdir -p /opt/i800-manager
sudo mkdir -p /usr/runtime
sudo chmod 755 /usr/runtime

# 上传程序和配置
scp i800-manager-linux user@server:/opt/i800-manager/
scp -r manifest/ user@server:/opt/i800-manager/
scp -r data/ user@server:/opt/i800-manager/

# 启动服务
ssh user@server "cd /opt/i800-manager && ./i800-manager-linux"
```

### 3. 配置修改
```yaml
# manifest/config/config.yaml
algorithm:
  downloadPath: "/usr/runtime"  # Linux路径
mqtt:
  broker: "tcp://your-production-mqtt-server:1883"
```

## 📖 开发指南

### 算法下发功能扩展

如需添加其他MQTT方法（如algorithm.delete），在 `internal/service/mqtt.go` 中扩展：

```go
case "algorithm.delete":
    s.handleAlgorithmDelete(&request, deviceId)
```

### 添加HTTP查询API

如需对外提供算法查询接口：

1. **创建API结构** (`api/algorithm/v1/`)
2. **实现Controller** (`internal/controller/algorithm/`)  
3. **调用Service层** 复用现有业务逻辑

### 数据库操作

使用GoFrame自动生成的DAO层：
```go
// 查询算法
algorithms, err := dao.Algorithm.Ctx(ctx).All()

// 新增算法 
_, err := dao.Algorithm.Ctx(ctx).Data(algorithmData).Insert()
```

## 🔧 故障排查

### 常见问题
1. **MQTT连接失败** - 检查配置文件中的broker地址
2. **路径权限错误** - Linux环境确保/usr/runtime目录权限
3. **下载超时** - 检查网络连接和防火墙设置
4. **MD5不匹配** - 确认文件传输完整性

### 日志监控
启动时关注这些日志：
```
[INFO] 算法下载服务初始化 {"downloadPath":"./runtime","platform":"windows"}
[INFO] MQTT服务配置 {"broker":"tcp://127.0.0.1:1883","clientId":"goframe-edge-device"}
[INFO] MQTT算法处理服务启动成功
```

---

## 🎯 测试流程总结

**是的，你的理解完全正确！** 

✅ **测试流程**：
1. **启动项目**: `go run .`
2. **MQTTx发送**: 向 `/sys/i800/edge-device-001/request` 发送algorithm.add消息
3. **验证成功**: 项目根目录出现 `runtime/` 文件夹，包含下载的算法文件
4. **检查响应**: MQTTx收到 `code: 0` 的成功回复

这就表示MQTT算法下发功能测试成功了！🎉

## 🗑️ 算法删除功能 (algorithm.delete)

### 📋 删除流程架构

```
MQTT消息接收 → algorithm.delete
    ↓
参数验证 (algorithmId不能为空)
    ↓
数据库存在性检查 (CheckAlgorithmExists)
    ↓
删除算法文件目录 (/usr/runtime/{algorithmId}/)
    ↓
删除数据库记录 (algorithm表)
    ↓
发送MQTT响应
```

### 🧪 删除功能测试

**发送algorithm.delete请求**:
- **主题**: `/sys/i800/edge-device-001/request`
- **消息**:
```json
{
  "cmdId": "delete-test-001",
  "version": "1.0",
  "method": "algorithm.delete",
  "timestamp": "2025-09-15 14:30:00",
  "params": {
    "algorithmId": "algo-test-001"
  }
}
```

**期望响应**:
```json
{
  "cmdId": "delete-test-001",
  "version": "1.0", 
  "method": "algorithm.delete",
  "timestamp": "2025-09-15 14:30:01",
  "code": 0,
  "message": "算法删除成功",
  "data": {
    "algorithmId": "algo-test-001"
  }
}
```

### ✅ 删除验证步骤

1. **文件目录验证**: 确认 `runtime/algo-test-001/` 目录已完全删除
2. **数据库验证**: algorithm表中对应记录已删除
3. **MQTT响应验证**: 收到 `code: 0` 成功响应
4. **幂等性测试**: 重复删除同一算法ID应返回 `code: 1006` (算法不存在)

### ⚠️ 删除错误场景

| 场景 | 错误码 | 错误信息 | 处理方法 |
|------|--------|----------|----------|
| algorithmId为空 | 1005 | 算法ID不能为空 | 检查请求参数 |
| 算法不存在 | 1006 | 算法不存在: {algorithmId} | 确认算法ID正确性 |
| 文件删除失败 | 1003 | 删除算法目录失败 | 检查文件权限 |
| 数据库操作失败 | 1004 | 删除算法数据库记录失败 | 检查数据库连接 |

## 🔍 算法查询功能 (algorithm.show) 【新增 v2.1.0】

### 📋 查询流程架构

```
MQTT消息接收 → algorithm.show
    ↓
数据库查询所有算法记录 (DAO层)
    ↓
遍历算法记录，读取每个config.yaml
    ↓
解析 algo.runStatus 运行状态 (1-运行, 0-停止)
    ↓
构建算法列表响应数据
    ↓
发送MQTT响应 (包含运行状态)
```

### 🎯 核心特性

- **🔍 状态实时读取**: 从算法目录下的 `config.yaml` 读取实时运行状态
- **📊 完整信息**: 返回算法名称、ID、版本、运行状态
- **⚡ 零参数查询**: 无需传入参数，自动查询所有算法
- **🛡️ 异常容错**: 配置文件不存在时默认返回停止状态

### 🧪 查询功能测试

**发送algorithm.show请求**:
- **主题**: `/sys/i800/edge-device-001/request`
- **消息**:
```json
{
  "cmdId": "show-test-001",
  "version": "1.0",
  "method": "algorithm.show",
  "timestamp": "2025-09-18 15:30:00",
  "params": null
}
```

**期望响应**:
```json
{
  "cmdId": "show-test-001",
  "version": "1.0", 
  "method": "algorithm.show",
  "timestamp": "2025-09-18 15:30:01",
  "code": 0,
  "message": "success",
  "data": [
    {
      "algorithmName": "夜间节能策略算法",
      "algorithmId": "algo-power-save-001",
      "algorithmVersion": "v1.2.0",
      "runStatus": 1
    },
    {
      "algorithmName": "智能调度算法", 
      "algorithmId": "algo-schedule-002",
      "algorithmVersion": "v2.1.0",
      "runStatus": 0
    }
  ]
}
```

### 📁 配置文件结构

算法运行状态从以下位置读取：
```yaml
# 路径: /usr/runtime/algorithm/{algorithmId}/{algorithmVersionId}/config.yaml
algo:
  runStatus: 1    # 1表示开启，0表示关闭
  name: "算法名称"
  version: "1.0.0"
  # 其他配置...
```

### ⚠️ 查询错误场景

| 场景 | 错误码 | 错误信息 | 处理方法 |
|------|--------|----------|----------|
| 数据库查询失败 | 1004 | 查询算法列表失败 | 检查数据库连接和表结构 |
| 配置文件格式错误 | - | 记录警告，默认runStatus=0 | 检查YAML格式 |
| 配置文件不存在 | - | 记录警告，默认runStatus=0 | 算法部署不完整 |

## ⚙️ 算法启停功能 (algorithm.config) 【新增 v2.1.0】

### 📋 启停流程架构

```
MQTT消息接收 → algorithm.config  
    ↓
参数验证 (algorithmId + runStatus)
    ↓
数据库存在性检查 (验证算法是否存在)
    ↓
读取原config.yaml配置文件
    ↓
更新 algo.runStatus 为新值
    ↓
写入更新后的config.yaml
    ↓
发送MQTT响应
```

### 🎯 核心特性

- **🎛️ 远程控制**: 通过MQTT消息控制算法启停
- **✅ 存在性校验**: 启停前验证算法是否在数据库中存在
- **📁 配置持久化**: 直接修改算法目录下的 `config.yaml` 文件
- **🔄 状态同步**: 设备根据配置文件实时调整算法运行状态
- **🛡️ 参数验证**: 严格校验 runStatus 只能是 0 或 1

### 🧪 启停功能测试

**发送algorithm.config请求**:
- **主题**: `/sys/i800/edge-device-001/request`
- **消息格式**:
```json
{
  "cmdId": "config-test-001",
  "version": "1.0",
  "method": "algorithm.config", 
  "timestamp": "2025-09-18 16:00:00",
  "params": {
    "algorithmId": "algo-power-save-001",
    "runStatus": 0
  }
}
```

**成功响应**:
```json
{
  "cmdId": "config-test-001",
  "version": "1.0",
  "method": "algorithm.config",
  "timestamp": "2025-09-18 16:00:01", 
  "code": 0,
  "message": "success",
  "data": null
}
```

### 📝 配置文件更新示例

**更新前的config.yaml**:
```yaml
algo:
  runStatus: 1    # 运行中
  name: "夜间节能策略算法"
  version: "v1.2.0"
```

**更新后的config.yaml**:
```yaml
algo:
  runStatus: 0    # 已停止
  name: "夜间节能策略算法" 
  version: "v1.2.0"
```

### ⚠️ 启停错误场景

| 场景 | 错误码 | 错误信息 | 处理方法 |
|------|--------|----------|----------|
| algorithmId为空 | 1005 | 算法ID不能为空 | 检查请求参数 |
| runStatus无效 | 1005 | 运行状态值无效，只能是0(关闭)或1(开启) | 使用0或1 |
| 算法不存在 | 1006 | 算法不存在，algorithmId: xxx | 先下发算法 |
| 配置文件不存在 | 1006 | 配置文件不存在: /path/to/config.yaml | 算法部署不完整 |
| 文件读写失败 | 1003 | 读取/写入配置文件失败 | 检查文件权限 |

### 🔄 启停与查询联动测试

1. **查询当前状态**: 发送 `algorithm.show` 查看所有算法状态
2. **执行启停操作**: 发送 `algorithm.config` 修改特定算法状态
3. **验证状态变更**: 再次发送 `algorithm.show` 确认状态已更新
4. **文件系统验证**: 检查对应的 `config.yaml` 文件内容已修改

## 📊 新增服务架构说明 【v2.1.0】

### 📁 新增服务文件

```
internal/service/
├── algorithm_common.go         # 🔧 公共组件 (扩展)
│   ├── AlgorithmShowRequest    # 🆕 查询请求结构
│   ├── AlgorithmConfigRequest  # 🆕 启停请求结构
│   ├── AlgorithmShowResponseData # 🆕 查询响应数据
│   ├── AlgorithmConfig         # 🆕 config.yaml解析结构
│   ├── AlgorithmVersionExistsError # 🆕 版本已存在错误类型
│   └── CodeVersionExists = 1007 # 🆕 新增错误码
│
├── algorithm_show.go           # 🆕 算法查询服务
│   ├── AlgorithmShowService    # 查询服务类
│   ├── GetAlgorithmList()      # 获取算法列表及运行状态 
│   ├── readAlgorithmRunStatus() # 读取算法运行状态
│   ├── parseConfigFile()       # 解析config.yaml文件
│   └── buildConfigPath()       # 构建配置文件路径
│
└── algorithm_config.go         # 🆕 算法启停服务
    ├── AlgorithmConfigService  # 启停服务类
    ├── UpdateAlgorithmRunStatus() # 更新算法运行状态
    ├── queryAlgorithmFromDB()  # 数据库存在性查询
    ├── updateConfigFile()      # 更新配置文件
    └── buildConfigPath()       # 构建配置文件路径
```

### 🎯 设计原则

- **📦 模块化**: 每个MQTT方法对应独立的服务文件
- **🔄 可复用**: 配置文件路径构建等公共逻辑抽象复用
- **🛡️ 容错性**: 配置文件不存在时提供默认值和友好错误处理
- **📝 日志记录**: 所有操作详细记录，便于调试和监控

### 🏗️ 内部业务逻辑层 (`internal/`)

#### 🚀 命令行启动层 (`internal/cmd/`)
```
internal/cmd/
└── cmd.go              # HTTP服务器启动命令，配置路由和中间件
```

#### 🎮 控制器层 (`internal/controller/`)
```
internal/controller/
├── hello/
│   ├── hello.go              # Hello控制器基础定义
│   ├── hello_new.go          # Hello控制器构造函数
│   └── hello_v1_hello.go     # Hello v1版本具体处理逻辑
└── user/
    ├── user.go               # User控制器基础定义
    ├── user_new.go           # User控制器构造函数
    ├── user_v1_create.go     # 创建用户处理逻辑
    ├── user_v1_delete.go     # 删除用户处理逻辑
    ├── user_v1_get_list.go   # 获取用户列表处理逻辑
    ├── user_v1_get_one.go    # 获取单个用户处理逻辑
    └── user_v1_update.go     # 更新用户处理逻辑
```

**作用**:
- 处理HTTP请求路由
- 调用service层业务逻辑
- 按功能模块拆分文件，便于维护

#### 📊 数据模型层 (`internal/model/`)
```
internal/model/
├── do/user.go          # Data Object - 数据库操作对象
└── entity/user.go      # Entity - 数据库表对应的实体结构体
```

---

## 🆕 最新功能特性 (v2.1.0)

### 🎯 1. 智能版本管理系统
- **复合唯一约束**: 基于 `(algorithmId + algorithmVersion)` 防重复
- **版本并存**: 支持同一算法的多个版本独立存储
- **智能去重**: 相同版本请求自动跳过，避免重复下载
- **版本隔离**: 独立目录结构 `algorithm/{algorithmId}/{algorithmVersionId}/`

### 📦 2. ZIP自动解压系统
- **无缝解压**: MD5校验通过后自动解压ZIP算法包
- **安全防护**: 内置ZIP炸弹和路径遍历攻击防护
- **文件保留**: 原始ZIP包与解压文件并存
- **跨平台**: 自动识别动态库文件 (.so/.dll) 和配置文件

### 📝 3. 滚动日志系统
- **文件滚动**: 20MB自动滚动，保持日志文件大小可控
- **自动压缩**: 历史日志文件自动gzip压缩节省空间
- **保留策略**: 保留5个历史文件，30天自动清理
- **持久化**: 所有操作完整记录，便于故障排查和审计

### 🛡️ 4. 增强数据约束
- **MD5格式**: 强制32位十六进制格式校验 `CHECK(length(md5) = 32)`
- **自动时间戳**: `created_at`/`updated_at` 触发器自动维护
- **复合约束**: 数据库级别防止版本重复
- **完整性保护**: 多层校验确保数据一致性

### 🔧 5. 配置化增强
```yaml
# 新增配置项
logger:
  rotateSize: "20MB"              # 日志文件滚动大小
  rotateBackupLimit: 5            # 保留历史文件数
  rotateBackupExpire: "30d"       # 过期清理时间
  rotateBackupCompress: true      # 自动压缩

algorithm:
  downloadPath: ""                # 支持自定义下载路径
  # 新的目录结构: algorithm/{algorithmId}/{algorithmVersionId}/
```

---

## 📈 性能优化与稳定性

### ⚡ 性能改进
- **流式MD5计算**: 下载过程中实时计算，无需二次读取
- **版本预检**: 避免不必要的网络下载和文件操作
- **延迟清理**: 优化Windows平台文件句柄释放
- **并行安全**: 支持多个算法同时下载不冲突

### 🛡️ 稳定性提升
- **故障自愈**: 完整的错误回滚和清理机制
- **路径安全**: 防止ZIP解压时的安全漏洞
- **资源管理**: 确保HTTP连接和文件句柄正确释放
- **幂等性**: 支持请求重试而不产生副作用

---

## 🔄 升级指南

### 从 v1.0 升级到 v2.1

1. **数据库迁移**:
```sql
-- 新增复合唯一约束和CHECK约束
-- 系统会自动应用 data/init.sql 中的新约束
```

2. **配置更新**:
```yaml
# manifest/config/config.yaml 新增日志配置
logger:
  path: "logs"
  rotateSize: "20MB"
  rotateBackupLimit: 5
  rotateBackupExpire: "30d"
  rotateBackupCompress: true
```

3. **目录结构适配**:
- 旧版: `runtime/{algorithmId}/{md5}/`
- 新版: `runtime/algorithm/{algorithmId}/{algorithmVersionId}/`
- 自动向前兼容，无需手动迁移

### 兼容性说明
- ✅ MQTT接口完全向后兼容
- ✅ HTTP API接口保持不变
- ✅ 数据库表结构向前兼容
- ✅ 配置文件向后兼容（新增项可选）

---

## 📋 版本更新日志

### 🚀 v2.2.0 (2025-09-24) - MQTT可靠性增强与生产级稳定性升级

#### 🎯 核心问题解决
- **启动阻塞修复**: 解决MQTT连接阻塞导致程序启动卡死的关键问题
- **假订阅现象修复**: 修复MQTT重连后出现的"假订阅"问题，确保消息能够正常接收
- **双状态订阅管理**: 创新性地引入 `subscriptions` 和 `savedSubscriptions` 双状态管理机制
- **异步连接启动**: MQTT连接采用异步非阻塞方式，不影响其他服务启动

#### 🛡️ 可靠性增强
- **自动重连机制**: 完善的MQTT自动重连，支持指数退避和最大重连间隔控制
- **健康检查系统**: 30秒间隔的连接健康监控，主动发现和修复连接问题
- **参数优化配置**: 全面的MQTT连接参数调优，包括心跳、超时、重连等关键参数
- **会话状态管理**: 支持 `cleanSession=false` 保持会话状态，确保离线消息不丢失

#### 🔧 技术架构改进
- **OnConnectionLost优化**: 连接丢失时正确保存和清理订阅状态
- **订阅状态同步**: 确保订阅状态与实际MQTT连接状态保持一致
- **故障隔离设计**: MQTT连接失败不影响HTTP服务等其他功能模块
- **监控诊断完善**: 详细的连接状态日志，便于生产环境问题排查

#### 📊 性能与稳定性提升
- **启动时间**: 从不确定阻塞 → 稳定2-3秒
- **故障恢复**: 从手动重启 → 自动重连(5秒间隔)
- **消息可靠性**: 从连接丢失后失效 → 自动恢复订阅
- **系统可用性**: 从MQTT依赖 → 服务解耦，部分功能可独立运行

#### 🧪 生产验证
- **设备部署测试**: 解决实际I-800设备部署中的连接稳定性问题
- **网络异常测试**: 验证各种网络中断场景下的自动恢复能力
- **长期运行验证**: 确保24/7运行环境下的连接稳定性

---

### � v2.1.0 (2025-01-20) - 算法管理功能大幅增强

#### 🆕 新增功能
- **🔍 算法查询功能** (`algorithm.show`): 完整的算法列表查询，支持实时运行状态读取
- **⚙️ 算法启停控制** (`algorithm.config`): 远程控制算法启动/停止状态，支持配置文件持久化
- **� 配置文件管理**: 自动解析和修改算法目录下的 `config.yaml` 文件
- **🎯 智能路径处理**: 跨平台配置文件路径构建，Windows/Linux环境兼容

#### 🔧 重要Bug修复
- **JSON兼容性修复**: `fileSize` 字段从 `int64` 改为 `float64`，解决客户端JSON解析兼容性问题
- **版本冲突处理**: 新增 `AlgorithmVersionExistsError` 错误类型，版本重复时返回明确错误信息 (错误码1007)
- **版本更新Bug修复**: 修复版本更新时新下载文件被意外删除的严重bug
- **错误响应完善**: 所有MQTT方法添加完整错误响应机制，防止客户端卡死

#### 🎨 架构优化
- **服务模块化**: 新增 `algorithm_show.go` 和 `algorithm_config.go` 独立服务模块
- **公共组件扩展**: `algorithm_common.go` 增加查询和配置相关数据结构
- **错误处理增强**: 完善的异常捕获和友好错误提示机制
- **日志记录改进**: 详细的操作日志和错误追踪

#### 🧪 测试覆盖
- **功能测试**: algorithm.show 查询功能完整测试用例
- **启停测试**: algorithm.config 启动/停止状态控制验证
- **异常测试**: 配置文件不存在、格式错误等边界情况处理
- **集成测试**: MQTT消息收发和状态同步验证

---

### � v2.0.0 (2025-01-15) - 核心架构重构

#### 🆕 新增功能
- **📦 算法下发功能** (`algorithm.add`): HTTP/FTP多协议下载，MD5校验，ZIP自动解压
- **🗑️ 算法删除功能** (`algorithm.delete`): 数据库清理和文件系统清理
- **🔄 版本管理机制**: 基于 `algorithmId + algorithmVersion` 的复合唯一约束
- **� 滚动日志系统**: 20MB自动轮转，压缩存储，30天自动清理

#### 🏗️ 架构特性
- **GoFrame v2.9.3**: 现代化Go Web框架，强大的ORM和中间件支持
- **SQLite3数据库**: 轻量级嵌入式数据库，复合唯一约束和CHECK约束
- **MQTT Eclipse Paho v1.5.0**: 稳定的MQTT客户端，支持自动重连
- **跨平台部署**: Windows开发环境和Linux生产环境完全兼容

---

## 📞 技术支持

如有问题，请查看：
1. 📝 **日志文件**: `logs/app-YYYYMMDD.log`
2. 🔍 **错误码说明**: 参考上述错误码表格
3. 🛠️ **配置检查**: 确认MQTT连接和路径配置
4. 📊 **数据库状态**: 检查SQLite文件权限和表结构

---

**作用**:
- **entity**: 数据库表映射的Go结构体（通过GoFrame工具自动生成）
- **do**: 数据库操作时使用的对象

#### 💾 数据访问层 (`internal/dao/`)
```
internal/dao/
├── user.go             # User数据访问对象定义
└── internal/user.go    # User数据访问内部实现
```

**作用**:
- 封装数据库CRUD操作
- 提供统一的数据访问接口

#### 📦 其他内部模块
```
internal/
├── consts/             # 常量定义
├── logic/              # 业务逻辑层
├── service/            # 服务层
└── packed/             # 资源文件打包
```

### 🚀 部署运维相关

#### 📋 部署配置 (`manifest/`)
```
manifest/
├── config/
│   └── config.yaml                 # 应用配置文件
├── deploy/kustomize/               # Kubernetes部署配置
│   ├── base/                      # 基础部署配置
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── kustomization.yaml
│   └── overlays/develop/          # 开发环境覆盖配置
│       ├── configmap.yaml
│       ├── deployment.yaml
│       └── kustomization.yaml
├── docker/                        # Docker容器化
│   ├── Dockerfile
│   └── docker.sh
├── i18n/                          # 国际化文件
└── protobuf/                      # Protocol Buffers定义
```

#### 🔧 开发工具 (`hack/`)
```
hack/
├── config.yaml         # 开发配置
├── hack-cli.mk         # CLI工具makefile
└── hack.mk             # 开发脚本makefile
```

#### 📱 静态资源 (`resource/`)
```
resource/
├── public/                        # 公共静态文件
│   ├── html/                     # HTML文件
│   ├── plugin/                   # 插件文件
│   └── resource/                 # 资源文件
│       ├── css/                  # 样式文件
│       ├── image/                # 图片资源
│       └── js/                   # JavaScript文件
└── template/                     # 模板文件
```

#### 🛠️ 工具函数 (`utility/`)
```
utility/                          # 工具函数库
```

## 🌟 架构特点

1. **🏗️ 分层架构**: 清晰的分层设计，职责分离
2. **📦 模块化设计**: 按业务模块（hello、user）组织代码
3. **🔄 版本管理**: API按版本（v1）管理，便于向后兼容
4. **🤖 自动生成**: entity和dao代码可通过GoFrame工具自动生成
5. **🐳 容器化支持**: 完整的Docker和Kubernetes部署配置
6. **📏 标准化**: 遵循GoFrame框架的项目结构规范

## 🔍 主要功能模块

### Hello模块
- `GET /hello` - 简单的Hello接口

### User模块  
- `POST /user` - 创建用户
- `GET /user` - 获取用户列表（支持年龄和状态筛选）
- `GET /user/{id}` - 获取单个用户信息
- `PUT /user/{id}` - 更新用户信息
- `DELETE /user/{id}` - 删除用户

## 📖 开发指南

1. **API接口开发流程**:
   ```
   1. 在 api/模块名/v1/ 中定义接口结构体
   2. 在 internal/controller/模块名/ 中实现控制器逻辑
   3. 在 internal/service/ 中实现业务逻辑（可选）
   4. 在 internal/dao/ 中实现数据访问逻辑
   ```

2. **数据库表生成**:
   ```bash
   # 使用GoFrame CLI工具生成entity和dao
   gf gen dao
   ```

3. **API文档更新**:
   ```bash
   # 生成OpenAPI文档
   gf gen openapi
   ```

这是一个典型的GoFrame企业级项目结构，具备良好的可扩展性、可维护性和标准化程度。

## 🏷️ 智能设备注册系统 【v2.3.0 - 多设备部署解决方案】

### 🎯 功能概述

为解决多设备部署中的设备识别和网络适配问题，实现了完整的智能设备注册功能，支持即插即用和自动化管理。

### 📋 核心需求解决

1. **动态设备ID**: 使用网卡MAC地址作为设备唯一标识，避免手动配置冲突
2. **智能网卡检测**: 自动检测可用网卡，过滤虚拟网卡，确保网络连通性
3. **连通性验证**: 测试网卡与MQTT Broker的连通性，选择最优网络接口
4. **自动注册**: MQTT连接成功和重连后自动发送设备注册消息
5. **ClientID格式**: edge-{MAC地址} 格式，符合设备命名规范

### 🛠️ 技术架构

#### 1. 网络检测服务 (`internal/service/network.go`)

**三种检测模式**:
- 🔍 **auto模式（推荐）**: 遍历所有网卡，智能过滤，连通性测试
- ⚙️ **config模式**: 按配置的网卡优先级依次检测
- 🔧 **manual模式**: 使用配置文件中指定的网卡信息

**智能过滤算法**:
```go
// 虚拟网卡过滤模式
virtualPatterns := []string{
    "docker", "veth", "br-", "virbr", "vmnet", "vbox", 
    "tap", "tun", "lo", "dummy", "bond", "team", "vlan"
}

// 连通性测试机制
dialer := &net.Dialer{
    LocalAddr: &net.TCPAddr{IP: net.ParseIP(localIP)},
    Timeout:   timeout,
}
```

#### 2. 设备注册服务 (`internal/service/device_register.go`)

**核心特性**:
- 📤 构建符合MQTT接口定义的注册消息
- 🔄 支持重试机制，最大3次重试，可配置间隔
- ⏰ 可配置的重试间隔和超时时间
- 📊 详细的日志记录和错误处理
- 🏷️ 自动生成UUID作为消息cmdId

**注册消息格式**:
```json
{
  "cmdId": "uuid-1234",
  "version": "1.0",
  "method": "event.register",
  "timestamp": "2025-09-25 10:01:18",
  "params": {
    "deviceModule": "I-800-RK",
    "deviceId": "84-A9-3E-1C-75-5F",  // MAC地址
    "heartBeat": 60,                   // 心跳间隔
    "IP": "10.77.26.2"                // 网卡IP地址
  }
}
```

#### 3. MQTT服务深度集成 (`internal/service/mqtt.go`)

**集成点**:
- 🏁 **初始化阶段**: 执行网卡检测，生成动态ClientID
- 🔗 **OnConnect回调**: 首次连接成功后自动发送注册消息
- 🔄 **OnReconnect回调**: 重连成功后重新发送注册消息
- 📊 **状态管理**: 维护网络接口状态，支持网络变化检测

**状态区分机制**:
```go
// 正确区分首次连接和重连
isReconnect := !mqttService.isFirstConnect
if mqttService.isFirstConnect {
    mqttService.isFirstConnect = false
}
```

### ⚙️ 配置系统增强

#### 设备网络配置
```yaml
device:
  id: "F2-D5-4F-C7-2B-AD"  # 运行时动态更新为实际MAC
  module: "I-800-RK"        # 设备模块名
  
  # 🌐 网卡检测配置
  network:
    mode: "auto"            # 检测模式: auto/manual/config
    
    # 手动指定网卡配置 (mode=manual时使用)
    interface: ""           # 网卡名称，如 "eth0", "wlan0"
    mac: ""                # MAC地址
    ip: ""                 # IP地址
    
    # 网卡优先级配置 (mode=config时使用)
    priority:
      - "eth0"             # 有线网卡优先
      - "ens"              # Ubuntu网卡
      - "enp"              # 新版Ubuntu网卡
      - "wlan"             # 无线网卡
      - "wlp"              # 新版无线网卡
    
    # 连通性测试配置
    connectivity:
      timeout: 3           # 连接测试超时时间(秒)
      retries: 2           # 测试重试次数
```

#### MQTT注册配置
```yaml
mqtt:
  # ClientID动态生成: edge-{MAC地址}
  
  # 📡 主题配置
  topics:
    device:
      register: "/sys/i800/{deviceId}/event/register"
  
  # 🏷️ 设备注册配置
  register:
    enable: true           # 启用设备注册功能
    onConnect: true        # 连接成功时自动注册
    onReconnect: true     # 重连成功时自动注册
    retryInterval: 30     # 注册失败重试间隔(秒)
    maxRetries: 3         # 最大重试次数
```

### 📊 实际测试验证

#### 网卡检测结果
```bash
🔍 开始自动检测可用网卡，共找到 8 个网络接口
⏭️ 跳过无效网卡: 本地连接* 1 (虚拟网卡)
⏭️ 跳过无效网卡: 本地连接* 2 (虚拟网卡)
✅ 连通性测试成功: 10.77.26.2 -> tcp://10.77.26.5:16003
✅ 自动选择网卡: 以太网, MAC: 84-A9-3E-1C-75-5F, IP: 10.77.26.2
```

#### 设备注册结果
```bash
🏷️ 设备ID: 84-A9-3E-1C-75-5F
🔗 MQTT客户端ID: edge-84-A9-3E-1C-75-5F
🟢 MQTT连接成功 {"connectionType":"首次连接",...}
🏷️ 开始设备注册 (首次连接): DeviceId=84-A9-3E-1C-75-5F
📤 设备注册消息发送成功
   📡 主题: /sys/i800/84-A9-3E-1C-75-5F/event/register
   🏷️ 设备ID: 84-A9-3E-1C-75-5F
   🌐 IP地址: 10.77.26.2
   💻 网卡: 以太网 (84-A9-3E-1C-75-5F)
```

### 🎯 功能亮点

#### 1. 智能网卡检测
- **全自动化**: 无需手动配置，自动检测最佳网卡
- **智能过滤**: 自动排除Docker、虚拟机、回环等虚拟网卡
- **连通性保证**: 确保选择的网卡能够正常连接MQTT服务器
- **多平台兼容**: Windows开发环境和Linux生产环境完全适配

#### 2. 多模式灵活支持
- **auto模式**: 完全自动检测，推荐生产环境使用
- **config模式**: 按优先级配置检测，适合特定网络环境
- **manual模式**: 手动指定网卡信息，适合固定网络配置

#### 3. 企业级可靠性
- **重试机制**: 注册失败自动重试，最大3次，可配置间隔
- **状态同步**: 网卡状态变化时自动重新检测和注册
- **错误处理**: 完善的异常捕获、日志记录和故障恢复
- **监控友好**: 详细的操作日志，便于生产环境监控

#### 4. 生产部署优化
- **即插即用**: 设备接入网络后自动完成检测和注册
- **批量部署**: 多设备部署无需逐一配置，支持大规模部署
- **网络适配**: 自动适配不同的网络环境和网卡类型
- **零配置**: 程序启动自动完成所有网络相关配置

### 🔄 系统集成效果

#### 与MQTT可靠性增强的协同
- ✅ **异步启动**: 网卡检测和MQTT连接不阻塞主程序启动
- ✅ **自动重连**: 网络恢复后自动重新注册设备
- ✅ **健康检查**: 30秒间隔检查，及时发现网络状态变化
- ✅ **状态同步**: 连接状态与注册状态保持一致

#### 与算法管理功能的融合
- ✅ **设备标识**: 使用真实MAC地址作为唯一设备标识
- ✅ **主题动态化**: 算法下发主题自动适配真实设备ID
- ✅ **多设备支持**: 每个设备独立的主题空间和消息隔离
- ✅ **运维统一**: 平台可通过MAC地址精确管理设备

### 📈 业务价值与效益

#### 1. 部署效率提升
- **零配置部署**: 设备启动即可自动接入，无需网络配置
- **批量扩展**: 支持大规模设备部署，降低运维成本
- **即插即用**: 新设备接入后自动完成注册和配置

#### 2. 运维管理优化
- **精确识别**: 平台通过MAC地址精确识别和管理设备
- **状态透明**: 设备注册消息提供完整的网络和状态信息
- **故障定位**: 详细日志帮助快速定位网络和连接问题

#### 3. 系统扩展能力
- **动态接入**: 支持设备动态加入和退出网络
- **网络切换**: 网卡变更时自动重新适配和注册
- **平台兼容**: 标准化的注册消息格式，便于平台集成

### 🚀 技术演进路线

#### 短期优化 (v2.4.0)
- **心跳机制**: 实现周期性设备状态心跳上报
- **网络监控**: 实时监控网卡状态变化并自动适配
- **注册缓存**: 优化注册逻辑，避免重复注册

#### 中期扩展 (v2.5.0)
- **多网卡支持**: 同时使用多个网卡提高可靠性
- **网络质量评估**: 基于延迟、丢包率选择最佳网卡
- **配置热更新**: 支持运行时动态调整网络配置

#### 长期规划 (v3.0.0)
- **智能路由**: 根据网络质量自动选择最优通信路径
- **边缘协作**: 设备间直接通信和协作能力
- **安全增强**: 设备身份认证和通信加密

---

**总结**: 智能设备注册系统完美解决了多设备部署的核心痛点，实现了真正的即插即用和自动化管理。结合MQTT可靠性增强和算法管理功能，构建了完整的企业级边缘设备管理解决方案，为大规模IoT部署提供了坚实的技术基础。
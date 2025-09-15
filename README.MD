# I-800 边缘设备算法管理系统

> 基于 **GoFrame v2.### 💼 核心业务层
```
internal/
├── service/              ⭐ 业务逻辑层
│   ├── mqtt.go              # MQTT通信服务
│   ├── algorithm_common.go  # 算法通用数据结构和错误码
│   ├── algorithm_add.go     # 算法添加/下发核心逻辑
│   └── algorithm_delete.go  # 算法删除核心逻辑
├── controller/          📡 HTTP API控制器
│   ├── algorithm/       # 算法管理API
│   └── user/           # 用户管理API
├── dao/                💾 数据访问层 (自动生成)
├── model/              📋 数据模型层
│   ├── entity/         # 数据库表实体缘设备算法管理系统，支持 MQTT 协议的云端算法下发、文件管理和数据同步功能。

# 详细版
internal/service/
├── algorithm_common.go     # 🔧 公共组件
│   ├── AlgorithmAddRequest
│   ├── AlgorithmDeleteRequest  
│   ├── AlgorithmReply
│   └── 错误码常量 (CodeSuccess, CodeDownloadFailed等)
│
├── algorithm_add.go        # ⬇️ 算法添加服务
│   ├── AlgorithmDownloadService
│   ├── DownloadAlgorithmFile()
│   ├── SyncAlgorithmToDatabase()
│   └── cleanupDirectoryOnFailure()
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

🎯 架构优势
- 🧩 单一职责 - 每个文件专注特定功能
- 🔄 易于扩展 - 添加新算法操作只需新增文件
- 🛠️ 便于维护 - 修改特定功能不影响其他模块
- 📦 代码复用 - 公共组件统一管理，避免重复
- 🧪 便于测试 - 每个服务可以独立进行单元测试

## 🚀 快速开始

```bash
# 启动项目
go run .

# 访问API文档
http://127.0.0.1:8000/swagger/
```

## ⭐ 核心功能

- 🔗 **MQTT算法下发**: 支持云端通过MQTT协议下发算法文件
- 📁 **文件管理**: 自动下载、MD5校验、版本管理
- 🗃️ **数据同步**: SQLite数据库存储算法信息
- 🔄 **跨平台支持**: Windows开发、Linux部署自适应
- 📊 **完整API**: HTTP RESTful API (用户管理、算法查询)

## 🏗️ 系统架构

### 技术栈
- **框架**: GoFrame v2.9.3
- **数据库**: SQLite3 (嵌入式)
- **通信协议**: MQTT + HTTP
- **部署**: Docker + Kubernetes

### 分层架构
```
MQTT协议层 → Service业务层 → DAO数据层 → SQLite数据库
HTTP API层 → Controller控制层 → Service业务层 ↗
```

## 📁 项目结构详解

### � 核心业务层
```
internal/
├── service/              ⭐ 业务逻辑层
│   ├── mqtt.go          # MQTT通信服务
│   └── algorithm_add.go # 算法下发核心逻辑
├── controller/          📡 HTTP API控制器
│   ├── algorithm/       # 算法管理API
│   └── user/           # 用户管理API
├── dao/                💾 数据访问层 (自动生成)
├── model/              📋 数据模型层
│   ├── entity/         # 数据库表实体
│   └── do/             # 数据操作对象
└── cmd/                🚀 启动命令层
```

### 🔧 配置和部署
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
  
# MQTT服务配置
mqtt:
  broker: "tcp://127.0.0.1:1883"  # 你的MQTT服务器地址
  clientId: "goframe-edge-device"
  keepAlive: 60

# 算法下载配置  
algorithm:
  downloadPath: ""  # 留空自动适配平台路径
  # Windows: ./runtime/
  # Linux: /usr/runtime/
```

### 跨平台路径自适应

- **开发环境(Windows)**: `./runtime/{algorithmId}/{md5}/`
- **生产环境(Linux)**: `/usr/runtime/{algorithmId}/{md5}/`
- **自定义路径**: 在配置文件中指定 `algorithm.downloadPath`

## 🔄 MQTT算法下发执行流程详解

### 📋 整体架构流程

```
程序启动 
   ↓
cmd/cmd.go 调用 startMQTTAlgorithmService()
   ↓
获取 service.Mqtt() 单例实例 (mqtt.go)
   ↓
调用 StartAlgorithmMessageListener() 订阅配置化主题
   ↓
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
// 功能: DownloadAlgorithmFile()

步骤1: 创建目录结构 {downloadPath}/{algorithmId}/{md5}/
步骤2: HTTP GET 下载文件到本地
步骤3: 实时计算MD5哈希值
步骤4: 校验MD5是否匹配
步骤5: 失败时自动清理不完整文件

错误处理:
- 下载失败 → 错误码 1001
- MD5校验失败 → 错误码 1002  
- 文件系统错误 → 错误码 1003
```

#### 4. **数据库同步** 🗃️
```go
// 位置: internal/service/algorithm_add.go  
// 功能: SyncAlgorithmToDatabase()

重要逻辑: 只有文件下载成功并通过MD5校验，才会写入数据库！

算法覆盖逻辑:
- 检查算法ID是否已存在
- 如果是新算法 → INSERT新记录
- 如果算法已存在 → 实现完全覆盖:
  * 无论版本是否相同，都清理旧算法的完整目录
  * UPDATE数据库记录（包括版本、MD5、文件路径等）
  * 记录版本变更或覆盖操作的日志

处理逻辑:
- 新算法: INSERT新记录到数据库
- 存在算法: 清理旧目录 → UPDATE记录 → 存储新的元数据
- 支持相同版本覆盖（比如算法代码修正但版本号未变）

错误处理:
- 数据库操作失败 → 错误码 1004
- 旧文件清理失败 → 记录警告但不阻止更新
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
- **消息**:
```json
{
  "cmdId": "test-001",
  "version": "1.0",
  "method": "algorithm.add",
  "timestamp": "2025-09-12 17:30:00",
  "params": {
    "algorithmId": "algo-test-001",
    "algorithmName": "测试算法",
    "algorithmVersion": "v1.0.0",
    "algorithmVersionId": "version-001",
    "algorithmDataUrl": "https://httpbin.org/base64/SGVsbG8gV29ybGQ=",
    "fileSize": 12,
    "lastModifyTime": "2025-09-12 17:30:00",
    "md5": "65a8e27d8879283831b664bd8b7f0ad4"
  }
}
```

**订阅响应结果**:
- **主题**: `/sys/i800/edge-device-001/reply`
- **成功响应示例**:
```json
{
  "cmdId": "test-001",
  "version": "1.0",
  "method": "algorithm.add", 
  "timestamp": "2025-09-12 17:30:01",
  "code": 0,
  "message": "算法添加成功",
  "data": {
    "algorithmId": "algo-test-001",
    "version": "v1.0.0",
    "localPath": "./runtime/algo-test-001/65a8e27d8879283831b664bd8b7f0ad4/SGVsbG8gV29ybGQ="
  }
}
```

### 3. 验证测试结果

✅ **成功标志**:
1. **目录创建**: 项目根目录下出现 `runtime/` 文件夹
2. **文件下载**: `runtime/algo-test-001/65a8e27d8879283831b664bd8b7f0ad4/` 目录下有下载的算法文件  
3. **数据库记录**: SQLite数据库中 `algorithm` 表新增记录
4. **MQTT回复**: 收到 `code: 0` 的成功响应

## 🔍 错误码详细说明

| 错误码 | 错误类型 | 触发条件 | 排查方法 | 对应处理阶段 |
|-------|---------|---------|---------|-------------|
| 0 | 成功 | 算法下发完整流程成功 | - | ✅ 全流程完成 |
| 1001 | 下载失败 | HTTP下载超时/网络异常/URL无效 | 检查网络连接、URL可访问性、服务器响应状态 | ❌ 文件下载阶段 |
| 1002 | MD5校验失败 | 下载文件MD5与请求不符 | 确认源文件完整性、重新计算MD5值、检查传输过程 | ❌ 文件校验阶段 |
| 1003 | 文件系统错误 | 磁盘空间不足/目录权限不够/路径创建失败 | 检查磁盘空间、目录读写权限、路径格式 | ❌ 文件系统操作 |
| 1004 | 数据库错误 | SQLite连接失败/SQL执行异常/表结构问题 | 检查数据库文件权限、表结构、SQL语法 | ❌ 数据库同步阶段 |
| 1005 | 参数错误 | algorithmId/algorithmDataUrl/md5缺失 | 检查MQTT消息格式和必要参数完整性 | ❌ 参数验证阶段 |
| 1006 | 算法不存在 | 删除不存在的算法ID | 检查算法ID是否正确、是否已被删除 | ❌ 算法删除阶段 |

### 🔄 错误处理流程

```
MQTT请求接收
    ↓
参数验证 (错误码1005)
    ↓
创建目录结构 (/usr/runtime/algorithmId/md5sum/)
    ↓
文件下载 (错误码1001) ❌→ 清理空目录
    ↓  
MD5校验 (错误码1002) ❌→ 删除文件 + 清理空目录
    ↓
文件系统操作 (错误码1003) ❌→ 清理空目录
    ↓
数据库同步 (错误码1004)
    ↓
响应发送 (错误码0-成功)

🧹 智能清理机制:
- cleanupDirectoryOnFailure() 自动清理创建的空目录
- 避免失败后留下 runtime/algorithmId/ 空文件夹  
- 只删除确认为空的目录，保护现有数据
- 🔧 **Windows文件句柄优化**: 确保HTTP响应体正确关闭，防止"文件正在使用"错误
- 🔄 **重试机制**: 带延迟重试的目录删除，提高清理成功率
- ⏱️ **延迟清理**: 等待文件句柄释放后再执行清理操作
```

### 🔄 **算法覆盖机制**

**覆盖逻辑**:
```
如果算法已存在 (相同algorithmId):
  ✅ 完全覆盖模式
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
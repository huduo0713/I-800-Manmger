# I-800 边缘设备算法管理系统

> 基于 **GoFrame v2.9.3** 构建的边缘设备算法管理系统，支持 MQTT 协议的云端算法下发、文件管理、自动解压和数据同步功能。

## ⭐ 核心功能

- 🔗 **MQTT算法下发**: 支持云端通过MQTT协议下发算法文件
- 📁 **智能文件管理**: 自动下载、MD5校验、版本管理、ZIP自动解压
- 🗃️ **数据同步**: SQLite数据库存储算法信息，复合唯一约束
- 🔄 **跨平台支持**: Windows开发、Linux部署自适应
- 📊 **完整API**: HTTP RESTful API (用户管理、算法查询)
- 🎯 **版本控制**: 基于算法ID+版本的智能去重机制
- 📝 **滚动日志**: 20MB文件滚动、自动压缩、持久化存储
- 🛡️ **复合校验**: MD5完整性 + 版本唯一性 + 数据库约束 + 文件结构校验

## 🚀 快速开始

```bash
# 启动项目
go run .

# 访问API文档
http://127.0.0.1:8000/swagger/
```

## 🏗️ 系统架构

### 技术栈
- **框架**: GoFrame v2.9.3
- **数据库**: SQLite3 (嵌入式) - 增强约束和触发器
- **通信协议**: MQTT (Eclipse Paho v1.5.0) + HTTP
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
  
# MQTT服务配置
mqtt:
  broker: "tcp://127.0.0.1:1883"  # 你的MQTT服务器地址
  clientId: "goframe-edge-device"
  keepAlive: 60

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
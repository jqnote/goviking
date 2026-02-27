# GoViking 项目架构文档

## 1. 项目概述

GoViking 是 OpenViking (火山引擎开源的 AI Agent 上下文数据库) 的 Go 语言实现。

### 1.1 核心特性

- **文件系统式上下文管理 (AGFS)** - 像管理本地文件一样管理 Agent 的记忆、资源和技能
- **层级上下文加载 (L0/L1/L2)** - 三层结构，按需加载，节省 Token 成本
- **语义检索** - 支持向量嵌入和混合搜索
- **会话管理** - 自动提取长期记忆
- **多 LLM 提供商支持** - OpenAI、Anthropic 等

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI (cmd/goviking)                    │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│    Server     │   │    Client     │   │   Config     │
│   (REST API)  │   │     SDK       │   │   Management │
└───────────────┘   └───────────────┘   └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Context   │  │   Session   │  │  Retrieval  │        │
│  │  Service   │  │  Service    │  │  Service    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│     Core      │   │     AGFS      │   │     LLM      │
│ (Context DB)  │   │  (FileSystem) │   │  (Providers) │
└───────────────┘   └───────────────┘   └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Storage (SQLite)                        │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 核心模块

### 3.1 pkg/storage - 存储层

**职责**: 数据持久化

**关键文件**:
- `interface.go` - 存储接口定义
- `models.go` - 数据模型
- `sqlite.go` - SQLite 实现

**核心接口**:
```go
type StorageInterface interface {
    // Context CRUD
    CreateContext(ctx context.Context, context *Context) error
    GetContext(ctx context.Context, id string) (*Context, error)
    QueryContexts(ctx context.Context, opts QueryOptions) ([]Context, error)

    // Session CRUD
    CreateSession(ctx context.Context, session *Session) error
    GetSession(ctx context.Context, id string) (*Session, error)

    // Memory CRUD
    CreateMemory(ctx context.Context, memory *Memory) error
    QueryMemories(ctx context.Context, opts QueryOptions) ([]Memory, error)
}
```

### 3.2 pkg/core - 核心上下文数据库

**职责**: 层级上下文管理、上下文构建、窗口管理

**关键文件**:
- `context.go` - 上下文类型和模型
- `tier.go` - L0/L1/L2 层级管理
- `builder.go` - 上下文构建器
- `window.go` - 上下文窗口管理
- `compression.go` - 上下文压缩

**核心概念**:
```go
// 上下文层级
const (
    TierL0 ContextTier = iota  // 核心上下文，始终加载
    TierL1                      // 按需加载
    TierL2                      // 归档上下文
)

// 上下文类型
type ContextType string
const (
    ContextTypeSkill   ContextType = "skill"
    ContextTypeMemory  ContextType = "memory"
    ContextTypeResource ContextType = "resource"
)

// 上下文分类
type Category string
const (
    CategoryPatterns    Category = "patterns"
    CategoryCases      Category = "cases"
    CategoryProfile    Category = "profile"
    CategoryPreferences Category = "preferences"
    CategoryEntities  Category = "entities"
    CategoryEvents    Category = "events"
)
```

### 3.3 pkg/agfs - Agent 图文件系统

**职责**: 类似文件系统的上下文管理

**关键文件**:
- `agfs.go` - 核心实现
- `dir.go` - 目录操作
- `file.go` - 文件操作
- `relations.go` - 关系管理
- `context.go` - 上下文集成

**核心概念**:
```go
// 文件类型
type FileType string
const (
    FileTypeMemory   FileType = "memory"
    FileTypeResource FileType = "resource"
    FileTypeSkill   FileType = "skill"
    FileTypeDirectory FileType = "directory"
)

// 条目 (文件或目录)
type Entry struct {
    Name     string    `json:"name"`
    Path     string    `json:"path"`
    IsDir    bool      `json:"is_dir"`
    Size     int64     `json:"size"`
    FileType FileType  `json:"file_type"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**URI 规范**:
- `viking://agent/skills/` - Agent 技能
- `viking://agent/memories/` - Agent 记忆
- `viking://user/memories/` - 用户记忆
- `viking://resources/` - 资源文件

### 3.4 pkg/retrieval - 检索引擎

**职责**: 语义搜索、混合搜索、检索轨迹

**关键文件**:
- `types.go` - 类型定义
- `semantic.go` - 语义搜索
- `hybrid.go` - 混合搜索
- `retriever.go` - 检索器
- `trajectory.go` - 检索轨迹
- `embedder.go` - 向量化接口

**检索模式**:
```go
type RetrieverMode string
const (
    RetrieverModeThinking RetrieverMode = "thinking"  // 深度思考模式
    RetrieverModeQuick    RetrieverMode = "quick"      // 快速检索模式
)
```

### 3.5 pkg/session - 会话管理

**职责**: 会话生命周期、消息跟踪、内存提取

**关键文件**:
- `session.go` - 会话管理
- `extractor.go` - 内存提取

**会话状态**:
```go
type State string
const (
    StateActive  State = "active"   // 活跃
    StatePaused  State = "paused"   // 暂停
    StateClosed  State = "closed"    // 已关闭
)
```

### 3.6 pkg/llm - LLM 提供商

**职责**: 多提供商支持

**关键文件**:
- `provider.go` - 提供商接口
- `openai.go` - OpenAI 实现
- `anthropic.go` - Anthropic 实现
- `factory.go` - 工厂函数

**接口定义**:
```go
type Provider interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (StreamReader, error)
    Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
    Close() error
}
```

### 3.7 pkg/config - 配置管理

**职责**: YAML 配置、环境变量覆盖

**配置结构**:
```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Storage  StorageConfig  `mapstructure:"storage"`
    LLM      LLMConfig      `mapstructure:"llm"`
    Retrieval RetrievalConfig `mapstructure:"retrieval"`
}
```

### 3.8 pkg/server - HTTP 服务器

**职责**: REST API

**端点**:
- `GET /health` - 健康检查
- `GET/POST /api/v1/contexts` - 上下文 CRUD
- `GET/POST /api/v1/sessions` - 会话 CRUD

### 3.9 pkg/client - 客户端 SDK

**职责**: 同步/异步客户端

### 3.10 pkg/service - 服务层

**职责**: 业务逻辑、验证

### 3.11 pkg/utils - 工具函数

**职责**: 通用工具 (ID生成、时间、字符串)

---

## 4. 数据流

### 4.1 上下文创建流程

```
用户请求
    │
    ▼
CLI/Client ──▶ Server ──▶ Service ──▶ Core (Builder)
    │                                 │
    │                                 ▼
    │                          Storage (SQLite)
    │                                 │
    ◀────────────────────────────────┘
```

### 4.2 检索流程

```
用户查询
    │
    ▼
Retrieval ──▶ AGFS (Tree) ──▶ Semantic Search
    │                       │
    │                       ▼
    │                 Vector Embedding
    │                       │
    ▼                       ▼
Hybrid Search ◀────────── Results
    │
    ▼
返回结果 + 轨迹
```

---

## 5. 依赖关系

```
github.com/spf13/cobra      // CLI 框架
github.com/spf13/viper     // 配置管理
github.com/gorilla/mux      // HTTP 路由
github.com/mattn/go-sqlite3 // SQLite 驱动
github.com/google/uuid      // UUID 生成
```

---

## 6. 测试覆盖

| 包 | 测试状态 |
|---|---------|
| agfs | ✓ |
| core | ✓ |
| retrieval | ✓ |
| storage | ✓ |
| client | ✓ |
| config | ✓ |
| session | ✓ |
| service | ✓ |
| utils | ✓ |

---

## 7. 项目结构

```
goviking/
├── cmd/
│   └── goviking/          # CLI 入口
│       └── main.go
├── pkg/
│   ├── agfs/              # 文件系统上下文
│   ├── client/            # 客户端 SDK
│   ├── config/            # 配置管理
│   ├── core/              # 核心数据库
│   ├── llm/               # LLM 提供商
│   ├── message/           # 消息处理
│   ├── retrieval/          # 检索引擎
│   ├── server/            # HTTP 服务器
│   ├── service/           # 服务层
│   ├── session/           # 会话管理
│   ├── storage/           # 存储后端
│   └── utils/             # 工具函数
├── go.mod
└── README.md
```

---

## 8. 后续开发建议

1. **gRPC 支持** - 添加 gRPC 服务
2. **更多 LLM 提供商** - 添加 DeepSeek、Gemini 等
3. **向量数据库集成** - 支持 Qdrant、Chroma
4. **WebSocket 支持** - 实时推送
5. **指标监控** - Prometheus 集成

# GoViking API 文档

## 1. 概述

GoViking 提供两种 API 访问方式:
- **HTTP REST API** - 远程调用
- **Go SDK** - 直接库调用

---

## 2. REST API

### 2.1 健康检查

```bash
GET /health
```

**响应**:
```json
{
  "status": "ok",
  "time": "2026-02-26T12:00:00Z"
}
```

### 2.2 上下文管理

#### 创建上下文

```bash
POST /api/v1/contexts
Content-Type: application/json

{
  "uri": "viking://agent/skills/search",
  "type": "skill",
  "name": "Search Skill",
  "content": "..."
}
```

#### 获取上下文

```bash
GET /api/v1/contexts/{id}
```

#### 列出上下文

```bash
GET /api/v1/contexts
```

#### 删除上下文

```bash
DELETE /api/v1/contexts/{id}
```

### 2.3 会话管理

#### 创建会话

```bash
POST /api/v1/sessions
Content-Type: application/json

{
  "user_id": "user123"
}
```

#### 获取会话

```bash
GET /api/v1/sessions/{id}
```

#### 列出会话

```bash
GET /api/v1/sessions
```

---

## 3. Go SDK

### 3.1 初始化

```go
import "github.com/goviking/goviking/pkg/client"

// 创建客户端
c, err := client.NewClient("http://localhost:8080")
if err != nil {
    log.Fatal(err)
}
```

### 3.2 上下文操作

```go
// 创建上下文
ctx := &client.Context{
    URI:     "viking://agent/skills/search",
    Type:    "skill",
    Name:    "Search Skill",
    Content: "...",
}
result, err := c.CreateContext(context.Background(), ctx)

// 获取上下文
result, err := c.GetContext(context.Background(), "context-id")

// 列出上下文
list, err := c.ListContexts(context.Background())

// 删除上下文
err := c.DeleteContext(context.Background(), "context-id")
```

### 3.3 会话操作

```go
// 创建会话
session := &client.Session{
    UserID: "user123",
}
result, err := c.CreateSession(context.Background(), session)

// 获取会话
result, err := c.GetSession(context.Background(), "session-id")

// 列出会话
list, err := c.ListSessions(context.Background())
```

---

## 4. 配置

### 4.1 YAML 配置

```yaml
server:
  host: localhost
  port: 8080

storage:
  type: sqlite
  path: openviking.db
  in_memory: false

llm:
  provider: openai
  api_key: ${OPENAI_API_KEY}
  model: gpt-4
  base_url: https://api.openai.com/v1

retrieval:
  embedding_model: text-embedding-3-small
  similarity_threshold: 0.7
  max_results: 10
```

### 4.2 环境变量

```bash
# 可用环境变量
GOVIKING_SERVER_HOST
GOVIKING_SERVER_PORT
GOVIKING_STORAGE_TYPE
GOVIKING_STORAGE_PATH
GOVIKING_LLM_PROVIDER
GOVIKING_LLM_API_KEY
GOVIKING_LLM_MODEL
```

---

## 5. LLM 提供商

### 5.1 OpenAI

```go
import "github.com/goviking/goviking/pkg/llm"

provider, _ := llm.NewProvider(llm.Config{
    Type:    llm.ProviderOpenAI,
    APIKey:  "sk-...",
    Model:   "gpt-4",
})

resp, err := provider.Chat(ctx, &llm.ChatRequest{
    Model: "gpt-4",
    Messages: []llm.Message{
        {Role: llm.RoleUser, Content: "Hello"},
    },
})
```

### 5.2 Anthropic

```go
provider, _ := llm.NewProvider(llm.Config{
    Type:    llm.ProviderAnthropic,
    APIKey:  "sk-ant-...",
    Model:   "claude-3-opus-20240229",
})
```

---

## 6. CLI 命令

```bash
# 上下文
goviking context list
goviking context show <id>
goviking context create <path>

# 会话
goviking session list
goviking session show <id>
goviking session resume <id>

# 文件系统
goviking fs tree
goviking fs ls <path>

# 搜索
goviking search <query>

# 配置
goviking config show
goviking config init
```

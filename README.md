# GoViking

GoViking 是 [OpenViking](https://github.com/volcengine/OpenViking) (火山引擎开源的 AI Agent 上下文数据库) 的 Go 语言实现。

## 特性

- **文件系统式上下文管理 (AGFS)** - 像管理本地文件一样管理 Agent 的记忆、资源和技能
- **层级上下文加载 (L0/L1/L2)** - 三层结构，按需加载，节省 Token 成本
- **语义检索** - 支持向量嵌入和混合搜索
- **会话管理** - 自动提取长期记忆
- **多 LLM 提供商支持** - OpenAI、Anthropic 等
- **REST API** - 提供 HTTP 接口
- **单二进制部署** - 无需 Python 运行时

## 安装

```bash
go install ./cmd/goviking
```

## 快速开始

### 1. 初始化配置

```bash
goviking config init
```

### 2. 修改配置

编辑 `~/.goviking/config.yaml`:

```yaml
server:
  host: localhost
  port: 8080

storage:
  type: sqlite
  path: openviking.db

llm:
  provider: openai
  api_key: your-api-key
  model: gpt-4
```

### 3. 启动服务器

```bash
goviking server
```

### 4. 使用 CLI

```bash
# 列出上下文
goviking context list

# 搜索
goviking search "machine learning"

# 查看目录树
goviking fs tree
```

## 使用 Go SDK

```go
package main

import (
    "context"
    "fmt"

    "github.com/goviking/goviking/pkg/client"
)

func main() {
    c, _ := client.NewClient("http://localhost:8080")

    // 创建上下文
    ctx := &client.Context{
        URI:     "viking://agent/skills/search",
        Type:    "skill",
        Name:    "Search Skill",
        Content: "You are a search expert...",
    }
    result, _ := c.CreateContext(context.Background(), ctx)
    fmt.Printf("Created: %s\n", result.ID)
}
```

## 文档

- [架构文档](docs/architecture.md)
- [API 文档](docs/api.md)

## 项目结构

```
goviking/
├── cmd/               # CLI 入口
├── pkg/
│   ├── agfs/         # 文件系统上下文
│   ├── client/       # 客户端 SDK
│   ├── config/       # 配置管理
│   ├── core/         # 核心数据库
│   ├── llm/         # LLM 提供商
│   ├── message/     # 消息处理
│   ├── retrieval/     # 检索引擎
│   ├── server/      # HTTP 服务器
│   ├── service/     # 服务层
│   ├── session/     # 会话管理
│   ├── storage/     # 存储后端
│   └── utils/       # 工具函数
├── docs/            # 文档
└── go.mod
```

## 测试

```bash
go test ./...
```

## 许可证

Apache License 2.0

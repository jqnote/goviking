## Why

OpenViking is a powerful Context Database for AI Agents originally implemented in Python with a Rust CLI. Implementing it in Go would provide:
- **Single binary deployment** - no Python runtime dependencies
- **Better performance** for high-throughput scenarios
- **Easier integration** with existing Go-based AI infrastructure
- **Native gRPC support** for inter-service communication

The Go ecosystem lacks a mature "Context Database" solution for AI agents, making this a valuable port.

## What Changes

- Create a new Go implementation of the OpenViking core library
- Implement filesystem-based context management (AGFS)
- Build tiered context loading (L0/L1/L2)
- Develop directory recursive retrieval with semantic search
- Create session management with automatic memory extraction
- Implement storage backends (SQLite, etc.)
- Build CLI tool mirroring Python/Rust functionality
- Support multiple LLM providers (OpenAI, Anthropic, etc.)

## Capabilities

### New Capabilities

- **context-database**: Core context database functionality with tiered loading (L0/L1/L2)
- **filesystem-context**: Filesystem-like context management (AGFS) for unified memory/resources/skills
- **retrieval-engine**: Directory recursive retrieval with semantic vector search
- **session-manager**: Automatic session management with long-term memory extraction
- **llm-provider**: Multi-provider LLM integration (OpenAI, Anthropic, etc.)
- **storage-backend**: Pluggable storage backends (SQLite, file-based)
- **cli-tool**: Command-line interface for context management

### Modified Capabilities

(None - this is a new implementation)

## Impact

- New `goviking` Go module at `/Users/junqiang/aiwork/goviking/`
- Must maintain API compatibility with Python OpenViking where practical
- Dependencies: Go standard library, common Go AI libraries
- No impact on existing Python OpenViking project

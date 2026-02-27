## Context

**Background**: GoViking core modules are complete but lack configuration, client SDK, server, and utilities. This design covers implementing these remaining modules.

**Current State**:
- Core: storage, core, agfs, retrieval, session, llm, cli ✓
- Missing: config, client, message, server, service, utils

**Constraints**:
- Go 1.21+
- Keep dependencies minimal
- Match Python OpenViking APIs where practical
- Single binary deployment

## Goals / Non-Goals

**Goals:**
1. Configuration management with YAML + env vars
2. Client SDK (sync + async)
3. Message handling
4. REST API server
5. Utility functions

**Non-Goals:**
- gRPC server (future)
- WebSocket support (future)
- Metrics/observability (future)

## Decisions

### 1. Configuration
**Decision**: Use Viper for config management
**Alternatives**: custom YAML parser, envconfig
**Rationale**: Viper provides YAML + env vars + defaults out of box

### 2. Client SDK
**Decision**: Separate sync and async clients
**Alternatives**: single client with options
**Rationale**: Clear separation matches Python API

### 3. Server
**Decision**: Use standard library HTTP + Gorilla Mux
**Alternatives**: Gin, Echo, fiber
**Rationale**: Minimal deps, familiar to Go users

### 4. Message Format
**Decision**: Match OpenAI message format
**Alternatives**: custom format
**Rationale**: Easy integration with existing LLM providers

## Risks / Trade-offs

### Risk: API compatibility with Python
→ **Mitigation**: Design Go-native first, add compatibility layer if needed

### Risk: Server complexity
→ **Mitigation**: Start simple, add features as needed

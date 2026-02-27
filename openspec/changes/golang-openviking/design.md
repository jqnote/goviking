## Context

**Background**: OpenViking is a Context Database for AI Agents originally written in Python with a Rust CLI. It provides:
- Filesystem-like context management (AGFS)
- Tiered context loading (L0/L1/L2)
- Directory recursive retrieval with semantic search
- Session management with automatic memory extraction
- Multi-provider LLM integration

**Current State**: The Python implementation is mature with all core features. The Go implementation will start fresh but maintain API compatibility where practical.

**Constraints**:
- Go 1.21+ required
- Must support Linux, macOS, Windows
- Target: single binary deployment
- Keep dependencies minimal

## Goals / Non-Goals

**Goals:**
1. Implement core context database with tiered loading (L0/L1/L2)
2. Build filesystem-like context management (AGFS)
3. Create retrieval engine with vector search support
4. Implement session management with memory extraction
5. Support multiple LLM providers
6. Build storage backends (SQLite primary)
7. Create CLI tool

**Non-Goals:**
- Full feature parity with Python implementation initially
- MCP (Model Context Protocol) server - future phase
- Web server / REST API - future phase
- Evaluation framework - not porting

## Decisions

### 1. Project Structure
**Decision**: Use Go modules with clear package boundaries
```
goviking/
  pkg/
    core/           # Core context database logic
    storage/        # Storage backends
    retrieval/     # Retrieval engine
    session/       # Session management
    llm/           # LLM provider integrations
    agfs/          # Filesystem context
  cmd/
    goviking-cli/  # CLI tool
```
**Rationale**: Clean separation allows incremental development and testing

### 2. Storage Backend
**Decision**: SQLite as primary, with interface for others
**Alternatives considered**:
- BadgerDB: Good but adds CGO dependency
- Bolt: Too basic, no queries
- PostgreSQL: Overkill for single instance
**Rationale**: SQLite is zero-config, embedded, and sufficient for context DB

### 3. Vector Search
**Decision**: Use chroma go client or pure Go alternatives (pgvector-like)
**Alternatives considered**:
- Chroma: External service required
- Qdrant: External service required
- Pure Go (vecto): Simpler, no external deps
**Rationale**: Start with simple embedding + cosine similarity; defer complex vector DB

### 4. LLM Provider Integration
**Decision**: Use OpenAI SDK Go, extend for other providers
**Alternatives considered**:
- LangChain Go: Too heavy
- Each provider SDK: More work but cleaner
**Rationale**: OpenAI-compatible APIs are standard; other providers can wrap same interface

### 5. Configuration
**Decision**: YAML-based config file with env var overrides
**Rationale**: Matches Python OpenViking, familiar to users

## Risks / Trade-offs

### Risk: Feature Parity
→ **Mitigation**: Focus on core features first; iterate based on user feedback

### Risk: Performance
→ **Mitigation**: Go is naturally fast; profile before optimizing

### Risk: Vector Search Simplicity
→ **Mitigation**: Start simple (in-memory), add external vector DB support later

### Risk: API Compatibility
→ **Mitigation**: Design Go-native APIs first; add compatibility layer if needed

### Risk: CLI Complexity
→ **Mitigation**: Use Cobra/Viper for CLI; mirror Python CLI commands

## Open Questions

1. Should we support gRPC for inter-service communication?
2. How to handle embedding model selection and configuration?
3. Should we implement a web server for visualization?

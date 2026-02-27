## 1. Project Setup

- [ ] 1.1 Initialize Go module (`go mod init github.com/goviking/goviking`)
- [ ] 1.2 Create project directory structure (pkg/core, pkg/storage, pkg/retrieval, pkg/session, pkg/llm, pkg/agfs, cmd/goviking-cli)
- [ ] 1.3 Add dependencies (Cobra, Viper, SQLite driver, OpenAI SDK)
- [ ] 1.4 Create basic Go files with package imports

## 2. Storage Backend

- [ ] 2.1 Define storage interface (Storage interface with CRUD methods)
- [ ] 2.2 Implement SQLite backend with database schema
- [ ] 2.3 Create data models (Context, Session, Memory, File)
- [ ] 2.4 Implement CRUD operations for all models
- [ ] 2.5 Add query support and filtering

## 3. Context Database (Core)

- [ ] 3.1 Define tiered context structures (L0, L1, L2)
- [ ] 3.2 Implement context builder for merging multiple sources
- [ ] 3.3 Implement context window management
- [ ] 3.4 Add context persistence and restoration
- [ ] 3.5 Implement automatic compression for large contexts

## 4. Filesystem Context (AGFS)

- [ ] 4.1 Define directory and file structures
- [ ] 4.2 Implement directory operations (create, list, delete)
- [ ] 4.3 Implement file operations (read, write, delete, move)
- [ ] 4.4 Build tree structure for navigation
- [ ] 4.5 Support memories, resources, and skills as file types

## 5. Retrieval Engine

- [ ] 5.1 Implement recursive directory traversal
- [ ] 5.2 Add semantic search with vector embeddings
- [ ] 5.3 Implement hybrid search (keyword + semantic)
- [ ] 5.4 Add path-based and pattern-based retrieval
- [ ] 5.5 Implement retrieval trajectory tracking

## 6. Session Manager

- [ ] 6.1 Implement session lifecycle (create, resume, close)
- [ ] 6.2 Add message and tool call tracking
- [ ] 6.3 Implement automatic memory extraction
- [ ] 6.4 Add context summarization for long sessions
- [ ] 6.5 Implement session state persistence

## 7. LLM Provider

- [ ] 7.1 Define provider abstraction interface
- [ ] 7.2 Implement OpenAI provider
- [ ] 7.3 Add Anthropic provider support
- [ ] 7.4 Implement streaming responses
- [ ] 7.5 Add embedding generation support

## 8. CLI Tool

- [ ] 8.1 Set up Cobra CLI framework
- [ ] 8.2 Implement context commands (list, show, create)
- [ ] 8.3 Implement session commands (list, show, resume)
- [ ] 8.4 Implement filesystem commands (tree, ls)
- [ ] 8.5 Implement search command
- [ ] 8.6 Add config commands (show, init)

## 9. Configuration

- [ ] 9.1 Define configuration structure
- [ ] 9.2 Implement YAML config file loading
- [ ] 9.3 Add environment variable overrides
- [ ] 9.4 Create default configuration generation

## 10. Integration & Testing

- [ ] 10.1 Integrate all modules
- [ ] 10.2 Add unit tests for core functionality
- [ ] 10.3 Test CLI commands
- [ ] 10.4 Build and verify single binary

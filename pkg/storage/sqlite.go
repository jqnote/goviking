// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements StorageInterface using SQLite.
type SQLiteStorage struct {
	db *sql.DB
	cfg Config
}

// NewSQLiteStorage creates a new SQLite storage instance.
func NewSQLiteStorage(cfg Config) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{
		db:  db,
		cfg: cfg,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates all necessary tables.
func (s *SQLiteStorage) initSchema() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS contexts (
			id TEXT PRIMARY KEY,
			uri TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			context_type TEXT,
			parent_uri TEXT,
			is_leaf INTEGER DEFAULT 1,
			name TEXT,
			description TEXT,
			tags TEXT,
			abstract TEXT,
			active_count INTEGER DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_contexts_uri ON contexts(uri)`,
		`CREATE INDEX IF NOT EXISTS idx_contexts_parent_uri ON contexts(parent_uri)`,
		`CREATE INDEX IF NOT EXISTS idx_contexts_type ON contexts(type)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			session_id TEXT UNIQUE NOT NULL,
			user_id TEXT,
			total_turns INTEGER DEFAULT 0,
			total_tokens INTEGER DEFAULT 0,
			compression_count INTEGER DEFAULT 0,
			contexts_used INTEGER DEFAULT 0,
			skills_used INTEGER DEFAULT 0,
			memoies_extracted INTEGER DEFAULT 0,
			summary TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,

		`CREATE TABLE IF NOT EXISTS session_messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			order_index INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_session_messages_session_id ON session_messages(session_id)`,

		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			user_id TEXT,
			content TEXT NOT NULL,
			importance REAL DEFAULT 0.0,
			tags TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_session_id ON memories(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_user_id ON memories(user_id)`,

		`CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			uri TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			size INTEGER DEFAULT 0,
			content_type TEXT,
			checksum TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_files_uri ON files(uri)`,

		`CREATE TABLE IF NOT EXISTS usage_records (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			uri TEXT NOT NULL,
			type TEXT NOT NULL,
			contribution REAL DEFAULT 0.0,
			input TEXT,
			output TEXT,
			success INTEGER DEFAULT 1,
			timestamp TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_records_session_id ON usage_records(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_records_uri ON usage_records(uri)`,

		`CREATE TABLE IF NOT EXISTS relations (
			id TEXT PRIMARY KEY,
			uris TEXT NOT NULL,
			reason TEXT,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_relations_uris ON relations(uris)`,
	}

	for _, schema := range schemas {
		if _, err := s.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	return nil
}

// Close closes the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Ping checks if the database is accessible.
func (s *SQLiteStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// timeToString converts time.Time to string for SQLite storage.
func timeToString(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

// parseTime parses a string to time.Time.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// Try RFC3339Nano first
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	// Try SQLite default format
	if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", s); err == nil {
		return t
	}
	// Fallback - return zero time
	return time.Time{}
}

// Transaction executes a function within a transaction.
func (s *SQLiteStorage) Transaction(ctx context.Context, fn func(tx interface{}) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}

// =============================================================================
// Context Operations
// =============================================================================

// CreateContext inserts a new context into the database.
func (s *SQLiteStorage) CreateContext(ctx context.Context, c *Context) error {
	query := `INSERT INTO contexts (id, uri, type, context_type, parent_uri, is_leaf, name, description, tags, abstract, active_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		c.ID, c.URI, c.Type, c.ContextType, c.ParentURI, c.IsLeaf, c.Name,
		c.Description, c.Tags, c.Abstract, c.ActiveCount, c.CreatedAt, c.UpdatedAt)
	return err
}

// GetContext retrieves a context by ID.
func (s *SQLiteStorage) GetContext(ctx context.Context, id string) (*Context, error) {
	query := `SELECT id, uri, type, context_type, parent_uri, is_leaf, name, description, tags, abstract, active_count, created_at, updated_at FROM contexts WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var c Context
	var isLeaf int
	var createdAt, updatedAt string
	err := row.Scan(&c.ID, &c.URI, &c.Type, &c.ContextType, &c.ParentURI, &isLeaf,
		&c.Name, &c.Description, &c.Tags, &c.Abstract, &c.ActiveCount, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.IsLeaf = isLeaf == 1
	c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
	c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
	return &c, nil
}

// UpdateContext updates an existing context.
func (s *SQLiteStorage) UpdateContext(ctx context.Context, c *Context) error {
	query := `UPDATE contexts SET uri = ?, type = ?, context_type = ?, parent_uri = ?, is_leaf = ?, name = ?, description = ?, tags = ?, abstract = ?, active_count = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query,
		c.URI, c.Type, c.ContextType, c.ParentURI, c.IsLeaf, c.Name,
		c.Description, c.Tags, c.Abstract, c.ActiveCount, c.UpdatedAt, c.ID)
	return err
}

// DeleteContext deletes a context by ID.
func (s *SQLiteStorage) DeleteContext(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM contexts WHERE id = ?", id)
	return err
}

// QueryContexts queries contexts with filter options.
func (s *SQLiteStorage) QueryContexts(ctx context.Context, opts QueryOptions) ([]Context, error) {
	query := "SELECT id, uri, type, context_type, parent_uri, is_leaf, name, description, tags, abstract, active_count, created_at, updated_at FROM contexts"
	args := []interface{}{}

	if opts.Filter != nil && len(opts.Filter.Conds) > 0 {
		whereClause, filterArgs := buildFilterClause(opts.Filter)
		query += " WHERE " + whereClause
		args = append(args, filterArgs...)
	}

	if opts.OrderBy != "" {
		orderDir := "ASC"
		if opts.OrderDesc {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contexts []Context
	for rows.Next() {
		var c Context
		var isLeaf int
		var createdAt, updatedAt string
		err := rows.Scan(&c.ID, &c.URI, &c.Type, &c.ContextType, &c.ParentURI, &isLeaf,
			&c.Name, &c.Description, &c.Tags, &c.Abstract, &c.ActiveCount, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		c.IsLeaf = isLeaf == 1
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
		contexts = append(contexts, c)
	}

	return contexts, rows.Err()
}

// =============================================================================
// Session Operations
// =============================================================================

// CreateSession inserts a new session into the database.
func (s *SQLiteStorage) CreateSession(ctx context.Context, session *Session) error {
	query := `INSERT INTO sessions (id, session_id, user_id, total_turns, total_tokens, compression_count, contexts_used, skills_used, memoies_extracted, summary, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		session.ID, session.SessionID, session.UserID, session.TotalTurns, session.TotalTokens,
		session.CompressionCount, session.ContextsUsed, session.SkillsUsed,
		session.MemoriesExtracted, session.Summary, session.CreatedAt, session.UpdatedAt)
	return err
}

// GetSession retrieves a session by ID.
func (s *SQLiteStorage) GetSession(ctx context.Context, id string) (*Session, error) {
	query := `SELECT id, session_id, user_id, total_turns, total_tokens, compression_count, contexts_used, skills_used, memoies_extracted, summary, created_at, updated_at FROM sessions WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var session Session
	var createdAt, updatedAt string
	err := row.Scan(&session.ID, &session.SessionID, &session.UserID, &session.TotalTurns,
		&session.TotalTokens, &session.CompressionCount, &session.ContextsUsed, &session.SkillsUsed,
		&session.MemoriesExtracted, &session.Summary, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	session.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
	session.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
	return &session, nil
}

// UpdateSession updates an existing session.
func (s *SQLiteStorage) UpdateSession(ctx context.Context, session *Session) error {
	query := `UPDATE sessions SET session_id = ?, user_id = ?, total_turns = ?, total_tokens = ?, compression_count = ?, contexts_used = ?, skills_used = ?, memoies_extracted = ?, summary = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query,
		session.SessionID, session.UserID, session.TotalTurns, session.TotalTokens,
		session.CompressionCount, session.ContextsUsed, session.SkillsUsed,
		session.MemoriesExtracted, session.Summary, session.UpdatedAt, session.ID)
	return err
}

// DeleteSession deletes a session by ID.
func (s *SQLiteStorage) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	return err
}

// QuerySessions queries sessions with filter options.
func (s *SQLiteStorage) QuerySessions(ctx context.Context, opts QueryOptions) ([]Session, error) {
	query := "SELECT id, session_id, user_id, total_turns, total_tokens, compression_count, contexts_used, skills_used, memoies_extracted, summary, created_at, updated_at FROM sessions"
	args := []interface{}{}

	if opts.Filter != nil && len(opts.Filter.Conds) > 0 {
		whereClause, filterArgs := buildFilterClause(opts.Filter)
		query += " WHERE " + whereClause
		args = append(args, filterArgs...)
	}

	if opts.OrderBy != "" {
		orderDir := "ASC"
		if opts.OrderDesc {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		var createdAt, updatedAt string
		err := rows.Scan(&session.ID, &session.SessionID, &session.UserID, &session.TotalTurns,
			&session.TotalTokens, &session.CompressionCount, &session.ContextsUsed, &session.SkillsUsed,
			&session.MemoriesExtracted, &session.Summary, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		session.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		session.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// =============================================================================
// SessionMessage Operations
// =============================================================================

// CreateSessionMessage inserts a new session message.
func (s *SQLiteStorage) CreateSessionMessage(ctx context.Context, msg *SessionMessage) error {
	query := `INSERT INTO session_messages (id, session_id, role, content, order_index, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		msg.ID, msg.SessionID, msg.Role, msg.Content, msg.OrderIndex, msg.CreatedAt)
	return err
}

// GetSessionMessages retrieves all messages for a session.
func (s *SQLiteStorage) GetSessionMessages(ctx context.Context, sessionID string) ([]SessionMessage, error) {
	query := `SELECT id, session_id, role, content, order_index, created_at FROM session_messages WHERE session_id = ? ORDER BY order_index`
	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []SessionMessage
	for rows.Next() {
		var msg SessionMessage
		var createdAt string
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.OrderIndex, &createdAt)
		if err != nil {
			return nil, err
		}
		msg.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// DeleteSessionMessages deletes all messages for a session.
func (s *SQLiteStorage) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM session_messages WHERE session_id = ?", sessionID)
	return err
}

// =============================================================================
// Memory Operations
// =============================================================================

// CreateMemory inserts a new memory.
func (s *SQLiteStorage) CreateMemory(ctx context.Context, memory *Memory) error {
	query := `INSERT INTO memories (id, session_id, user_id, content, importance, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		memory.ID, memory.SessionID, memory.UserID, memory.Content, memory.Importance,
		memory.Tags, memory.CreatedAt, memory.UpdatedAt)
	return err
}

// GetMemory retrieves a memory by ID.
func (s *SQLiteStorage) GetMemory(ctx context.Context, id string) (*Memory, error) {
	query := `SELECT id, session_id, user_id, content, importance, tags, created_at, updated_at FROM memories WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var memory Memory
	var createdAt, updatedAt string
	err := row.Scan(&memory.ID, &memory.SessionID, &memory.UserID, &memory.Content,
		&memory.Importance, &memory.Tags, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	memory.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
	memory.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
	return &memory, nil
}

// UpdateMemory updates an existing memory.
func (s *SQLiteStorage) UpdateMemory(ctx context.Context, memory *Memory) error {
	query := `UPDATE memories SET session_id = ?, user_id = ?, content = ?, importance = ?, tags = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query,
		memory.SessionID, memory.UserID, memory.Content, memory.Importance,
		memory.Tags, memory.UpdatedAt, memory.ID)
	return err
}

// DeleteMemory deletes a memory by ID.
func (s *SQLiteStorage) DeleteMemory(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	return err
}

// QueryMemories queries memories with filter options.
func (s *SQLiteStorage) QueryMemories(ctx context.Context, opts QueryOptions) ([]Memory, error) {
	query := "SELECT id, session_id, user_id, content, importance, tags, created_at, updated_at FROM memories"
	args := []interface{}{}

	if opts.Filter != nil && len(opts.Filter.Conds) > 0 {
		whereClause, filterArgs := buildFilterClause(opts.Filter)
		query += " WHERE " + whereClause
		args = append(args, filterArgs...)
	}

	if opts.OrderBy != "" {
		orderDir := "ASC"
		if opts.OrderDesc {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var memory Memory
		var createdAt, updatedAt string
		err := rows.Scan(&memory.ID, &memory.SessionID, &memory.UserID, &memory.Content,
			&memory.Importance, &memory.Tags, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		memory.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		memory.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
		memories = append(memories, memory)
	}

	return memories, rows.Err()
}

// =============================================================================
// File Operations
// =============================================================================

// CreateFile inserts a new file.
func (s *SQLiteStorage) CreateFile(ctx context.Context, file *File) error {
	query := `INSERT INTO files (id, uri, name, size, content_type, checksum, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		file.ID, file.URI, file.Name, file.Size, file.ContentType,
		file.Checksum, file.CreatedAt, file.UpdatedAt)
	return err
}

// GetFile retrieves a file by ID.
func (s *SQLiteStorage) GetFile(ctx context.Context, id string) (*File, error) {
	query := `SELECT id, uri, name, size, content_type, checksum, created_at, updated_at FROM files WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var file File
	var createdAt, updatedAt string
	err := row.Scan(&file.ID, &file.URI, &file.Name, &file.Size, &file.ContentType,
		&file.Checksum, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	file.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
	file.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
	return &file, nil
}

// UpdateFile updates an existing file.
func (s *SQLiteStorage) UpdateFile(ctx context.Context, file *File) error {
	query := `UPDATE files SET uri = ?, name = ?, size = ?, content_type = ?, checksum = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query,
		file.URI, file.Name, file.Size, file.ContentType, file.Checksum, file.UpdatedAt, file.ID)
	return err
}

// DeleteFile deletes a file by ID.
func (s *SQLiteStorage) DeleteFile(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM files WHERE id = ?", id)
	return err
}

// QueryFiles queries files with filter options.
func (s *SQLiteStorage) QueryFiles(ctx context.Context, opts QueryOptions) ([]File, error) {
	query := "SELECT id, uri, name, size, content_type, checksum, created_at, updated_at FROM files"
	args := []interface{}{}

	if opts.Filter != nil && len(opts.Filter.Conds) > 0 {
		whereClause, filterArgs := buildFilterClause(opts.Filter)
		query += " WHERE " + whereClause
		args = append(args, filterArgs...)
	}

	if opts.OrderBy != "" {
		orderDir := "ASC"
		if opts.OrderDesc {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		var createdAt, updatedAt string
		err := rows.Scan(&file.ID, &file.URI, &file.Name, &file.Size, &file.ContentType,
			&file.Checksum, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		file.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		file.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", updatedAt)
		files = append(files, file)
	}

	return files, rows.Err()
}

// =============================================================================
// Usage Operations
// =============================================================================

// CreateUsage inserts a new usage record.
func (s *SQLiteStorage) CreateUsage(ctx context.Context, usage *Usage) error {
	query := `INSERT INTO usage_records (id, session_id, uri, type, contribution, input, output, success, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		usage.ID, usage.SessionID, usage.URI, usage.Type, usage.Contribution,
		usage.Input, usage.Output, usage.Success, usage.Timestamp)
	return err
}

// QueryUsage queries usage records with filter options.
func (s *SQLiteStorage) QueryUsage(ctx context.Context, opts QueryOptions) ([]Usage, error) {
	query := "SELECT id, session_id, uri, type, contribution, input, output, success, timestamp FROM usage_records"
	args := []interface{}{}

	if opts.Filter != nil && len(opts.Filter.Conds) > 0 {
		whereClause, filterArgs := buildFilterClause(opts.Filter)
		query += " WHERE " + whereClause
		args = append(args, filterArgs...)
	}

	if opts.OrderBy != "" {
		orderDir := "ASC"
		if opts.OrderDesc {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []Usage
	for rows.Next() {
		var usage Usage
		var success int
		var timestamp string
		err := rows.Scan(&usage.ID, &usage.SessionID, &usage.URI, &usage.Type,
			&usage.Contribution, &usage.Input, &usage.Output, &success, &timestamp)
		if err != nil {
			return nil, err
		}
		usage.Success = success == 1
		usage.Timestamp, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", timestamp)
		usages = append(usages, usage)
	}

	return usages, rows.Err()
}

// =============================================================================
// Relation Operations
// =============================================================================

// CreateRelation inserts a new relation.
func (s *SQLiteStorage) CreateRelation(ctx context.Context, relation *RelationEntry) error {
	query := `INSERT INTO relations (id, uris, reason, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		relation.ID, relation.URIs, relation.Reason, relation.CreatedAt)
	return err
}

// QueryRelations retrieves relations for a URI.
func (s *SQLiteStorage) QueryRelations(ctx context.Context, uri string) ([]RelationEntry, error) {
	query := `SELECT id, uris, reason, created_at FROM relations WHERE uris LIKE ?`
	rows, err := s.db.QueryContext(ctx, query, "%"+uri+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []RelationEntry
	for rows.Next() {
		var relation RelationEntry
		var createdAt string
		err := rows.Scan(&relation.ID, &relation.URIs, &relation.Reason, &createdAt)
		if err != nil {
			return nil, err
		}
		relation.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", createdAt)
		relations = append(relations, relation)
	}

	return relations, rows.Err()
}

// DeleteRelation deletes a relation by ID.
func (s *SQLiteStorage) DeleteRelation(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM relations WHERE id = ?", id)
	return err
}

// =============================================================================
// Collection Management (for interface compatibility)
// =============================================================================

// CreateCollection creates a collection (no-op for SQLite, tables exist).
func (s *SQLiteStorage) CreateCollection(name string, schema map[string]interface{}) error {
	// SQLite uses tables instead of collections; already created in initSchema
	return nil
}

// DropCollection drops a collection (no-op for SQLite).
func (s *SQLiteStorage) DropCollection(name string) error {
	// SQLite tables are created on init; we don't drop them
	return nil
}

// CollectionExists checks if a collection exists.
func (s *SQLiteStorage) CollectionExists(name string) bool {
	switch name {
	case "contexts", "sessions", "memories", "files", "usage_records", "relations":
		return true
	default:
		return false
	}
}

// ListCollections lists all collections.
func (s *SQLiteStorage) ListCollections() ([]string, error) {
	return []string{"contexts", "sessions", "memories", "files", "usage_records", "relations"}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildFilterClause builds a SQL WHERE clause from filter conditions.
func buildFilterClause(filter *Filter) (string, []interface{}) {
	if filter == nil || len(filter.Conds) == 0 {
		return "", nil
	}

	var clauses []string
	var args []interface{}

	for _, cond := range filter.Conds {
		switch cond.Op {
		case "must":
			// Exact match
			clauses = append(clauses, fmt.Sprintf("%s = ?", cond.Field))
			args = append(args, cond.Value)
		case "range":
			// Range query
			if cond.GTE != nil {
				clauses = append(clauses, fmt.Sprintf("%s >= ?", cond.Field))
				args = append(args, cond.GTE)
			}
			if cond.GT != nil {
				clauses = append(clauses, fmt.Sprintf("%s > ?", cond.Field))
				args = append(args, cond.GT)
			}
			if cond.LTE != nil {
				clauses = append(clauses, fmt.Sprintf("%s <= ?", cond.Field))
				args = append(args, cond.LTE)
			}
			if cond.LT != nil {
				clauses = append(clauses, fmt.Sprintf("%s < ?", cond.Field))
				args = append(args, cond.LT)
			}
		case "prefix":
			// Prefix match (LIKE)
			clauses = append(clauses, fmt.Sprintf("%s LIKE ?", cond.Field))
			args = append(args, cond.prefix+"%")
		case "contains":
			// Contains substring
			clauses = append(clauses, fmt.Sprintf("%s LIKE ?", cond.Field))
			args = append(args, "%"+cond.substr+"%")
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}

	connector := " AND "
	if filter.Op == "or" {
		connector = " OR "
	}

	return strings.Join(clauses, connector), args
}

// Ensure SQLiteStorage implements StorageInterface
var _ StorageInterface = (*SQLiteStorage)(nil)

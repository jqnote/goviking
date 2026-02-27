// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package main is the CLI entry point for goviking.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jqnote/goviking/pkg/client"
	"github.com/jqnote/goviking/pkg/config"
	"github.com/jqnote/goviking/pkg/server"
)

var (
	// Version is the version of the CLI.
	Version = "0.1.0"
	// Commit is the git commit.
	Commit = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "goviking",
		Short: "GoViking - Context Database for AI Agents",
		Long: `GoViking is a Context Database for AI Agents.

It provides filesystem-like context management, tiered context loading,
semantic search, and automatic memory extraction.`,
		Version: Version,
	}

	// Add subcommands
	rootCmd.AddCommand(contextCmd())
	rootCmd.AddCommand(sessionCmd())
	rootCmd.AddCommand(fsCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getClient() (*client.Client, error) {
	// Get server address from config
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	serverAddr := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	return client.NewClient(serverAddr)
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GoViking CLI %s (commit: %s)\n", Version, Commit)
		},
	}
}

func contextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage context entries",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all contexts",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			contexts, err := c.ListContexts(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(contexts) == 0 {
				fmt.Println("No contexts found.")
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tNAME\tTYPE\tURI\n")
			for _, ctx := range contexts {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ctx.ID, ctx.Name, ctx.Type, ctx.URI)
			}
			w.Flush()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show a context by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			context, err := c.GetContext(ctx, args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			data, err := json.MarshalIndent(context, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create [path]",
		Short: "Create a new context from a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				os.Exit(1)
			}

			// Get filename as name
			name := path
			if idx := len(path) - 1; idx > 0 {
				for i := len(path) - 1; i >= 0; i-- {
					if path[i] == '/' || path[i] == '\\' {
						name = path[i+1:]
						break
					}
				}
			}

			// Determine type from extension
			typ := "file"
			if len(path) > 3 {
				switch path[len(path)-3:] {
				case ".md":
					typ = "document"
				case ".py":
					typ = "code"
				case ".go":
					typ = "code"
				case ".js":
					typ = "code"
				case ".ts":
					typ = "code"
				}
			}

			ctx := &client.Context{
				URI:     fmt.Sprintf("viking://local/%s", name),
				Type:    typ,
				Name:    name,
				Content: string(content),
			}

			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			result, err := c.CreateContext(context.Background(), ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created context: %s\n", result.ID)
		},
	})

	return cmd
}

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage sessions",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tUSER_ID\tSTATE\tCREATED_AT\n")
			for _, s := range sessions {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.UserID, s.State, s.CreatedAt.Format("2006-01-02 15:04"))
			}
			w.Flush()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show a session by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			session, err := c.GetSession(ctx, args[0], true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			data, err := json.MarshalIndent(session, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "resume [id]",
		Short: "Resume a session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Resume is essentially the same as show - it shows the session state
			// In a full implementation, this would update the session state to "active"
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			session, err := c.GetSession(ctx, args[0], true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Update state to active
			session.State = "active"

			data, err := json.MarshalIndent(session, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Session resumed:\n%s\n", string(data))
		},
	})

	return cmd
}

func fsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Filesystem operations",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "tree",
		Short: "Show directory tree",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			contexts, err := c.ListContexts(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(contexts) == 0 {
				fmt.Println("No contexts found.")
				return
			}

			// Group contexts by type (simulating directory structure)
			typeGroups := make(map[string][]client.Context)
			for _, ctx := range contexts {
				typeGroups[ctx.Type] = append(typeGroups[ctx.Type], ctx)
			}

			// Print tree
			fmt.Println("viking://")
			for t, ctxs := range typeGroups {
				fmt.Printf("├── %s/\n", t)
				for i, c := range ctxs {
					prefix := "│   "
					if i == len(ctxs)-1 {
						prefix = "    "
					}
					fmt.Printf("%s└── %s (%s)\n", prefix, c.Name, c.ID[:8])
				}
			}
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls [path]",
		Short: "List files in a directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			contexts, err := c.ListContexts(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Filter contexts by path prefix
			var filtered []client.Context
			for _, c := range contexts {
				if len(c.URI) >= len(path) && c.URI[:len(path)] == path {
					filtered = append(filtered, c)
				}
			}

			if len(filtered) == 0 {
				// Show all if no matches
				filtered = contexts
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tID\n")
			for _, c := range filtered {
				fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, c.Type, c.ID)
			}
			w.Flush()
		},
	})

	return cmd
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search contexts",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]

			c, err := getClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			contexts, err := c.ListContexts(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Simple text search (in a real implementation, this would use semantic search)
			var results []client.Context
			lowerQuery := toLower(query)
			for _, c := range contexts {
				if contains(toLower(c.Name), lowerQuery) ||
					contains(toLower(c.Content), lowerQuery) ||
					contains(toLower(c.URI), lowerQuery) {
					results = append(results, c)
				}
			}

			if len(results) == 0 {
				fmt.Printf("No results found for: %s\n", query)
				return
			}

			fmt.Printf("Search results for: %s\n\n", query)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tTYPE\tID\n")
			for _, r := range results {
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Type, r.ID)
			}
			w.Flush()
		},
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func serverCmd() *cobra.Command {
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the GoViking HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			// Load config for server settings
			cfg, err := config.LoadDefault()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}

			// Override with command line flags if provided
			if host == "" {
				host = cfg.Server.Host
			}
			if port == 0 {
				port = cfg.Server.Port
			}

			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("Starting GoViking server at %s...\n", addr)

			s := server.New()
			s.SetAddr(addr)

			// Handle graceful shutdown
			go func() {
				if err := s.Start(addr); err != nil && err != http.ErrServerClosed {
					fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
					os.Exit(1)
				}
			}()

			// Wait for interrupt signal
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			fmt.Println("\nShutting down server...")
			ctx, cancel := context.WithTimeout(context.Background(), 5)
			defer cancel()
			if err := s.Shutdown(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v\n", err)
			}
			fmt.Println("Server exited")
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Server host (default from config)")
	cmd.Flags().IntVar(&port, "port", 0, "Server port (default from config)")

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadDefault()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Get default config path
			configPath := config.GetConfigPath()

			// Check if config already exists
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf("Config already exists at: %s\n", configPath)
				fmt.Println("Use 'goviking config show' to view it.")
				os.Exit(0)
			}

			// Create config directory if needed
			dir := configPath
			for i := len(configPath) - 1; i >= 0; i-- {
				if configPath[i] == '/' || configPath[i] == '\\' {
					dir = configPath[:i]
					break
				}
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
				os.Exit(1)
			}

			// Create default config
			defaultConfig := `server:
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
`

			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Configuration initialized at: %s\n", configPath)
			fmt.Println("\nPlease edit the config file to set your API key.")
		},
	})

	return cmd
}

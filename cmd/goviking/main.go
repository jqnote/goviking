// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package main is the CLI entry point for goviking.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
			fmt.Println("Listing contexts...")
			// TODO: Implement context list
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show a context by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Showing context: %s\n", args[0])
			// TODO: Implement context show
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create [path]",
		Short: "Create a new context",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Creating context: %s\n", args[0])
			// TODO: Implement context create
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
			fmt.Println("Listing sessions...")
			// TODO: Implement session list
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show [id]",
		Short: "Show a session by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Showing session: %s\n", args[0])
			// TODO: Implement session show
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "resume [id]",
		Short: "Resume a session",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Resuming session: %s\n", args[0])
			// TODO: Implement session resume
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
			fmt.Println("Showing directory tree...")
			// TODO: Implement fs tree
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls [path]",
		Short: "List files in a directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Listing: %s\n", args[0])
			// TODO: Implement fs ls
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
			fmt.Printf("Searching for: %s\n", args[0])
			// TODO: Implement search
			ctx := context.Background()
			_ = ctx
		},
	}
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
			fmt.Println("Showing configuration...")
			// TODO: Implement config show
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initializing configuration...")
			// TODO: Implement config init
		},
	})

	return cmd
}

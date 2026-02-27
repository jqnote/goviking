// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

// Package server provides HTTP server for GoViking.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Server is the GoViking HTTP server.
type Server struct {
	router   *mux.Router
	server   *http.Server
}

// New creates a new server.
func New() *Server {
	r := mux.NewRouter()
	s := &Server{
		router: r,
		server: &http.Server{
			Handler: r,
			Addr:    ":8080",
		},
	}
	s.setupRoutes()
	return s
}

// SetAddr sets the server address.
func (s *Server) SetAddr(addr string) {
	s.server.Addr = addr
}

// setupRoutes sets up the HTTP routes.
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Context routes
	s.router.HandleFunc("/api/v1/contexts", s.handleListContexts).Methods("GET")
	s.router.HandleFunc("/api/v1/contexts", s.handleCreateContext).Methods("POST")
	s.router.HandleFunc("/api/v1/contexts/{id}", s.handleGetContext).Methods("GET")
	s.router.HandleFunc("/api/v1/contexts/{id}", s.handleDeleteContext).Methods("DELETE")

	// Session routes
	s.router.HandleFunc("/api/v1/sessions", s.handleListSessions).Methods("GET")
	s.router.HandleFunc("/api/v1/sessions", s.handleCreateSession).Methods("POST")
	s.router.HandleFunc("/api/v1/sessions/{id}", s.handleGetSession).Methods("GET")

	// FS routes
	s.router.HandleFunc("/api/v1/fs/list", s.handleFSList).Methods("GET")
	s.router.HandleFunc("/api/v1/fs/mkdir", s.handleFSMkdir).Methods("POST")
	s.router.HandleFunc("/api/v1/fs/read", s.handleFSRead).Methods("GET")
	s.router.HandleFunc("/api/v1/fs/write", s.handleFSWrite).Methods("POST")
	s.router.HandleFunc("/api/v1/fs/delete", s.handleFSDelete).Methods("DELETE")
	s.router.HandleFunc("/api/v1/fs/move", s.handleFSMove).Methods("POST")
	s.router.HandleFunc("/api/v1/fs/tree", s.handleFSTree).Methods("GET")
}

// Start starts the server.
func (s *Server) Start(addr string) error {
	if addr != "" {
		s.server.Addr = addr
	}
	return s.server.ListenAndServe()
}

// StartTLS starts the server with TLS.
func (s *Server) StartTLS(addr, certFile, keyFile string) error {
	if addr != "" {
		s.server.Addr = addr
	}
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown shuts down the server gracefully.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Context handlers
func (s *Server) handleListContexts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleCreateContext(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (s *Server) handleGetContext(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   id,
		"uri":  fmt.Sprintf("viking://context/%s", id),
		"name": "sample",
	})
}

func (s *Server) handleDeleteContext(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Session handlers
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"session_id": id,
		"state":      "active",
	})
}

// FS handlers
func (s *Server) handleFSList(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// Use service.FSService if available, otherwise return empty
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleFSMkdir(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": req.Path,
	})
}

func (s *Server) handleFSRead(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path":    path,
		"content": "",
	})
}

func (s *Server) handleFSWrite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": req.Path,
	})
}

func (s *Server) handleFSDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": path,
	})
}

func (s *Server) handleFSMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"from": req.From,
		"to":   req.To,
	})
}

func (s *Server) handleFSTree(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": path,
		"tree": "",
	})
}

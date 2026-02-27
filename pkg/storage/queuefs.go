// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrQueueNotFound is returned when queue is not found.
	ErrQueueNotFound = errors.New("queue not found")
	// ErrMessageNotFound is returned when message is not found.
	ErrMessageNotFound = errors.New("message not found")
	// ErrDependencyNotMet is returned when dependency is not met.
	ErrDependencyNotMet = errors.New("dependency not met")
)

// Message represents a queue message.
type Message struct {
	ID         string         `json:"id"`
	Queue      string         `json:"queue"`
	Content    string         `json:"content"`
	Payload    map[string]any `json:"payload"`
	Dependencies []string      `json:"dependencies"`
	Status     MessageStatus  `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
}

// MessageStatus represents the status of a message.
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusProcessing MessageStatus = "processing"
	MessageStatusCompleted  MessageStatus = "completed"
	MessageStatusFailed    MessageStatus = "failed"
)

// QueueManager manages message queues.
type QueueManager struct {
	queues    map[string]*Queue
	mu        sync.RWMutex
	handlers  map[string]MessageHandler
	processor *MessageProcessor
}

// Queue represents a message queue.
type Queue struct {
	Name        string
	Messages    []*Message
	MaxSize     int
	CreatedAt   time.Time
}

// MessageHandler handles messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// NewQueueManager creates a new queue manager.
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues:   make(map[string]*Queue),
		handlers: make(map[string]MessageHandler),
	}
}

// CreateQueue creates a new queue.
func (qm *QueueManager) CreateQueue(ctx context.Context, name string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, exists := qm.queues[name]; exists {
		return fmt.Errorf("queue %s already exists", name)
	}

	qm.queues[name] = &Queue{
		Name:      name,
		Messages:  make([]*Message, 0),
		MaxSize:   1000, // Default max size
		CreatedAt: time.Now(),
	}

	return nil
}

// Enqueue adds a message to a queue.
func (qm *QueueManager) Enqueue(ctx context.Context, queue string, msg *Message) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	q, ok := qm.queues[queue]
	if !ok {
		return ErrQueueNotFound
	}

	msg.ID = uuid.New().String()
	msg.Queue = queue
	msg.Status = MessageStatusPending
	msg.CreatedAt = time.Now()

	// Check dependencies
	if len(msg.Dependencies) > 0 {
		if !qm.dependenciesMet(ctx, msg.Dependencies) {
			return ErrDependencyNotMet
		}
	}

	q.Messages = append(q.Messages, msg)
	return nil
}

// dependenciesMet checks if all dependencies are met.
func (qm *QueueManager) dependenciesMet(ctx context.Context, deps []string) bool {
	for _, depID := range deps {
		found := false
		for _, q := range qm.queues {
			for _, msg := range q.Messages {
				if msg.ID == depID && msg.Status == MessageStatusCompleted {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Dequeue removes and returns a message from the queue.
func (qm *QueueManager) Dequeue(ctx context.Context, queue string) (*Message, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	q, ok := qm.queues[queue]
	if !ok {
		return nil, ErrQueueNotFound
	}

	// Find next ready message
	for i, msg := range q.Messages {
		if msg.Status == MessageStatusPending {
			if len(msg.Dependencies) == 0 || qm.dependenciesMet(ctx, msg.Dependencies) {
				msg.Status = MessageStatusProcessing
				// Remove from queue
				q.Messages = append(q.Messages[:i], q.Messages[i+1:]...)
				return msg, nil
			}
		}
	}

	return nil, nil // No ready messages
}

// Complete marks a message as completed.
func (qm *QueueManager) Complete(ctx context.Context, msgID string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, q := range qm.queues {
		for _, msg := range q.Messages {
			if msg.ID == msgID {
				now := time.Now()
				msg.Status = MessageStatusCompleted
				msg.ProcessedAt = &now
				return nil
			}
		}
	}

	return ErrMessageNotFound
}

// Fail marks a message as failed.
func (qm *QueueManager) Fail(ctx context.Context, msgID string) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	for _, q := range qm.queues {
		for _, msg := range q.Messages {
			if msg.ID == msgID {
				msg.Status = MessageStatusFailed
				return nil
			}
		}
	}

	return ErrMessageNotFound
}

// GetQueueSize returns the size of a queue.
func (qm *QueueManager) GetQueueSize(ctx context.Context, queue string) (int, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	q, ok := qm.queues[queue]
	if !ok {
		return 0, ErrQueueNotFound
	}

	return len(q.Messages), nil
}

// ListQueues lists all queues.
func (qm *QueueManager) ListQueues(ctx context.Context) ([]string, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	var names []string
	for name := range qm.queues {
		names = append(names, name)
	}

	return names, nil
}

// MessageProcessor processes messages concurrently.
type MessageProcessor struct {
	queueManager *QueueManager
	concurrency  int
	handlers     map[string]MessageHandler
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewMessageProcessor creates a new message processor.
func NewMessageProcessor(qm *QueueManager, concurrency int) *MessageProcessor {
	return &MessageProcessor{
		queueManager: qm,
		concurrency:  concurrency,
		handlers:     make(map[string]MessageHandler),
		stopCh:      make(chan struct{}),
	}
}

// RegisterHandler registers a message handler for a queue.
func (mp *MessageProcessor) RegisterHandler(queue string, handler MessageHandler) {
	mp.handlers[queue] = handler
}

// Start starts processing messages.
func (mp *MessageProcessor) Start(ctx context.Context) {
	for i := 0; i < mp.concurrency; i++ {
		mp.wg.Add(1)
		go mp.processLoop(ctx, i)
	}
}

// Stop stops processing messages.
func (mp *MessageProcessor) Stop() {
	close(mp.stopCh)
	mp.wg.Wait()
}

func (mp *MessageProcessor) processLoop(ctx context.Context, workerID int) {
	defer mp.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-mp.stopCh:
			return
		case <-ticker.C:
			// Process each queue
			queues, _ := mp.queueManager.ListQueues(ctx)
			for _, queue := range queues {
				handler, ok := mp.handlers[queue]
				if !ok {
					continue
				}

				msg, err := mp.queueManager.Dequeue(ctx, queue)
				if err != nil || msg == nil {
					continue
				}

				// Process message
				if err := handler(ctx, msg); err != nil {
					mp.queueManager.Fail(ctx, msg.ID)
				} else {
					mp.queueManager.Complete(ctx, msg.ID)
				}
			}
		}
	}
}

// Copyright (c) 2026 Beijing Volcano Engine Technology Co., Ltd.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jqnote/goviking/pkg/llm"
)

// Memory categories (new 6-category system)
const (
	CategoryProfile   Category = "profile"   // User profile information
	CategoryPreference Category = "preference" // User preferences
	CategoryEntity    Category = "entity"    // Entities mentioned
	CategoryEvent    Category = "event"    // Events occurred
	CategoryCase     Category = "case"     // Cases/scenarios
	CategoryPattern  Category = "pattern"  // Patterns detected
)

// Legacy categories (for backward compatibility)
const (
	CategoryFact     Category = "fact"     // Factual information
	CategorySkill    Category = "skill"    // Learned skills
	CategoryGoal     Category = "goal"     // Goals and objectives
	CategoryContext  Category = "context"  // Context information
	CategoryOther    Category = "other"    // Other information
)

// Category represents the category of extracted memory.
type Category string

// LLMExtractor uses LLM to extract memories from session messages.
type LLMExtractor struct {
	client     llm.Provider
	config     ExtractorConfig
	promptTemplate string
}

// NewLLMExtractor creates a new LLM-based memory extractor.
func NewLLMExtractor(client llm.Provider, config ExtractorConfig) *LLMExtractor {
	if config.MinImportance == 0 {
		config.MinImportance = 0.5
	}
	if config.MaxMemories == 0 {
		config.MaxMemories = 10
	}

	return &LLMExtractor{
		client:  client,
		config:  config,
		promptTemplate: defaultMemoryExtractionPrompt,
	}
}

// Extract extracts memories from session messages using LLM.
func (e *LLMExtractor) Extract(ctx context.Context, messages []*Message) ([]*ExtractedMemory, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	// Format messages for the prompt
	content := e.formatMessages(messages)

	// Build the prompt
	prompt := fmt.Sprintf(e.promptTemplate, content)

	// Call LLM
	resp, err := e.client.Chat(ctx, &llm.ChatRequest{
		Model:       "",
		Temperature: 0.3,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a memory extraction assistant. Extract important information from the conversation and return a JSON array."},
			{Role: llm.RoleUser, Content: prompt},
		},
		MaxTokens: 2000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract memories: %w", err)
	}

	// Parse the response
	if len(resp.Choices) == 0 {
		return nil, nil
	}

	responseContent := resp.Choices[0].Message.Content

	// Try to parse JSON from response
	memories, err := e.parseMemoryResponse(responseContent)
	if err != nil {
		// Try to extract JSON from markdown code blocks
		memories, err = e.extractJSONFromMarkdown(responseContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse memory response: %w", err)
		}
	}

	// Filter by importance and limit
	var filtered []*ExtractedMemory
	for _, m := range memories {
		if m.Importance >= e.config.MinImportance {
			m.SessionID = e.config.SessionID
			m.CreatedAt = time.Now().UTC()
			filtered = append(filtered, m)
			if len(filtered) >= e.config.MaxMemories {
				break
			}
		}
	}

	return filtered, nil
}

// formatMessages formats messages for the prompt.
func (e *LLMExtractor) formatMessages(messages []*Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		roleStr := string(msg.Role)
		sb.WriteString(fmt.Sprintf("%s: %s\n", roleStr, msg.Content))
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("  Tool call: %s(%s)\n", tc.Function.Name, tc.Function.Arguments))
			}
		}
	}
	return sb.String()
}

// parseMemoryResponse parses the LLM response into memories.
func (e *LLMExtractor) parseMemoryResponse(response string) ([]*ExtractedMemory, error) {
	// Try direct JSON parse first
	var memories []*ExtractedMemory
	if err := json.Unmarshal([]byte(response), &memories); err == nil {
		return memories, nil
	}

	// Try to find JSON array in response
	lines := strings.Split(response, "\n")
	var jsonLines []string
	inArray := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inArray = true
		}
		if inArray {
			jsonLines = append(jsonLines, line)
		}
		if strings.HasSuffix(trimmed, "]") {
			break
		}
	}

	if len(jsonLines) > 0 {
		jsonStr := strings.Join(jsonLines, "\n")
		if err := json.Unmarshal([]byte(jsonStr), &memories); err == nil {
			return memories, nil
		}
	}

	return nil, fmt.Errorf("no valid JSON found in response")
}

// extractJSONFromMarkdown extracts JSON from markdown code blocks.
func (e *LLMExtractor) extractJSONFromMarkdown(response string) ([]*ExtractedMemory, error) {
	// Look for JSON in code blocks
	start := strings.Index(response, "```json")
	if start == -1 {
		start = strings.Index(response, "```")
	}
	if start != -1 {
		end := strings.Index(response[start+3:], "```")
		if end != -1 {
			jsonStr := response[start+7 : start+3+end]
			var memories []*ExtractedMemory
			if err := json.Unmarshal([]byte(jsonStr), &memories); err == nil {
				return memories, nil
			}
		}
	}

	// Last resort: try to find any JSON-like structure
	lines := strings.Split(response, "\n")
	var sb strings.Builder
	inBracket := false
	for _, line := range lines {
		if strings.Contains(line, "[") || strings.Contains(line, "{") {
			inBracket = true
		}
		if inBracket {
			sb.WriteString(line)
		}
		if strings.Contains(line, "]") || strings.Contains(line, "}") {
			break
		}
	}

	var memories []*ExtractedMemory
	if err := json.Unmarshal([]byte(sb.String()), &memories); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return memories, nil
}

const defaultMemoryExtractionPrompt = `Extract important information from the following conversation that should be remembered for future interactions.

For each piece of information, extract:
1. content: The actual information to remember
2. importance: A score from 0 to 1 indicating how important this is
3. category: One of: preference, fact, skill, goal, context, other

Conversation:
%s

Return a JSON array of memories. Example:
[
  {"content": "User prefers concise responses", "importance": 0.8, "category": "preference"},
  {"content": "User is interested in machine learning", "importance": 0.7, "category": "fact"}
]

Only return the JSON array, no other text.`

// CategoryWeights defines importance weights for each category.
var CategoryWeights = map[Category]float64{
	CategoryProfile:   0.9,  // User profile is highly important
	CategoryPreference: 0.8,  // Preferences are important
	CategoryEntity:    0.7,  // Entities are moderately important
	CategoryEvent:    0.6,  // Events are less important
	CategoryCase:     0.7,  // Cases are moderately important
	CategoryPattern:  0.5,  // Patterns are less critical
}

// GetCategoryImportance returns the base importance weight for a category.
func GetCategoryImportance(cat Category) float64 {
	if weight, ok := CategoryWeights[cat]; ok {
		return weight
	}
	return 0.5 // Default weight
}

// CategoryPrompts contains prompts for each memory category.
var CategoryPrompts = map[Category]string{
	CategoryProfile: `Extract user profile information from the conversation:
- Name, identity, role
- Professional background
- Skills and expertise
- Personal characteristics

Conversation:
%s

Return a JSON array with profile information.`,

	CategoryPreference: `Extract user preferences from the conversation:
- Communication style preferences
- Topic interests
- Working style preferences
- Tool and technology preferences

Conversation:
%s

Return a JSON array with preference information.`,

	CategoryEntity: `Extract entities mentioned in the conversation:
- People names
- Company/organization names
- Product names
- Project names
- Technical terms

Conversation:
%s

Return a JSON array with entity information.`,

	CategoryEvent: `Extract events that occurred in the conversation:
- Meetings or discussions
- Decisions made
- Actions taken
- Milestones reached

Conversation:
%s

Return a JSON array with event information.`,

	CategoryCase: `Extract cases or scenarios from the conversation:
- Problem descriptions
- Use cases
- Examples mentioned
- Situations described

Conversation:
%s

Return a JSON array with case information.`,

	CategoryPattern: `Extract patterns detected in the conversation:
- Behavioral patterns
- Communication patterns
- Common themes
- Recurring topics

Conversation:
%s

Return a JSON array with pattern information.`,
}

// ExtractByCategory extracts memories for a specific category.
func (e *LLMExtractor) ExtractByCategory(ctx context.Context, messages []*Message, category Category) ([]*ExtractedMemory, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	promptTemplate, ok := CategoryPrompts[category]
	if !ok {
		// Fall back to default prompt
		promptTemplate = defaultMemoryExtractionPrompt
	}

	// Format messages for the prompt
	content := e.formatMessages(messages)
	prompt := fmt.Sprintf(promptTemplate, content)

	// Call LLM
	resp, err := e.client.Chat(ctx, &llm.ChatRequest{
		Model:       "",
		Temperature: 0.3,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a memory extraction assistant. Extract important information and return a JSON array."},
			{Role: llm.RoleUser, Content: prompt},
		},
		MaxTokens: 2000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract %s memories: %w", category, err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil
	}

	responseContent := resp.Choices[0].Message.Content
	memories, err := e.parseMemoryResponse(responseContent)
	if err != nil {
		memories, err = e.extractJSONFromMarkdown(responseContent)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s memory response: %w", category, err)
		}
	}

	// Apply category-specific importance weighting
	baseWeight := GetCategoryImportance(category)
	var filtered []*ExtractedMemory
	for _, m := range memories {
		m.Category = string(category)
		m.Importance = m.Importance * baseWeight // Apply category weight
		if m.Importance >= e.config.MinImportance {
			m.SessionID = e.config.SessionID
			m.CreatedAt = time.Now().UTC()
			filtered = append(filtered, m)
			if len(filtered) >= e.config.MaxMemories {
				break
			}
		}
	}

	return filtered, nil
}

// ExtractAllCategories extracts memories from all categories.
func (e *LLMExtractor) ExtractAllCategories(ctx context.Context, messages []*Message) (map[Category][]*ExtractedMemory, error) {
	results := make(map[Category][]*ExtractedMemory)

	// Extract from each category
	for _, cat := range []Category{CategoryProfile, CategoryPreference, CategoryEntity, CategoryEvent, CategoryCase, CategoryPattern} {
		memories, err := e.ExtractByCategory(ctx, messages, cat)
		if err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", cat, err)
		}
		if len(memories) > 0 {
			results[cat] = memories
		}
	}

	return results, nil
}

// LLMSummarizer uses LLM to create summaries of session content.
type LLMSummarizer struct {
	client llm.Provider
	config SummarizerConfig
}

// NewLLMSummarizer creates a new LLM-based summarizer.
func NewLLMSummarizer(client llm.Provider, config SummarizerConfig) *LLMSummarizer {
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}
	if config.KeepRecentMsgs == 0 {
		config.KeepRecentMsgs = 5
	}
	return &LLMSummarizer{
		client: client,
		config: config,
	}
}

// Summarize creates a summary of messages.
func (s *LLMSummarizer) Summarize(ctx context.Context, messages []*Message) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	content := formatMessagesForSummary(messages)

	prompt := fmt.Sprintf(`Summarize the following conversation concisely, capturing the key points and any important information:

%s

Provide a brief summary (2-3 sentences):`, content)

	resp, err := s.client.Chat(ctx, &llm.ChatRequest{
		Model:       "",
		Temperature: 0.3,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are a conversation summarization assistant."},
			{Role: llm.RoleUser, Content: prompt},
		},
		MaxTokens: s.config.MaxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("failed to summarize: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}

// Compress compresses messages into a summary while keeping recent messages.
func (s *LLMSummarizer) Compress(ctx context.Context, messages []*Message, maxTokens int) (string, int64, error) {
	if len(messages) == 0 {
		return "", 0, nil
	}

	// Keep recent messages unchanged
	recentCount := s.config.KeepRecentMsgs
	if recentCount > len(messages) {
		recentCount = len(messages)
	}

	// Keep recent messages as-is
	_ = messages[len(messages)-recentCount:]
	olderMsgs := messages[:len(messages)-recentCount]

	// Estimate tokens (rough estimate: 1 token ≈ 4 characters)
	estimatedTokens := int64(len(formatMessagesForSummary(olderMsgs)) / 4)

	// If already under limit, no compression needed
	if estimatedTokens <= int64(maxTokens) {
		return formatMessagesForSummary(olderMsgs), 0, nil
	}

	// Summarize older messages
	summary, err := s.Summarize(ctx, olderMsgs)
	if err != nil {
		return "", 0, err
	}

	// Calculate tokens saved
	tokensSaved := estimatedTokens - int64(len(summary)/4)

	return summary, tokensSaved, nil
}

// Extract extracts memories from messages (LLMSummarizer also implements MemoryExtractor).
func (s *LLMSummarizer) Extract(ctx context.Context, messages []*Message) ([]*ExtractedMemory, error) {
	summary, err := s.Summarize(ctx, messages)
	if err != nil {
		return nil, err
	}
	if summary == "" {
		return nil, nil
	}
	return []*ExtractedMemory{
		{
			Content:    summary,
			Importance: 0.6,
			Category:   string(CategoryContext),
			CreatedAt:  time.Now().UTC(),
		},
	}, nil
}

// formatMessagesForSummary formats messages for summary prompt.
func formatMessagesForSummary(messages []*Message) string {
	var sb strings.Builder
	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("[%d] %s: %s\n", i+1, msg.Role, msg.Content))
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("    Tool: %s(%s)\n", tc.Function.Name, tc.Function.Arguments))
			}
		}
	}
	return sb.String()
}

// AutoExtractor automatically extracts memories during session lifecycle.
type AutoExtractor struct {
	extractor  MemoryExtractor
	summarizer SummarizerExtractor
	config     Config
	messages   []*Message
	lastExtracted time.Time
	interval   time.Duration
}

// SummarizerExtractor combines summarization and extraction.
type SummarizerExtractor interface {
	Summarizer
	Extract(ctx context.Context, messages []*Message) ([]*ExtractedMemory, error)
}

// NewAutoExtractor creates a new automatic memory extractor.
func NewAutoExtractor(client llm.Provider, config Config) *AutoExtractor {
	ae := &AutoExtractor{
		extractor: NewLLMExtractor(client, config.Extractor),
		config:    config,
		interval:  5 * time.Minute, // Extract every 5 minutes by default
	}

	// Create combined summarizer/extractor if possible
	if sc, ok := ae.extractor.(SummarizerExtractor); ok {
		ae.summarizer = sc
	} else {
		ae.summarizer = NewLLMSummarizer(client, config.Summarizer)
	}

	return ae
}

// AddMessage adds a message and potentially triggers extraction.
func (ae *AutoExtractor) AddMessage(ctx context.Context, msg *Message) ([]*ExtractedMemory, error) {
	ae.messages = append(ae.messages, msg)

	// Check if we should extract memories
	shouldExtract := len(ae.messages) >= ae.config.MaxMessages ||
		time.Since(ae.lastExtracted) >= ae.interval

	if shouldExtract && ae.extractor != nil {
		memories, err := ae.Extract(ctx)
		if err != nil {
			return nil, err
		}
		ae.lastExtracted = time.Now()
		return memories, nil
	}

	return nil, nil
}

// Extract extracts memories from accumulated messages.
func (ae *AutoExtractor) Extract(ctx context.Context) ([]*ExtractedMemory, error) {
	if ae.extractor == nil || len(ae.messages) == 0 {
		return nil, nil
	}

	return ae.extractor.Extract(ctx, ae.messages)
}

// Summarize creates a summary of accumulated messages.
func (ae *AutoExtractor) Summarize(ctx context.Context) (string, error) {
	if ae.summarizer == nil || len(ae.messages) == 0 {
		return "", nil
	}

	return ae.summarizer.Summarize(ctx, ae.messages)
}

// Compress compresses accumulated messages.
func (ae *AutoExtractor) Compress(ctx context.Context) (string, int64, error) {
	if ae.summarizer == nil || len(ae.messages) == 0 {
		return "", 0, nil
	}

	return ae.summarizer.Compress(ctx, ae.messages, ae.config.Summarizer.MaxTokens)
}

// GetMessages returns all accumulated messages.
func (ae *AutoExtractor) GetMessages() []*Message {
	return ae.messages
}

// Clear clears accumulated messages.
func (ae *AutoExtractor) Clear() {
	ae.messages = nil
}

// SetInterval sets the extraction interval.
func (ae *AutoExtractor) SetInterval(interval time.Duration) {
	ae.interval = interval
}

// Deduper removes duplicate memories based on content similarity.
type Deduper struct {
	threshold float64
}

// NewDeduper creates a new memory deduper.
func NewDeduper(threshold float64) *Deduper {
	if threshold == 0 {
		threshold = 0.8
	}
	return &Deduper{threshold: threshold}
}

// Dedup removes duplicate memories.
func (d *Deduper) Dedup(memories []*ExtractedMemory) []*ExtractedMemory {
	if len(memories) <= 1 {
		return memories
	}

	var result []*ExtractedMemory
	for _, m := range memories {
		isDuplicate := false
		for _, existing := range result {
			if d.isSimilar(m.Content, existing.Content) {
				// Keep the one with higher importance
				if m.Importance > existing.Importance {
					*existing = *m
				}
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			result = append(result, m)
		}
	}

	return result
}

// isSimilar checks if two strings are similar (simple implementation).
func (d *Deduper) isSimilar(a, b string) bool {
	// Simple implementation: check if one is substring of another
	if len(a) > len(b) {
		return strings.Contains(a, b)
	}
	return strings.Contains(b, a)
}

package service

import (
	"backend/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Mem0Message represents a message sent to mem0 API
type Mem0Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Mem0Memory represents a memory returned from mem0 API
type Mem0Memory struct {
	ID        string `json:"id"`
	Memory    string `json:"memory"`
	UserID    string `json:"user_id"`
	Metadata  any    `json:"metadata,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// mem0AddRequest is the request body for POST /v1/memories/
type mem0AddRequest struct {
	Messages []Mem0Message `json:"messages"`
	UserID   string        `json:"user_id"`
}

// mem0SearchRequest is the request body for POST /v1/memories/search/
type mem0SearchRequest struct {
	Query  string `json:"query"`
	UserID string `json:"user_id"`
	TopK   int    `json:"top_k,omitempty"`
}

// Mem0Service provides interactions with the mem0 REST API
type Mem0Service struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewMem0Service creates a new Mem0Service; returns nil if no API key is configured
func NewMem0Service(cfg config.Mem0Config) *Mem0Service {
	if cfg.APIKey == "" {
		log.Println("mem0 API key not configured, memory features disabled")
		return nil
	}
	return &Mem0Service{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns whether the mem0 service has a valid API key
func (s *Mem0Service) IsConfigured() bool {
	return s != nil && s.apiKey != ""
}

// decodeMemories tries to decode a response body as either a plain JSON array
// or a {"results": [...]} wrapper, returning the memories either way.
func decodeMemories(body []byte) ([]Mem0Memory, error) {
	// Try plain array first (most common mem0 API response format)
	var memories []Mem0Memory
	if err := json.Unmarshal(body, &memories); err == nil {
		return memories, nil
	}

	// Fall back to wrapped format {"results": [...]}
	var wrapped struct {
		Results []Mem0Memory `json:"results"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return wrapped.Results, nil
}

// AddMemory saves a conversation as a memory for the given user
func (s *Mem0Service) AddMemory(userID uint, messages []Mem0Message) error {
	if !s.IsConfigured() {
		return nil
	}

	body := mem0AddRequest{
		Messages: messages,
		UserID:   strconv.FormatUint(uint64(userID), 10),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/memories/", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mem0 API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	log.Printf("mem0 add memory user=%d messages=%d", userID, len(messages))
	return nil
}

// SearchMemories searches for relevant memories based on a query
func (s *Mem0Service) SearchMemories(userID uint, query string, topK int) ([]Mem0Memory, error) {
	if !s.IsConfigured() {
		return nil, nil
	}

	if topK <= 0 {
		topK = 5
	}

	body := mem0SearchRequest{
		Query:  query,
		UserID: strconv.FormatUint(uint64(userID), 10),
		TopK:   topK,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/memories/search/", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mem0 API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return decodeMemories(respBody)
}

// ListMemories returns all memories for the given user
func (s *Mem0Service) ListMemories(userID uint) ([]Mem0Memory, error) {
	if !s.IsConfigured() {
		return nil, nil
	}

	url := fmt.Sprintf("%s/memories/?user_id=%d", s.baseURL, userID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mem0 API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return decodeMemories(respBody)
}

// DeleteMemory deletes a specific memory by ID
func (s *Mem0Service) DeleteMemory(memoryID string) error {
	if !s.IsConfigured() {
		return nil
	}

	url := fmt.Sprintf("%s/memories/%s/", s.baseURL, memoryID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mem0 API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	log.Printf("mem0 delete memory id=%s", memoryID)
	return nil
}

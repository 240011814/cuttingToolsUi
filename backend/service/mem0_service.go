package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Mem0Message represents a message sent to mem0 API
type Mem0Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Mem0Memory represents a memory returned from mem0 API
type Mem0Memory struct {
	ID        string   `json:"id"`
	Memory    string   `json:"memory"`
	Score     *float64 `json:"score,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	Metadata  any      `json:"metadata,omitempty"`
	Categories []string `json:"categories,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
}

// mem0AddV3Response is the async response from POST /v3/memories/add/
type mem0AddV3Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	EventID string `json:"event_id"`
}

// mem0ListV3Response is the paginated response from POST /v3/memories/
type mem0ListV3Response struct {
	Count    int          `json:"count"`
	Next     *string      `json:"next"`
	Previous *string      `json:"previous"`
	Results  []Mem0Memory `json:"results"`
}

// Mem0Service provides interactions with the mem0 REST API
type Mem0Service struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewMem0Service creates a new Mem0Service
func NewMem0Service(cfg Mem0Config) *Mem0Service {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.mem0.ai/v1"
	}
	svc := &Mem0Service{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	if cfg.APIKey == "" {
		log.Println("mem0 API key not configured, memory features disabled")
	} else {
		log.Println("mem0 service initialized")
	}
	return svc
}

// IsConfigured returns whether the mem0 service has a valid API key
func (s *Mem0Service) IsConfigured() bool {
	return s != nil && s.apiKey != ""
}

// ReloadConfig 热更新 mem0 配置，无需重启服务
func (s *Mem0Service) ReloadConfig(cfg Mem0Config) {
	s.apiKey = cfg.APIKey
	if cfg.BaseURL != "" {
		s.baseURL = cfg.BaseURL
	}
	log.Printf("mem0 config reloaded: baseURL=%s", s.baseURL)
}

// v3BaseURL derives the v3 base URL from the configured baseURL.
// If baseURL ends with /v1, replaces with /v3; otherwise appends /v3.
func (s *Mem0Service) v3BaseURL() string {
	base := strings.TrimRight(s.baseURL, "/")
	if strings.HasSuffix(base, "/v1") {
		return base[:len(base)-3] + "/v3"
	}
	return base + "/v3"
}

// doRequest is a helper to perform an HTTP request with auth header
func (s *Mem0Service) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mem0 API error: status=%d body=%s", resp.StatusCode, string(body))
	}
	return body, nil
}

// AddMemory uses the v3 async API (POST /v3/memories/add/) to add memories.
// Returns the event ID for tracking the async processing status.
func (s *Mem0Service) AddMemory(userID uint, messages []Mem0Message, metadata map[string]any) (*mem0AddV3Response, error) {
	if !s.IsConfigured() {
		return nil, nil
	}

	reqBody := map[string]any{
		"messages": messages,
		"user_id":  strconv.FormatUint(uint64(userID), 10),
	}
	if len(metadata) > 0 {
		reqBody["metadata"] = metadata
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.v3BaseURL()+"/memories/add/", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	body, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var result mem0AddV3Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	log.Printf("mem0 v3 add memory user=%d event_id=%s status=%s", userID, result.EventID, result.Status)
	return &result, nil
}

// SearchMemories uses v3 API (POST /v3/memories/search/) to search for relevant memories.
func (s *Mem0Service) SearchMemories(userID uint, query string, topK int) ([]Mem0Memory, error) {
	if !s.IsConfigured() {
		return nil, nil
	}

	if topK <= 0 {
		topK = 5
	}

	body := map[string]any{
		"query": query,
		"filters": map[string]string{
			"user_id": strconv.FormatUint(uint64(userID), 10),
		},
		"top_k": topK,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.v3BaseURL()+"/memories/search/", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	respBody, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []Mem0Memory `json:"results"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result.Results, nil
}

// ListMemories uses v3 API (POST /v3/memories/) to list all memories for the given user.
// Supports pagination via page (1-indexed) and pageSize (max 200).
func (s *Mem0Service) ListMemories(userID uint, page, pageSize int) (*mem0ListV3Response, error) {
	if !s.IsConfigured() {
		return nil, nil
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	body := map[string]any{
		"filters": map[string]string{
			"user_id": strconv.FormatUint(uint64(userID), 10),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/memories/?page=%d&page_size=%d", s.v3BaseURL(), page, pageSize)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	respBody, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var result mem0ListV3Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// DeleteMemory deletes a specific memory by ID (v1 — no v3 equivalent)
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

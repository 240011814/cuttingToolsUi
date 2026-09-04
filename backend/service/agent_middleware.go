package service

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"sync"

	"backend/model"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func init() {
	gob.Register(&ToolApprovalInfo{})
	gob.Register(&ToolApprovalResult{})
}

type ToolApprovalInfo struct {
	ToolName  string `json:"tool_name"`
	Arguments string `json:"arguments"`
	CallID    string `json:"call_id"`
}

type ToolApprovalResult struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

// confirmRequiredCache caches tool confirm_required status
var (
	confirmRequiredCache = make(map[string]bool)
	confirmRequiredMu    sync.RWMutex
)

// InvalidateConfirmRequiredCache clears the cache for a specific tool or all tools
func InvalidateConfirmRequiredCache(toolName string) {
	confirmRequiredMu.Lock()
	defer confirmRequiredMu.Unlock()
	if toolName == "" {
		confirmRequiredCache = make(map[string]bool)
	} else {
		delete(confirmRequiredCache, toolName)
	}
}

// IsConfirmRequired checks if a tool requires confirmation, with caching
func isConfirmRequired(toolName string) bool {
	confirmRequiredMu.RLock()
	if cached, ok := confirmRequiredCache[toolName]; ok {
		confirmRequiredMu.RUnlock()
		return cached
	}
	confirmRequiredMu.RUnlock()

	var dbTool model.AITool
	if err := DB.Where("name = ? AND enabled = ?", toolName, true).First(&dbTool).Error; err != nil {
		return false
	}

	confirmRequiredMu.Lock()
	confirmRequiredCache[toolName] = dbTool.ConfirmRequired
	confirmRequiredMu.Unlock()

	return dbTool.ConfirmRequired
}

func NewApprovalMiddleware() adk.AgentMiddleware {
	return adk.AgentMiddleware{
		WrapToolCall: compose.ToolMiddleware{
			Invokable: func(endpoint compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
				return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
					if !isConfirmRequired(input.Name) {
						return endpoint(ctx, input)
					}

					wasInterrupted, _, storedArgs := compose.GetInterruptState[string](ctx)

					if !wasInterrupted {
						return nil, compose.StatefulInterrupt(ctx, &ToolApprovalInfo{
							ToolName:  input.Name,
							Arguments: input.Arguments,
							CallID:    input.CallID,
						}, input.Arguments)
					}

					isTarget, hasData, data := compose.GetResumeContext[*ToolApprovalResult](ctx)
					if isTarget && hasData {
						if data.Approved {
							input.Arguments = storedArgs
							return endpoint(ctx, input)
						}
						reason := "操作已被用户拒绝"
						if data.Reason != "" {
							reason = fmt.Sprintf("操作已被用户拒绝: %s", data.Reason)
						}
						return &compose.ToolOutput{Result: reason}, nil
					}

					return nil, compose.StatefulInterrupt(ctx, &ToolApprovalInfo{
						ToolName:  input.Name,
						Arguments: storedArgs,
						CallID:    input.CallID,
					}, storedArgs)
				}
			},
			Streamable: func(endpoint compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
				return func(ctx context.Context, input *compose.ToolInput) (*compose.StreamToolOutput, error) {
					if !isConfirmRequired(input.Name) {
						return endpoint(ctx, input)
					}

					wasInterrupted, _, storedArgs := compose.GetInterruptState[string](ctx)

					if !wasInterrupted {
						return nil, compose.StatefulInterrupt(ctx, &ToolApprovalInfo{
							ToolName:  input.Name,
							Arguments: input.Arguments,
							CallID:    input.CallID,
						}, input.Arguments)
					}

					isTarget, hasData, data := compose.GetResumeContext[*ToolApprovalResult](ctx)
					if isTarget && hasData {
						if data.Approved {
							input.Arguments = storedArgs
							return endpoint(ctx, input)
						}
						reason := "操作已被用户拒绝"
						if data.Reason != "" {
							reason = fmt.Sprintf("操作已被用户拒绝: %s", data.Reason)
						}
						return &compose.StreamToolOutput{
							Result: schema.StreamReaderFromArray([]string{reason}),
						}, nil
					}

					return nil, compose.StatefulInterrupt(ctx, &ToolApprovalInfo{
						ToolName:  input.Name,
						Arguments: storedArgs,
						CallID:    input.CallID,
					}, storedArgs)
				}
			},
		},
	}
}

func NewLoggingMiddleware() adk.AgentMiddleware {
	return adk.AgentMiddleware{
		BeforeChatModel: func(ctx context.Context, state *adk.ChatModelAgentState) error {
			log.Printf("[eino] chat model start: %d messages", len(state.Messages))
			for _, m := range state.Messages {
				log.Printf("[eino]   [%s] %s", m.Role, m.Content)
			}
			return nil
		},
		AfterChatModel: func(ctx context.Context, state *adk.ChatModelAgentState) error {
			for _, m := range state.Messages {
				if m.Role == schema.Assistant {
					if m.ResponseMeta != nil && m.ResponseMeta.Usage.TotalTokens > 0 {
						log.Printf("[eino] chat model end: prompt_tokens=%d completion_tokens=%d total_tokens=%d",
							m.ResponseMeta.Usage.PromptTokens,
							m.ResponseMeta.Usage.CompletionTokens,
							m.ResponseMeta.Usage.TotalTokens)
					}
					if m.Content != "" {
						log.Printf("[eino] chat model end: reply=%s", m.Content)
					}
				}
			}
			return nil
		},
	}
}

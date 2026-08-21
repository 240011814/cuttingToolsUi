package service

import (
	"context"
	"encoding/gob"
	"fmt"
	"sync"

	"backend/model"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
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

type ApprovalMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
}

func (m *ApprovalMiddleware) WrapInvokableToolCall(
	_ context.Context,
	endpoint adk.InvokableToolCallEndpoint,
	tCtx *adk.ToolContext,
) (adk.InvokableToolCallEndpoint, error) {
	if !isConfirmRequired(tCtx.Name) {
		return endpoint, nil
	}

	return func(ctx context.Context, args string, opts ...tool.Option) (string, error) {
		wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)

		if !wasInterrupted {
			return "", tool.StatefulInterrupt(ctx, &ToolApprovalInfo{
				ToolName:  tCtx.Name,
				Arguments: args,
				CallID:    tCtx.CallID,
			}, args)
		}

		isTarget, hasData, data := tool.GetResumeContext[*ToolApprovalResult](ctx)
		if isTarget && hasData {
			if data.Approved {
				return endpoint(ctx, storedArgs, opts...)
			}
			reason := "操作已被用户拒绝"
			if data.Reason != "" {
				reason = fmt.Sprintf("操作已被用户拒绝: %s", data.Reason)
			}
			return reason, nil
		}

		return "", tool.StatefulInterrupt(ctx, &ToolApprovalInfo{
			ToolName:  tCtx.Name,
			Arguments: storedArgs,
			CallID:    tCtx.CallID,
		}, storedArgs)
	}, nil
}

func (m *ApprovalMiddleware) WrapStreamableToolCall(
	_ context.Context,
	endpoint adk.StreamableToolCallEndpoint,
	tCtx *adk.ToolContext,
) (adk.StreamableToolCallEndpoint, error) {
	if !isConfirmRequired(tCtx.Name) {
		return endpoint, nil
	}

	return func(ctx context.Context, args string, opts ...tool.Option) (*schema.StreamReader[string], error) {
		wasInterrupted, _, storedArgs := tool.GetInterruptState[string](ctx)

		if !wasInterrupted {
			return nil, tool.StatefulInterrupt(ctx, &ToolApprovalInfo{
				ToolName:  tCtx.Name,
				Arguments: args,
				CallID:    tCtx.CallID,
			}, args)
		}

		isTarget, hasData, data := tool.GetResumeContext[*ToolApprovalResult](ctx)
		if isTarget && hasData {
			if data.Approved {
				return endpoint(ctx, storedArgs, opts...)
			}
			reason := "操作已被用户拒绝"
			if data.Reason != "" {
				reason = fmt.Sprintf("操作已被用户拒绝: %s", data.Reason)
			}
			return schema.StreamReaderFromArray([]string{reason}), nil
		}

		return nil, tool.StatefulInterrupt(ctx, &ToolApprovalInfo{
			ToolName:  tCtx.Name,
			Arguments: storedArgs,
			CallID:    tCtx.CallID,
		}, storedArgs)
	}, nil
}

package api

import (
	"backend/model"
	"backend/service"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type ToolApprovalRequest struct {
	CheckPointID    string        `json:"checkpoint_id" binding:"required"`
	InterruptID     string        `json:"interrupt_id" binding:"required"`
	Approved        bool          `json:"approved"`
	Reason          string        `json:"reason,omitempty"`
	HistoryID       uint          `json:"history_id"`
	TrainingType    string        `json:"training_type"`
	CustomTrainingID *uint        `json:"custom_training_id"`
	Messages        []ChatMessage `json:"messages"`
}

func HandleToolApproval(agentService *service.AIAgentService, historyService *service.HistoryService, mem0Service *service.Mem0Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ToolApprovalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Parse checkpoint_id format: userID_historyID
		parts := strings.SplitN(req.CheckPointID, "_", 2)
		var historyID uint
		if len(parts) == 2 {
			if hid, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				historyID = uint(hid)
			}
		}
		if historyID == 0 {
			historyID = req.HistoryID
		}

		inputMessages := make([]*schema.Message, len(req.Messages))
		for i, m := range req.Messages {
			inputMessages[i] = &schema.Message{
				Role:    schema.RoleType(m.Role),
				Content: m.Content,
			}
		}

		iter, err := agentService.ResumeToolApproval(req.CheckPointID, req.InterruptID, req.Approved, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resume: " + err.Error()})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		var fullAssistantReply string
		var fullThinkingContent string

		c.Stream(func(w io.Writer) bool {
			event, ok := iter.Next()
			if !ok {
				log.Printf("[tool-approval] stream completed user=%d reply_chars=%d", userID.(uint), len(fullAssistantReply))
				c.SSEvent("message", "[DONE]")

				// Save history
				allMessages := make([]*schema.Message, 0, len(inputMessages)+1)
				allMessages = append(allMessages, inputMessages...)
				if fullAssistantReply != "" {
					assistantMsg := &schema.Message{Role: schema.Assistant, Content: fullAssistantReply}
					allMessages = append(allMessages, assistantMsg)
				}

				msgs := make([]model.OpenAIMessage, len(allMessages))
				for i, m := range allMessages {
					msgs[i] = model.OpenAIMessage{Role: string(m.Role), Content: m.Content}
				}
				if fullThinkingContent != "" && len(msgs) > 0 {
					msgs[len(msgs)-1].ThinkingContent = fullThinkingContent
				}

				title := "AI 训练对话"
				for _, m := range inputMessages {
					if m.Role == schema.User && m.Content != "" {
						title = m.Content
						if len(title) > 20 {
							title = title[:20] + "..."
						}
						break
					}
				}

				saveFunc := func() {
					_, saveErr := historyService.SaveHistory(userID.(uint), historyID, req.TrainingType, req.CustomTrainingID, title, msgs, false)
					if saveErr != nil {
						log.Printf("[tool-approval] save history failed user=%d err=%v", userID.(uint), saveErr)
					}
				}

				saveMemoryFunc := func() {
					if mem0Service != nil && mem0Service.IsConfigured() {
						memMessages := make([]service.Mem0Message, 0, len(allMessages))
						for _, m := range allMessages {
							if m.Role == schema.System {
								continue
							}
							memMessages = append(memMessages, service.Mem0Message{
								Role:    string(m.Role),
								Content: m.Content,
							})
						}
						if len(memMessages) > 0 {
							if _, err := mem0Service.AddMemory(userID.(uint), memMessages, nil); err != nil {
								log.Printf("[tool-approval] mem0 save memory failed user=%d err=%v", userID.(uint), err)
							}
						}
					}
				}

				go saveFunc()
				go saveMemoryFunc()

				return false
			}

			if event.Err != nil {
				c.SSEvent("error", gin.H{"error": event.Err.Error()})
				return false
			}

			if event.Action != nil && event.Action.Interrupted != nil {
				for _, ictx := range event.Action.Interrupted.InterruptContexts {
					if approvalInfo, ok := ictx.Info.(*service.ToolApprovalInfo); ok {
						c.SSEvent("tool_approval", gin.H{
							"tool_name":     approvalInfo.ToolName,
							"arguments":     approvalInfo.Arguments,
							"checkpoint_id": req.CheckPointID,
							"interrupt_id":  ictx.ID,
						})
						return false
					}
				}
			}

			if event.Output != nil && event.Output.MessageOutput != nil {
				mv := event.Output.MessageOutput
				if mv.Role == schema.Tool {
					c.SSEvent("message", gin.H{"thinking": "正在调用工具..."})
					return true
				}
				if mv.IsStreaming && mv.MessageStream != nil {
					for {
						msg, err := mv.MessageStream.Recv()
						if errors.Is(err, io.EOF) {
							break
						}
						if err != nil {
							break
						}
						if msg.ReasoningContent != "" {
							fullThinkingContent += msg.ReasoningContent
							c.SSEvent("message", gin.H{"reasoning_content": msg.ReasoningContent})
						}
						if msg.Content != "" {
							fullAssistantReply += msg.Content
							c.SSEvent("message", gin.H{"content": msg.Content})
						}
					}
				} else if mv.Message != nil {
					msg := mv.Message
					if msg.ReasoningContent != "" {
						fullThinkingContent += msg.ReasoningContent
						c.SSEvent("message", gin.H{"reasoning_content": msg.ReasoningContent})
					}
					if msg.Content != "" {
						fullAssistantReply += msg.Content
						c.SSEvent("message", gin.H{"content": msg.Content})
					}
				}
			}

			return true
		})
	}
}

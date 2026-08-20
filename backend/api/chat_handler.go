package api

import (
	"backend/model"
	"backend/service"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type ChatMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ChatRequest struct {
	HistoryID        uint          `json:"history_id"`
	TrainingType     string        `json:"training_type"`
	CustomTrainingID *uint         `json:"custom_training_id"`
	AgentID          uint          `json:"agent_id" binding:"required"`
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages" binding:"required"`
}

func HandleChatStream(agentService *service.AIAgentService, historyService *service.HistoryService, mem0Service *service.Mem0Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestStart := time.Now()
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
			return
		}

		userID, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		inputMessages := make([]*schema.Message, len(req.Messages))
		for i, m := range req.Messages {
			inputMessages[i] = &schema.Message{
				Role:    schema.RoleType(m.Role),
				Content: m.Content,
			}
		}

		iter, err := agentService.ChatStream(userID.(uint), req.AgentID, inputMessages, req.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI: " + err.Error()})
			return
		}
		log.Printf("[chat] user=%d agent=%d model=%s messages=%d init_ms=%d", userID.(uint), req.AgentID, req.Model, len(inputMessages), time.Since(requestStart).Milliseconds())
		for _, m := range inputMessages {
			log.Printf("[chat]   [%s] %s", m.Role, m.Content)
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		var fullAssistantReply string
		var fullThinkingContent string
		firstTokenLogged := false
		c.Stream(func(w io.Writer) bool {
			event, ok := iter.Next()
			if !ok {
				log.Printf("chat stream completed user=%d reply_chars=%d total_ms=%d", userID.(uint), len(fullAssistantReply), time.Since(requestStart).Milliseconds())
				c.SSEvent("message", "[DONE]")

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
					historyID, saveErr := historyService.SaveHistory(userID.(uint), req.HistoryID, req.TrainingType, req.CustomTrainingID, title, msgs, false)
					if saveErr == nil && req.HistoryID == 0 {
						c.SSEvent("history_id", gin.H{"history_id": historyID, "title": title})
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
								log.Printf("mem0 save memory failed user=%d err=%v", userID.(uint), err)
							}
						}
					}
				}

				if req.HistoryID == 0 {
					saveFunc()
				} else {
					go saveFunc()
				}
				go saveMemoryFunc()

				return false
			}

			if event.Err != nil {
				c.SSEvent("error", gin.H{"error": event.Err.Error()})
				return false
			}

			if event.Output != nil && event.Output.MessageOutput != nil {
				mv := event.Output.MessageOutput
				if mv.Role == schema.Tool {
					return true
				}
				if mv.IsStreaming && mv.MessageStream != nil {
					for {
						msg, err := mv.MessageStream.Recv()
						if errors.Is(err, io.EOF) {
							break
						}
						if err != nil {
							log.Printf("stream recv error: %v", err)
							break
						}
						if !firstTokenLogged {
							firstTokenLogged = true
							log.Printf("chat first token user=%d first_token_ms=%d", userID.(uint), time.Since(requestStart).Milliseconds())
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

// HandleListModels 返回所有已启用的模型列表
func HandleListModels(aiAgentService *service.AIAgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		models, err := aiAgentService.ListEnabledModels()
		if err != nil {
			SendError(c, "500", "获取模型列表失败: "+err.Error())
			return
		}
		SendSuccess(c, models)
	}
}

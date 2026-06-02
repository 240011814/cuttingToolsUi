package api

import (
	"backend/model"
	"backend/service"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type ChatRequest struct {
	HistoryID        uint                           `json:"history_id"`
	TrainingType     string                         `json:"training_type"`
	CustomTrainingID *uint                          `json:"custom_training_id"`
	Model            string                         `json:"model"`
	Messages         []openai.ChatCompletionMessage `json:"messages" binding:"required"`
}

func HandleChatStream(aiService *service.AIService, historyService *service.HistoryService) gin.HandlerFunc {
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

		stream, err := aiService.ChatStream(c.Request.Context(), req.Messages, req.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI: " + err.Error()})
			return
		}
		defer stream.Close()
		log.Printf("chat stream established user=%d model=%s messages=%d init_ms=%d", userID.(uint), req.Model, len(req.Messages), time.Since(requestStart).Milliseconds())

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		var fullAssistantReply string
		firstTokenLogged := false
		c.Stream(func(w io.Writer) bool {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				log.Printf("chat stream completed user=%d model=%s reply_chars=%d total_ms=%d", userID.(uint), req.Model, len(fullAssistantReply), time.Since(requestStart).Milliseconds())
				c.SSEvent("message", "[DONE]")

				allMessages := append(req.Messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: fullAssistantReply,
				})

				msgs := make([]model.OpenAIMessage, len(allMessages))
				for i, m := range allMessages {
					msgs[i] = model.OpenAIMessage{Role: m.Role, Content: m.Content}
				}

				title := "AI 训练对话"
				if len(allMessages) > 1 && allMessages[1].Role == openai.ChatMessageRoleUser {
					title = allMessages[1].Content
					if len(title) > 20 {
						title = title[:20] + "..."
					}
				}

				saveFunc := func() {
					historyID, saveErr := historyService.SaveHistory(userID.(uint), req.HistoryID, req.TrainingType, req.CustomTrainingID, title, msgs, false)
					if saveErr == nil && req.HistoryID == 0 {
						c.SSEvent("history_id", gin.H{"history_id": historyID, "title": title})
					}
				}

				if req.HistoryID == 0 {
					saveFunc()
				} else {
					go saveFunc()
				}

				return false
			}

			if err != nil {
				c.SSEvent("error", gin.H{"error": err.Error()})
				return false
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					if !firstTokenLogged {
						firstTokenLogged = true
						log.Printf("chat first token user=%d model=%s first_token_ms=%d", userID.(uint), req.Model, time.Since(requestStart).Milliseconds())
					}
					fullAssistantReply += content
					c.SSEvent("message", gin.H{"content": content})
				}
			}

			return true
		})
	}
}

// HandleListModels 返回所有已启用的模型列表
func HandleListModels(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		models, err := aiService.ListEnabledModels()
		if err != nil {
			SendError(c, "500", "获取模型列表失败: "+err.Error())
			return
		}
		SendSuccess(c, models)
	}
}

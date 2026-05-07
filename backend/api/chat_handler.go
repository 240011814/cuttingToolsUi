package api

import (
	"backend/service"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

// ChatRequest 定义了前端传入的数据格式
type ChatRequest struct {
	HistoryID    uint                           `json:"history_id"`
	TrainingType string                         `json:"training_type"`
	Model        string                         `json:"model"`
	Messages     []openai.ChatCompletionMessage `json:"messages" binding:"required"`
}

// HandleChatStream 处理流式聊天请求
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

		// 确保有 DEEPSEEK_API_KEY 才发起请求，否则返回模拟数据以方便前端调试
		// (可选逻辑：可以判断环境变量，为了简单这里直接调用)

		stream, err := aiService.ChatStream(c.Request.Context(), req.Messages, req.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI: " + err.Error()})
			return
		}
		defer stream.Close()
		log.Printf("chat stream established user=%d model=%s messages=%d init_ms=%d", userID.(uint), req.Model, len(req.Messages), time.Since(requestStart).Milliseconds())

		// 设置 SSE 必需的 HTTP Header
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// 允许前端处理流的代理
		c.Header("X-Accel-Buffering", "no")

		// 使用 c.Stream 实时写入数据
		var fullAssistantReply string
		firstTokenLogged := false
		c.Stream(func(w io.Writer) bool {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				log.Printf("chat stream completed user=%d model=%s reply_chars=%d total_ms=%d", userID.(uint), req.Model, len(fullAssistantReply), time.Since(requestStart).Milliseconds())
				// 发送结束标志
				c.SSEvent("message", "[DONE]")

				// 完整收集后，保存历史记录
				req.Messages = append(req.Messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: fullAssistantReply,
				})
				messagesBytes, _ := json.Marshal(req.Messages)
				title := "AI 训练对话"
				if len(req.Messages) > 1 && req.Messages[1].Role == openai.ChatMessageRoleUser {
					title = req.Messages[1].Content
					if len(title) > 20 {
						title = title[:20] + "..."
					}
				}
				// 提取保存逻辑
				saveFunc := func() {
					historyID, saveErr := historyService.SaveHistory(userID.(uint), req.HistoryID, req.TrainingType, title, string(messagesBytes), "auto")
					if saveErr == nil && req.HistoryID == 0 {
						c.SSEvent("history_id", gin.H{"history_id": historyID})
					}
				}

				if req.HistoryID == 0 {
					// 首次对话：同步保存以获取 ID
					saveFunc()
				} else {
					// 后续对话：异步保存，不阻塞连接关闭
					go saveFunc()
				}

				return false // 结束 stream
			}

			if err != nil {
				// 发生错误时，将错误发送给前端并结束流
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
					// 发送包含内容片段的 JSON
					c.SSEvent("message", gin.H{"content": content})
				}
			}

			return true // 继续读取
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

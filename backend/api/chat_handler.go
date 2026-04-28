package api

import (
	"errors"
	"io"
	"net/http"

	"backend/service"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

// ChatRequest 定义了前端传入的数据格式
type ChatRequest struct {
	Messages []openai.ChatCompletionMessage `json:"messages" binding:"required"`
}

// HandleChatStream 处理流式聊天请求
func HandleChatStream(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
			return
		}

		// 确保有 DEEPSEEK_API_KEY 才发起请求，否则返回模拟数据以方便前端调试
		// (可选逻辑：可以判断环境变量，为了简单这里直接调用)

		stream, err := aiService.ChatStream(c.Request.Context(), req.Messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call AI: " + err.Error()})
			return
		}
		defer stream.Close()

		// 设置 SSE 必需的 HTTP Header
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// 允许前端处理流的代理
		c.Header("X-Accel-Buffering", "no")

		// 使用 c.Stream 实时写入数据
		c.Stream(func(w io.Writer) bool {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// 发送结束标志
				c.SSEvent("message", "[DONE]")
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
					// 发送包含内容片段的 JSON
					c.SSEvent("message", gin.H{"content": content})
				}
			}

			return true // 继续读取
		})
	}
}

package ai

import (
	"context"
	"fmt"
	"strings"
	"ultahost-ai-gateway/internal/config"

	"github.com/sashabaranov/go-openai"
)

func SummarizeResponse(rawResponse string) (string, error) {
	client := openai.NewClient(config.AppConfig.OpenAIKey)

	systemMsg := `You are a helpful assistant that converts technical server output into user-friendly summaries.`

	userMsg := fmt.Sprintf(`Here is the raw server response: "%s". 
Convert it into a short, clear, and friendly sentence that a non-technical user can easily understand.`, rawResponse)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemMsg},
			{Role: "user", Content: userMsg},
		},
	})
	if err != nil {
		return "", err
	}

	summary := strings.TrimSpace(resp.Choices[0].Message.Content)
	return summary, nil
}

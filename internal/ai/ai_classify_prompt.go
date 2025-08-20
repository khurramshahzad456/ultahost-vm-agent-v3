package ai

import (
	"context"
	"fmt"
	"strings"
	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/pkg/models"

	"github.com/sashabaranov/go-openai"
)

func ClassifyPromptCategory(req *models.CategoryRequest) (string, error) {
	client := openai.NewClient(config.AppConfig.OpenAIKey)

	categoryList := strings.Join(req.Categories, ", ")

	systemMsg := fmt.Sprintf(`You are an intelligent assistant. Given a user query and a list of internal category keys [%s], 
your job is to return ONLY the single most relevant category key from the list — no explanations, no formatting, just the raw category key.
If there's no suitable match, return "unknown" only.`, categoryList)

	userMsg := fmt.Sprintf("User query: %q", req.Query)

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

	// Extra trim to clean up whitespace or formatting like "Category: billing"
	category := strings.TrimSpace(resp.Choices[0].Message.Content)
	category = strings.ToLower(category)
	category = strings.TrimPrefix(category, "category:")
	category = strings.TrimSpace(category)

	return category, nil
}

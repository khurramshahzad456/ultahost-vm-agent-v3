package ai

import (
	"context"
	"fmt"
	"strings"
	"ultahost-ai-gateway/internal/config"

	"github.com/sashabaranov/go-openai"
)

// ClassifyFunctionWithinAgent suggests the most appropriate function name
// from the given list based on the user query.
func ClassifyFunctionWithinAgent(query string, functionList []string) (string, error) {
	client := openai.NewClient(config.AppConfig.OpenAIKey)

	functions := strings.Join(functionList, ", ")

	systemMsg := fmt.Sprintf(`You are an intelligent assistant. Given a user query and a list of function names [%s], 
your job is to return ONLY the single most relevant function name from the list — no explanations, no formatting, just the raw function name.
If there's no suitable match, return "unknown" only.`, functions)

	userMsg := fmt.Sprintf("User query: %q", query)

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

	// Clean up and normalize the output
	fn := strings.TrimSpace(resp.Choices[0].Message.Content)
	fn = strings.ToLower(fn)
	fn = strings.TrimPrefix(fn, "function:")
	fn = strings.TrimSpace(fn)

	return fn, nil
}

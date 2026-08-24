package main

import (
	"context"
	"fmt"
	"os"
	"github.com/sashabaranov/go-openai"
)

func GenerateTitle(prompt string) string {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "تلاوة خاشعة - سورة البقرة - الطيبات"
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "z-ai/glm-4.5:free",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "انت خبير سيو لقناة الطيبات الاسلامية"},
			{Role: "user", Content: prompt},
		},
	})
	if err!= nil {
		fmt.Println("AI Error:", err)
		return prompt
	}
	return resp.Choices[0].Message.Content
}

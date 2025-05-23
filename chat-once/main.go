package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type ChatClient struct {
	client *openai.Client
	model  string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("go run main.go <your question>")
		return
	}
	question := os.Args[1]

	_ = godotenv.Load()

	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_API_BASE")
	model := os.Getenv("OPENAI_API_MODEL")
	if apiKey == "" || baseURL == "" || model == "" {
		fmt.Println("检查环境变量设置")
		return
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	chatClient := &ChatClient{
		client: client,
		model:  model,
	}

	reply, err := chatClient.ProcessQuery(question)
	if err != nil {
		fmt.Printf("调用失败: %v\n", err)
		return
	}
	fmt.Println(reply)
}

func (c *ChatClient) ProcessQuery(content string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: content},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}
	return resp.Choices[0].Message.Content, nil
}

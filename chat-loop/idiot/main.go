package main

// 该版本为无上下文版本

// 有上下文（保留历史消息）：
// User：谁是爱因斯坦？
// Assistant：爱因斯坦是20世纪著名的物理学家...
// User：他最著名的理论是什么？
// Assistant：他最著名的理论是相对论，特别是广义相对论和狭义相对论。

// 无上下文（每轮都单独提问）：
// User：谁是爱因斯坦？
// Assistant：爱因斯坦是20世纪著名的物理学家...
// User：他最著名的理论是什么？
// Assistant：请明确你说的“他”是谁？

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type ChatClient struct {
	client *openai.Client
	model  string
}

func main() {
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

	chatClient.ChatLoop()
}

func (c *ChatClient) ChatLoop() {
	fmt.Print("Type your queries or 'quit' to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nUser: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if strings.ToLower(userInput) == "quit" {
			break
		}
		if userInput == "" {
			continue
		}

		response, err := c.ProcessQuery(userInput)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}

		fmt.Printf("Assistant: %s\n", response)
	}
}

func (c *ChatClient) ProcessQuery(userInput string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: userInput},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}

	response := resp.Choices[0].Message.Content
	return response, nil
}

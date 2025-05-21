package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var (
	apiKey  string
	baseURL string
	model   string
)

func main() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		fmt.Println("未找到 .env 文件加载失败:", err)
	}

	// 从环境变量读取配置
	apiKey = os.Getenv("OPENAI_API_KEY")
	baseURL = os.Getenv("OPENAI_API_BASE")
	model = os.Getenv("OPENAI_API_MODEL")
	if apiKey == "" || baseURL == "" || model == "" {
		fmt.Println("请在 .env 文件中设置 OPENAI_API_KEY，OPENAI_API_BASE，OPENAI_API_MODEL")
		return
	}

	// 从命令行获取用户输入
	if len(os.Args) < 2 {
		fmt.Println("请提供要提问的内容，例如：go run main.go \"解释量子计算\"")
		return
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	question := os.Args[1]
	result, err := chat(client, question)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(result)
}

func chat(client *openai.Client, content string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: content},
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

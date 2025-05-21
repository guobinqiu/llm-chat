package main

// 该版本为无上下文版本

// 有上下文（保留历史消息）：
// 你：谁是爱因斯坦？
// 🤖：爱因斯坦是20世纪著名的物理学家...
// 你：他最著名的理论是什么？
// 🤖：他最著名的理论是相对论，特别是广义相对论和狭义相对论。

// 无上下文（每轮都单独提问）：
// 你：谁是爱因斯坦？
// 🤖：爱因斯坦是20世纪著名的物理学家...
// 你：他最著名的理论是什么？
// 🤖：请明确你说的“他”是谁？

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

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("欢迎使用 Chat 模式，输入内容与模型对话，输入 `exit` 退出。")
	for {
		fmt.Print("\n你：")
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "exit" || userInput == "quit" {
			break
		}
		if userInput == "" {
			continue
		}

		response, err := chat(client, userInput)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}

		fmt.Printf("🤖：%s\n", response)
	}
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

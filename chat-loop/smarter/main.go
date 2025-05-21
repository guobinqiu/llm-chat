package main

// 该版本为保留上下文版本
// 上下文信息越来越大需要优化
// 1. 上下文裁剪（保留最近对话）
// 2. 摘要长对话内容 （压缩旧消息）

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

const retainNum = 5

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

	// 用于存储历史消息，实现多轮对话
	var messages []openai.ChatCompletionMessage

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

		// 添加问题到历史消息
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user", // 用户问
			Content: userInput,
		})

		// 合并压缩上下文，保留最近 retainNum 条，摘要旧消息
		var err error
		messages, err = merge(client, messages, retainNum)
		if err != nil {
			fmt.Printf("合并上下文失败: %v\n", err)
			continue
		}

		response, err := chat(client, messages)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}

		fmt.Printf("🤖：%s\n", response)

		// 添加回答到历史消息
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "assistant", //助手答
			Content: response,
		})
	}
}

func chat(client *openai.Client, messages []openai.ChatCompletionMessage) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}
	return resp.Choices[0].Message.Content, nil
}

// 摘要长对话内容
func summarize(client *openai.Client, history []openai.ChatCompletionMessage) (string, error) {
	summaryPrompt := "请总结以下对话，保留用户意图与助手回答的核心内容：\n\n"
	for _, msg := range history {
		summaryPrompt += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
	}

	message := openai.ChatCompletionMessage{Role: "user", Content: summaryPrompt}
	response, err := chat(client, []openai.ChatCompletionMessage{message})
	if err != nil {
		return "", err
	}
	return response, nil
}

// 合并摘要和最近对话
func merge(client *openai.Client, messages []openai.ChatCompletionMessage, retainNum int) ([]openai.ChatCompletionMessage, error) {
	if len(messages) <= retainNum {
		return messages, nil
	}

	past := messages[:len(messages)-retainNum]
	recent := messages[len(messages)-retainNum:] // 上下文裁剪（保留最近对话）

	summary, err := summarize(client, past)
	if err != nil {
		return nil, err
	}

	// 压缩旧消息变成一句 system prompt
	newMessages := []openai.ChatCompletionMessage{
		{Role: "system", Content: "以下是之前对话的总结：" + summary},
	}

	newMessages = append(newMessages, recent...)
	return newMessages, nil
}

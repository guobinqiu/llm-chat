package main

// 该版本为保留上下文版本
// 优化了上下文信息
// 历史信息超过指定条数就合并成一条摘要信息

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
	client    *openai.Client
	model     string
	messages  []openai.ChatCompletionMessage // 用于存储历史消息，实现多轮对话
	retainNum int                            // 超过n条就合并
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
		client:    client,
		model:     model,
		messages:  make([]openai.ChatCompletionMessage, 0),
		retainNum: 5,
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
		if userInput == "quit" {
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
	// 添加问题到历史消息
	c.messages = append(c.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userInput,
	})

	// 合并上下文
	if err := c.Merge(); err != nil {
		return "", fmt.Errorf("合并上下文失败: %v", err)
	}

	// 调用大模型获取回答
	response, err := c.CallOpenAI(c.messages)
	if err != nil {
		return "", err
	}

	// 添加回答到历史消息 下一次用户提问时会用到它
	c.messages = append(c.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response,
	})

	return response, nil
}

func (c *ChatClient) Merge() error {
	if len(c.messages) <= c.retainNum {
		return nil
	}

	// 让大模型总结成一条摘要信息
	summary, err := c.Summarize(c.messages)
	if err != nil {
		return nil
	}

	// 重写messages
	c.messages = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "以下是之前对话的总结：" + summary},
	}

	return nil
}

func (c *ChatClient) Summarize(history []openai.ChatCompletionMessage) (string, error) {
	summaryPrompt := "以下是用户与助手之间的对话，请总结用户的提问意图和助手的关键回答，简洁准确，不要遗漏重要信息：\n\n"
	for _, msg := range history {
		summaryPrompt += fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content)
	}
	return c.CallOpenAI([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: summaryPrompt},
	})
}

func (c *ChatClient) CallOpenAI(messages []openai.ChatCompletionMessage) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Temperature: 0.7, // 控制回答的随机性，范围是 0 到 2（默认 1）
		MaxTokens:   512, // 限制返回的最大 token 数
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("未从API接收到任何响应")
	}

	return resp.Choices[0].Message.Content, nil
}

package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
	"strings"
)

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	})
	if err != nil {
		log.Fatalf("创建失败：%v", err)
	}

	// 对话历史
	message := []*schema.Message{
		schema.SystemMessage("你是一个友好的 AI 助手"),
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("开始对话，输入 'exit' 以结束对话：")

	for {
		fmt.Print("你: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "exit" {
			fmt.Println("对话结束。")
			break
		}
		if userInput == "" {
			continue
		}

		// 添加用户消息到对话历史
		message = append(message, schema.UserMessage(userInput))

		// 生成响应
		response, err := chatModel.Generate(ctx, message)
		if err != nil {
			log.Fatalf("生成失败：%v", err)
			continue
		}

		// 添加模型响应到对话历史
		message = append(message, response)

		fmt.Printf("\\nAI: %s\\n", response.Content)
	}
}

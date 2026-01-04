package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 1. 创建 ChatTemplate
	template := prompt.FromMessages(
		schema.FString, // 使用字符串格式化的系统消息
		schema.SystemMessage("你是一个{role}"),
		schema.UserMessage("{question}"),
	)

	// 2. 定义变量
	variables := map[string]any{
		"role":     "专业的 Go 语言工程师",
		"question": "请解释 Go 语言中的 goroutine 是什么？",
	}

	// 3. 格式化消息
	message, err := template.Format(ctx, variables)
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}

	// 4. 查看生成消息
	fmt.Printf("生成的消息:\\n")
	for i, msg := range message {
		fmt.Printf("%d: [%s] %s\\n", i, msg.Role, msg.Content)
	}

	// 5. 这里可以将生成的消息传递给 ChatModel 进行对话生成
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 6. 生成响应
	response, err := chatModel.Generate(ctx, message)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	fmt.Printf("\\AI 回答：\\n%s\\n", response.Content)
}

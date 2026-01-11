package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// 1. 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 2. 创建 ChatModelAgent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SimpleAssistant",
		Description: "你是一个乐于助人的 AI 助手，擅长回答各种问题。",
		Instruction: "请根据用户的问题，提供准确且有帮助的回答。",
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{}, // 不使用任何工具
	})
	if err != nil {
		log.Fatalf("创建 ChatModelAgent 失败: %v", err)
	}

	// 3. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	// 4. 运行 Agent
	query := "什么是 RAG？"
	fmt.Printf("=== 用户查询: %s ===\n", query)
	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			// 迭代器已关闭, 退出循环
			break
		}
		if event.Err != nil {
			log.Fatalf("运行 Agent 失败: %v", event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				fmt.Printf("AI 回答: %s\n", msg.Content)
			}
		}
	}
}

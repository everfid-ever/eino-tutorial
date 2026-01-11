package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "StreamableAssistant",
		Description: "一个支持流式输出的智能体，能够实时响应用户输入",
		Instruction: "你是一个高效且响应迅速的助手。请根据用户的输入，提供简洁明了的回答，并尽可能以流式方式输出结果。",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 ChatModelAgent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true, // 启用流式输出
	})

	query := "请简要介绍一下人工智能的发展历程。"
	fmt.Printf("用户输入: %s\n", query)

	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatalf("运行 Agent 失败: %v", event.Err)
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			// 实时输出流式响应内容
			if event.Output.MessageOutput.IsStreaming {
				stream := event.Output.MessageOutput.MessageStream
				for {
					msg, err := stream.Recv()
					if err != nil {
						if err == io.EOF {
							break
						}
						log.Fatalf("接收流式消息失败: %v", err)
					}
					if msg != nil && msg.Content != "" {
						fmt.Print(msg.Content) // 实时打印流式内容
					}
				}
			} else {
				msg := event.Output.MessageOutput.Message
				if msg != nil {
					fmt.Printf("Agent 回复: %s\n", msg.Content)
				}
			}
		}
	}
}

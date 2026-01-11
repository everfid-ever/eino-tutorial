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

	// 2. 创建两个子 Agent
	// Agent 1: 主任务解决 Agent
	mainAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "MainAgent",
		Description: "负责生成初步解决方案",
		Instruction: "你是一个专业的解决方案生成器。请根据用户的输入，生成一个初步的解决方案。",
		Model:       chatModel,
		OutputKey:   "solution", // 输出结果存储在 session 的 "main_solution" 键中
	})
	if err != nil {
		log.Fatalf("创建 MainAgent 失败: %v", err)
	}

	// Agent 2: 批判反馈 Agent
	critiqueAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CritiqueAgent",
		Description: "负责对初步解决方案进行批判性反馈",
		Instruction: "你是一个专业的批判性思考者。请对提供的初步解决方案进行评估，并提出改进建议。可使用 {solution} 变量获取初步解决方案。",
		Model:       chatModel,
		OutputKey:   "critique", // 输出结果存储在 session 的 "critique" 键中
	})
	if err != nil {
		log.Fatalf("创建 CritiqueAgent 失败: %v", err)
	}

	// 3. 创建 Looping Agent
	loopAgent, err := adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
		Name:          "ReflectionAgent",
		Description:   "迭代反思智能体, 通过不断改进解决方案以满足用户需求",
		SubAgents:     []adk.Agent{mainAgent, critiqueAgent},
		MaxIterations: 5, // 最多迭代 5 次
	})
	if err != nil {
		log.Fatalf("创建 LoopAgent 失败: %v", err)
	}

	// 4. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           loopAgent,
		EnableStreaming: false,
	})

	// 5. 运行 Agent 工作流
	query := "我需要一个电商网站的开发方案，要求支持多语言和移动端适配。"
	fmt.Printf("用户输入: %s\n", query)

	iter := runner.Query(ctx, query)
	iteration := 0
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatalf("运行 Agent 失败: %v", event.Err)
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				if event.AgentName == "MainAgent" {
					iteration++
					fmt.Printf("=== 迭代 %d: 初步解决方案 ===\n%s\n\n", iteration, msg.Content)
				} else if event.AgentName == "CritiqueAgent" {
					fmt.Printf("=== 迭代 %d: 批判性反馈 ===\n%s\n\n", iteration, msg.Content)
				}
				fmt.Printf("[%s] 回复: %s\n", event.AgentName, msg.Content)
			}
		}
	}
}

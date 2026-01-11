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

	// 2. 创建多个子 Agent
	// Agent 1: 分析用户需求
	analyzerAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Analyzer",
		Description: "分析用户需求, 提取关键信息",
		Instruction: "你是一个专业的需求分析师。请根据用户的输入，提取出关键信息和需求。",
		Model:       chatModel,
		OutputKey:   "analysis", // 输出结果存储在 session 的 "analysis" 键中
	})
	if err != nil {
		log.Fatalf("创建 Analyzer Agent 失败: %v", err)
	}

	// Agent 2: 根据分析结果生成解决方案
	solutionAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SolutionGenerator",
		Description: "根据分析结果生成解决方案",
		Instruction: "你是一个专业的解决方案生成器。请根据提供的分析结果，生成一个详细的解决方案。可翼使用 {analysis} 变量获取分析结果。",
		Model:       chatModel,
		OutputKey:   "solution", // 输出结果存储在 session 的 "solution" 键中
	})
	if err != nil {
		log.Fatalf("创建 SolutionGenerator Agent 失败: %v", err)
	}

	// 3. 创建 Sequential Agent
	sequentialAgent, err := adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        "AnalysisWorkflow",
		Description: "一个顺序执行多个子 Agent 的工作流 Agent，先进行需求分析再生成解决方案",
		SubAgents: []adk.Agent{
			analyzerAgent,
			solutionAgent,
		},
	})
	if err != nil {
		log.Fatalf("创建 Sequential Agent 失败: %v", err)
	}

	// 4. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           sequentialAgent,
		EnableStreaming: false,
	})

	// 5. 运行 Agent 工作流
	query := "我需要一个电商网站的开发方案，要求支持多语言和移动端适配。"
	fmt.Printf("用户输入: %s\\n", query)

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
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				fmt.Printf("Agent 回复: %s\\n", msg.Content)
			}
		}
	}
}

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
	// Agent 1: 技术调研
	techAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TechResearchAgent",
		Description: "负责进行技术调研，收集相关信息",
		Instruction: "你是一个专业的技术调研员。请根据用户的输入，收集并总结相关的技术信息。",
		Model:       chatModel,
		OutputKey:   "tech_research", // 输出结果存储在 session 的 "tech_research" 键中
	})
	if err != nil {
		log.Fatalf("创建 TechResearchAgent 失败: %v", err)
	}

	// Agent 2: 市场分析
	marketAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "MarketAnalysisAgent",
		Description: "负责进行市场分析，评估市场需求",
		Instruction: "你是一个专业的市场分析师。请根据用户的输入，评估并总结市场需求。",
		Model:       chatModel,
		OutputKey:   "market_analysis", // 输出结果存储在 session 的 "market_analysis" 键中
	})
	if err != nil {
		log.Fatalf("创建 MarketAnalysisAgent 失败: %v", err)
	}

	// Agent 3: 风险评估
	riskAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RiskAssessmentAgent",
		Description: "负责进行风险评估，识别潜在风险",
		Instruction: "你是一个专业的风险评估师。请根据用户的输入，识别并总结潜在的风险。",
		Model:       chatModel,
		OutputKey:   "risk_assessment", // 输出结果存储在 session 的 "risk_assessment" 键中
	})
	if err != nil {
		log.Fatalf("创建 RiskAssessmentAgent 失败: %v", err)
	}

	// 3. 创建 Parallel Agent
	parallelAgent, err := adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        "DataCollectionAgent",
		Description: "并发信息收集 Agent, 同时进行技术调研、市场分析和风险评估",
		SubAgents:   []adk.Agent{techAgent, marketAgent, riskAgent},
	})
	if err != nil {
		log.Fatalf("创建 ParallelAgent 失败: %v", err)
	}

	// 4. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           parallelAgent,
		EnableStreaming: false,
	})

	// 5. 运行 Agent 工作流
	query := "请帮我收集关于新兴技术的相关信息，并评估其市场需求和潜在风险。"
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
			msg := event.Output.MessageOutput.Message
			if msg != nil {
				fmt.Printf("[%s] 回复: %s\n", event.AgentName, msg.Content)
			}
		}
	}
}

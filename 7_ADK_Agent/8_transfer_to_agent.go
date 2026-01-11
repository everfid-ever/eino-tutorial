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

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	generalAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "GeneralAgent",
		Description: "通用智能体, 可以处理各种问题, 也可以将任务转移给专业的 Agent",
		Instruction: "你是一个通用助手。你可以: 1. 直接回答简单的问题; 2. 将复杂的技术问题转移给 TechExpert; 3. 将数学问题转移给 MathExpert",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 GeneralAgent 失败: %v", err)
	}

	TechExpert, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TechExpert",
		Description: "技术专家智能体, 专门处理复杂的技术问题",
		Instruction: "你是一个技术专家。请详细解答用户的技术问题。",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 TechExpert 失败: %v", err)
	}

	MathExpert, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "MathExpert",
		Description: "数学专家智能体, 专门处理数学问题",
		Instruction: "你是一个数学专家。请详细解答用户的数学问题。",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 MathExpert 失败: %v", err)
	}

	// 设置 Agent 关系
	generalAgentWithSubs, err := adk.SetSubAgents(ctx, generalAgent, []adk.Agent{TechExpert, MathExpert})
	if err != nil {
		log.Fatalf("为 GeneralAgent 设置子 Agent 失败: %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           generalAgentWithSubs,
		EnableStreaming: false,
	})

	queries := []string{
		"你好, 今天的天气怎么样?",
		"请解释一下区块链技术的基本原理。",
		"数学中的毕达哥拉斯定理是什么?",
	}

	for _, query := range queries {
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
					// 检查是否有工具调用
					if len(msg.ToolCalls) > 0 {
						for _, tc := range msg.ToolCalls {
							fmt.Printf("Agent 使用工具: %s, 参数: %v\n", event.AgentName, tc.Function.Arguments)
						}
					}
				} else if msg != nil {
					fmt.Printf("[%s] 回复: %s\n", event.AgentName, msg.Content)
				}
			}
		}
	}
}

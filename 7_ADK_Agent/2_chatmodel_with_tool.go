package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
	"time"
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

	// 2. 创建工具
	// 获取当前时间的工具
	timeTool := utils.NewTool(
		&schema.ToolInfo{
			Name:        "get_current_time",
			Desc:        "获取当前时间",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
	)

	// 计算工具
	calculatorTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "calculator",
			Desc: "一个简单的计算器，支持加减乘除运算",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"expression": {
					Desc:     "数学表达式，例如 '23 + 7'",
					Required: true,
					Type:     schema.String,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			_, ok := params["expression"].(string)
			if !ok {
				return "", fmt.Errorf("参数 expression 类型错误")
			}
			// 简化实现
			return "30", nil
		},
	)

	// 3. 创建 ChatModel Agent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ToolAssistant",
		Description: "一个可以使用工具的智能助手",
		Model:       chatModel,
		Instruction: `你可以使用以下工具来帮助用户完成任务：
1. get_current_time: 获取当前时间
2. calculator: 进行数学计算`,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{timeTool, calculatorTool},
			},
		},
		MaxIterations: 10, // 最多执行10次工具调用
	})
	if err != nil {
		log.Fatalf("创建 ChatModel Agent 失败: %v", err)
	}

	// 4. 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	// 5. 运行 Agent
	queries := []string{
		"现在的时间是什么？",
		"计算23加7的结果。",
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
					fmt.Printf("Agent 回复: %s\n", msg.Content)
				}
			}
		}
	}
}

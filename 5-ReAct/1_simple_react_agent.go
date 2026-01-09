package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
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

	// 2，创建工具
	// 获取当前时间工具
	timeTool := utils.NewTool(
		&schema.ToolInfo{
			Name:        "get_current_time",
			Desc:        "获取当前时间",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			now := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("工具被调用，返回当前时间: %s\n", now)
			return now, nil
		},
	)

	// 简单计算工具
	calculatorTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "calculator",
			Desc: "执行简单的数学计算 (加减乘除)",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"expression": {
					Type:     "string",
					Desc:     "数学表达式，例如: 2 + 2",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			expr := params["expression"].(string)
			// 简化实现: 只处理 "2 + 2"
			var result string
			if expr == "2 + 2" {
				result = "4"
			} else {
				result = "未知表达式"
			}
			fmt.Printf("工具被调用，计算表达式 %s，结果: %s\n", expr, result)
			return result, nil
		},
	)

	// 3. 创建 ReAct Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{timeTool, calculatorTool},
		},
	})
	if err != nil {
		log.Fatalf("创建 ReAct Agent 失败: %v", err)
	}

	// 4. 使用 Agent 处理用户请求
	message := []*schema.Message{
		schema.UserMessage("请告诉我当前时间"),
	}

	fmt.Println("=== 用户: 请告诉我当前时间 ===")

	response, err := agent.Generate(ctx, message)
	if err != nil {
		log.Fatalf("Agent 生成响应失败: %v", err)
	}

	fmt.Printf("=== Agent 响应: %s ===\n", response.Content)

}

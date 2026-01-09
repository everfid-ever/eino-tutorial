package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
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

	// 1. 创建多个工具
	// 计算工具
	calculaor := utils.NewTool(
		&schema.ToolInfo{
			Name: "calculator",
			Desc: "一个简单的计算器，支持加减乘除运算。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"operation": {
					Type:     "string",
					Desc:     "运算类型: add(加), subtract(减), multiply(乘), divide(除)",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			// 简化实现: 只处理简单加法
			return "30", nil
		},
	)

	// 时间工具
	timeTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "get_current_time",
			Desc: "获取当前时间，支持不同格式（date, time, datetime）。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"format": {
					Type:     "string",
					Desc:     "时间格式: date(日期), time(时间), datetime(完整时间)",
					Required: false,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			// 简化实现: 返回固定时间
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
	)

	// 2. 创建 ChatModel (支持 Function Calling)
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("CHAT_MODEL_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 3. 创建 ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{calculaor, timeTool},
	})
	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}

	// 4. 获取工具信息列表
	calcInfo, err := calculaor.Info(ctx)
	if err != nil {
		log.Fatalf("获取计算器工具信息失败: %v", err)
	}
	timeInfo, err := timeTool.Info(ctx)
	if err != nil {
		log.Fatalf("获取时间工具信息失败: %v", err)
	}

	toolsInfo := []*schema.ToolInfo{calcInfo, timeInfo}

	// 测试多个场景
	testCases := []string{
		"现在时间是多少？",
		"计算15加15等于多少？",
		"请告诉我当前的日期和时间，并计算5乘以6的结果。",
	}

	for i, question := range testCases {
		fmt.Printf("=== 测试用例 %d ===\n", i+1)
		fmt.Printf("\t%s\n", question)

		message := []*schema.Message{
			schema.UserMessage(question),
		}

		// AI 调用工具
		response, err := chatModel.Generate(ctx, message, model.WithTools(toolsInfo))
		if err != nil {
			log.Fatalf("生成失败: %v", err)
		}

		fmt.Printf("AI 回答：\\n")
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				fmt.Printf("使用工具: %s\\n", toolCall.Function.Name)
				fmt.Printf(" 参数: %v\\n", toolCall.Function.Arguments)

				// 通过 ToolsNode 执行工具调用
				toolResult, err := toolsNode.Invoke(ctx, response)
				if err != nil {
					log.Printf("工具调用失败: %v", err)
					continue
				}
				fmt.Printf(" 工具结果: %s\\n", toolResult)
			}
		} else {
			fmt.Printf("%s\\n", response.Content)
		}
	}
}

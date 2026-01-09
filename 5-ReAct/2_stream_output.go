package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// 创建模型和工具
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	searchTool := utils.NewTool(&schema.ToolInfo{
		Name: "search_tool",
		Desc: "搜索信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索关键词",
				Required: true,
			},
		}),
	},
		func(ctx context.Context, params map[string]any) (string, error) {
			query := params["query"].(string)
			// 模拟搜索结果
			result := "搜索结果: 关于 '" + query + "' 的信息。"
			fmt.Printf("工具被调用，返回结果: %s\n", query)
			return result, nil
		},
	)

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{searchTool},
		},
	})
	if err != nil {
		log.Fatalf("创建 ReAct Agent 失败: %v", err)
	}

	// 流式调用
	messages := []*schema.Message{
		schema.UserMessage("请帮我搜索关于人工智能的最新发展。"),
	}

	fmt.Println("用户: 请帮我搜索关于人工智能的最新发展。")
	fmt.Print("AI: ")

	stream, err := agent.Stream(ctx, messages)
	if err != nil {
		log.Fatalf("流式调用失败: %v", err)
	}
	defer stream.Close()

	// 逐块接受并打印
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("接收失败: %v", err)
		}

		// 打印接收到的内容块
		fmt.Print(chunk.Content)
	}

	fmt.Println("\\n\n === 完成 ===")
}

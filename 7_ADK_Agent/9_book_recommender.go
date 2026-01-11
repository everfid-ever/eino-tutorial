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
	"sync"
)

var _ compose.CheckPointStore = (*memoryCheckPointStore)(nil)

type memoryCheckPointStore1 struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemoryCheckPointStore1() *memoryCheckPointStore {
	return &memoryCheckPointStore{
		data: make(map[string][]byte),
	}
}

func (s *memoryCheckPointStore1) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.data[checkPointID]
	return data, exists, nil
}

func (s *memoryCheckPointStore1) Set(ctx context.Context, checkPointID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[checkPointID] = data
	return nil
}

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

	bookSearchTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "search_book",
			Desc: "搜索书籍信息",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"keword": {
					Desc:     "搜索关键词",
					Required: true,
					Type:     schema.String,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			keword := params["keword"].(string)
			// 模拟书籍搜索逻辑
			books := map[string][]string{
				"go":     {"《Go语言圣经》", "《Go并发编程实战》"},
				"python": {"《Python编程：从入门到实践》", "《流畅的Python》"},
			}
			if result, exists := books[keword]; exists {
				return fmt.Sprintf("找到以下书籍: %v", result), nil
			}
			return fmt.Sprintf("未找到与关键词 '%s' 相关的书籍", keword), nil
		},
	)

	askTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "ask_for_clarification",
			Desc: "向用户询问更多信息以澄清需求",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"question": {
					Desc:     "澄清问题",
					Required: true,
					Type:     schema.String,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			question := params["question"].(string)
			return fmt.Sprintf("用户需要澄清: %s", question), nil
		},
	)

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "BookRecommender",
		Description: "书籍推荐智能体, 能够搜索书籍并询问用户偏好",
		Instruction: "你是一个专业的书籍推荐专家。你可以使用以下工具来帮助用户找到合适的书籍：\n" +
			"1. search_book: 用于根据关键词搜索书籍信息。\n" +
			"2. ask_for_clarification: 用于向用户询问更多信息以澄清需求。",
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{bookSearchTool, askTool},
			},
		},
		Exit:          adk.ExitTool{},
		MaxIterations: 10,
	})
	if err != nil {
		log.Fatalf("创建 BookRecommender Agent 失败: %v", err)
	}

	store := newMemoryCheckPointStore1()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
		CheckPointStore: store,
	})

	checkPointID := "session_001"
	query := "我喜欢科幻和冒险类的书籍，能推荐几本吗？"
	fmt.Printf("用户输入: %s\n", query)
	iter := runner.Run(ctx, []adk.Message{
		schema.UserMessage(query),
	}, adk.WithCheckPointID(checkPointID))
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
				// 检查工具是否被调用
				if len(msg.ToolCalls) > 0 {
					for _, tc := range msg.ToolCalls {
						fmt.Printf("工具调用: %s", tc.Function.Name)
					}
				} else if msg.Content != "" {
					fmt.Printf("Agent 回复: %s\n", msg.Content)
				}
			}
		}
	}
}

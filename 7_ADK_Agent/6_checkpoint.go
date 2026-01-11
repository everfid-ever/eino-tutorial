package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
	"sync"
)

var _ compose.CheckPointStore = (*memoryCheckPointStore)(nil)

type memoryCheckPointStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemoryCheckPointStore() *memoryCheckPointStore {
	return &memoryCheckPointStore{
		data: make(map[string][]byte),
	}
}

func (s *memoryCheckPointStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.data[checkPointID]
	return data, exists, nil
}

func (s *memoryCheckPointStore) Set(ctx context.Context, checkPointID string, data []byte) error {
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

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "BookRecommender",
		Description: "根据用户的兴趣推荐书籍",
		Instruction: "你是一个专业的书籍推荐专家。请根据用户的兴趣，推荐适合的书籍。",
		Model:       chatModel,
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	store := newMemoryCheckPointStore()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		CheckPointStore: store,
		EnableStreaming: false,
	})

	checkPointID := "session_001"
	query := "我喜欢科幻和冒险类的书籍，能推荐几本吗？"
	fmt.Printf("用户输入: %s\n", query)

	iter := runner.Run(ctx, []adk.Message{
		schema.UserMessage(query),
	}, adk.WithCheckPointID(checkPointID))

	// 处理事件...
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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func generateWithRetry(ctx context.Context, chatModel *deepseek.ChatModel, message []*schema.Message, maxRetries int) (*schema.Message, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		response, err := chatModel.Generate(ctx, message)
		if err == nil {
			return response, nil
		}

		lastErr = err
		log.Printf("尝试第 %d (%d)次失败: %v", i+1, maxRetries, err)

		// 指数退避等待
		if i < maxRetries-1 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			log.Printf("等待 %v 后重试...", backoff)
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("所有 %d 次尝试均失败: %v", maxRetries, lastErr)
}

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
		// 设置超时
		Timeout: 30 * time.Second,
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	message := []*schema.Message{
		schema.UserMessage("你好"),
	}

	// 使用重试机制生成响应
	response, err := generateWithRetry(ctx, chatModel, message, 5)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("请求超时: %v", err)
		}
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("成功！AI 响应: %s\n", response.Content)
}

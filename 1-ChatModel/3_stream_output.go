package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
	"io"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	})
	if err != nil {
		log.Fatalf("创建失败：%v", err)
	}

	message := []*schema.Message{
		schema.SystemMessage("你是一个专业的 Go 语言工程师"),
		schema.UserMessage("请写一篇关于 Go 并发编程的短文（200字左右）？"),
	}

	// 流式生成响应
	stream, err := chatModel.Stream(ctx, message)
	if err != nil {
		log.Fatalf("流式生成失败：%v", err)
	}
	defer stream.Close()

	fmt.Print("AI 回答: ")

	// 逐步接收流式响应
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// 流结束
				break
			}
			log.Fatalf("接收流式响应失败：%v", err)
		}

		// 输出接收到的内容块
		fmt.Print(chunk.Content)
	}

	fmt.Println("\\n\n完成！")
}

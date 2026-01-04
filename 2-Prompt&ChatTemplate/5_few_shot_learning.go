package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// 创建 Few-Shot 模板
	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个情感分析助手。请分析文本的情感倾向，返回格式：情感：[正面/负面/中性] | 置信度：[0-100]。"),

		// 提供示例
		schema.UserMessage("这个产品非常好，我很喜欢！"),
		schema.AssistantMessage("情感：正面 | 置信度：95", nil),

		schema.UserMessage("服务态度差，体验很糟糕。"),
		schema.AssistantMessage("情感：负面 | 置信度：90", nil),

		schema.UserMessage("质量一般，没有特别出色的地方。"),
		schema.AssistantMessage("情感：中性 | 置信度：80", nil),

		// 实际分析文本
		schema.UserMessage("{text}"),
	)

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		Model:       "deepseek-chat",
		BaseURL:     "https://api.deepseek.com",
		Temperature: 0.8,
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 要分析的文本
	testTexts := []string{
		"这个框架的文档写的很详细，上手很快",
		"我对这个应用感到非常失望，功能太少了",
		"Bug 修复后，性能有所提升，但仍有改进空间",
	}

	for _, text := range testTexts {
		messages, err := template.Format(ctx, map[string]any{
			"text": text,
		})
		if err != nil {
			log.Fatalf("格式化失败: %v", err)
		}
		response, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Fatalf("生成失败: %v", err)
		}
		fmt.Printf("文本: %s\\n分析结果: %s\\n\\n", text, response.Content)
	}
}

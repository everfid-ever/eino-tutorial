package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	// 1. 创建 ChatTemplate
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个{role}"),
		schema.UserMessage("{question}"),
	)

	// 2. 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 3. 创建 Chain: ChatTemplate + ChatModel
	// 输入类型: map[string]any ->  输出类型: *schema.Message
	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.
		AppendChatTemplate(chatTemplate). // 第一步: 使用 ChatTemplate 格式化消息
		AppendChatModel(chatModel)        // 第二步: 使用 ChatModel 生成响应

	// 4. 编译 Chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	// 5. 运行 Chain
	input := map[string]any{
		"role":     "专业的 Go 语言工程师",
		"question": "请解释 Go 语言中的 goroutine 是什么？",
	}

	output, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	// 6. 查看 AI 回答
	log.Printf("AI 回答: %s", output.Content)

}

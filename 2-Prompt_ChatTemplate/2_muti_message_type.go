package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 支持多种消息角色
	template := prompt.FromMessages(
		schema.FString,
		// System：系统提示，定义 AI 的角色行为
		schema.SystemMessage("你是{role}, 你的专长是{expertise}"),

		// User：用户消息
		schema.UserMessage("我的问题是：{question}"),

		// Assistant：AI 的历史回复（用于多轮对话）
		schema.AssistantMessage("我理解了，让我思考一下...", nil),

		// User：继续对话
		schema.UserMessage("请详细说明"),
	)

	variable := map[string]any{
		"role":      "一个专业的 AI 助手",
		"expertise": "自然语言处理和机器学习",
		"question":  "什么是深度学习？",
	}

	// 格式化消息
	messages, err := template.Format(ctx, variable)
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}

	for i, msg := range messages {
		fmt.Printf("%d: [%s] %s\\n\n", i, msg.Role, msg.Content)
	}

}

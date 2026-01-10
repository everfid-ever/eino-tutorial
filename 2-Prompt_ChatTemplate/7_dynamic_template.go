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

type ConversationStyle string

const (
	StyleProfessional ConversationStyle = "professional"
	StyleCasual       ConversationStyle = "casual"
	StyleFriendly     ConversationStyle = "friendly"
	StyleFormal       ConversationStyle = "formal"
)

func createDynamicTemplate(style ConversationStyle, domain string) prompt.ChatTemplate {
	var systemMessage string

	switch style {
	case StyleProfessional:
		systemMessage = fmt.Sprintf("你是一个专业且高效的%s专家，提供准确且简洁的信息。", domain)
	case StyleCasual:
		systemMessage = fmt.Sprintf("你是一个随和且易于接近的%s专家，喜欢用轻松的语气与用户交流。", domain)
	case StyleFriendly:
		systemMessage = fmt.Sprintf("你是一个友好且乐于助人的%s专家，喜欢用温暖和鼓励的语气与用户交流。", domain)
	case StyleFormal:
		systemMessage = fmt.Sprintf("你是一个正式且尊重礼仪的%s专家，使用专业且恰当的语言与用户交流。", domain)
	default:
		systemMessage = fmt.Sprintf("你是一个%s专家，提供信息和帮助。", domain)
	}
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(systemMessage),
		schema.UserMessage("{query}"),
	)
}

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 示例：使用不同的对话风格和领域
	styles := []ConversationStyle{StyleProfessional, StyleCasual, StyleFriendly, StyleFormal}
	domain := "医疗健康"
	query := "请解释一下高血压的预防措施。"

	for _, style := range styles {
		template := createDynamicTemplate(style, domain)
		messages, err := template.Format(ctx, map[string]any{
			"query": query,
		})
		if err != nil {
			log.Fatalf("格式化失败: %v", err)
		}
		response, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Fatalf("生成失败: %v", err)
		}
		fmt.Printf("风格：%s\\nAI 回答：\\n%s\\n\\n", style, response.Content)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

type Translator struct {
	chatModel *deepseek.ChatModel
}

func NewTranslator(apiKey string) (*Translator, error) {
	ctx := context.Background()
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:      apiKey,
		Model:       "deepseek-chat",
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}
	return &Translator{chatModel: chatModel}, nil
}

func (t *Translator) Translate(ctx context.Context, text, targetLang string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf("你是一个专业的翻译助手。请将用户输入的文本翻译成%s，只返回翻译结果，不要添加任何解释。", targetLang)),
		schema.UserMessage(text),
	}

	response, err := t.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func main() {
	translator, err := NewTranslator(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		log.Fatalf("创建翻译器失败: %v", err)
	}

	// 测试翻译
	texts := []struct {
		content string
		target  string
	}{
		{"Hello, how are you?", "中文"},
		{"Eino 是一个强大的 AI 开发框架", "English"},
		{"Les roses sont rouges", "中文"},
	}

	for _, item := range texts {
		result, err := translator.Translate(context.Background(), item.content, item.target)
		if err != nil {
			log.Printf("翻译失败: %v", err)
			continue
		}
		fmt.Printf("原文: %s\\n翻译: %s\\n\\n", item.content, result)
	}
}

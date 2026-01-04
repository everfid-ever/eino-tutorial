package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"log"
)

type CodeAssistant struct {
	chatModel *deepseek.ChatModel
}

func NewCodeAssistant(apiKey string) (*CodeAssistant, error) {
	ctx := context.Background()
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  apiKey,
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}
	return &CodeAssistant{
		chatModel: chatModel,
	}, nil
}

func (ca *CodeAssistant) ExplainCode(ctx context.Context, code, language string) (string, error) {
	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个专业的代码助手，擅长解释各种编程语言的代码片段。"),
		schema.UserMessage("请解释以下{language}代码的功能和作用：\\n\n```{language}\\n{code}\\n```"),
	)
	messages, err := template.Format(ctx, map[string]any{
		"language": language,
		"code":     code,
	})
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}
	response, err := ca.chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	return response.Content, nil
}

func (ca *CodeAssistant) OptimizeCode(ctx context.Context, code, language string) (string, error) {
	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个专业的代码助手，擅长优化各种编程语言的代码片段。请你从以下方面进行优化："+
			"1. 提高代码性能；"+
			"2. 增强代码可读性；"+
			"3. 遵循最佳实践。"+
			"4. 错误处理和边界情况。"),
		schema.UserMessage("请优化以下{language}代码，提高其性能和可读性：\\n\n```{language}\\n{code}\\n```"),
	)
	messages, err := template.Format(ctx, map[string]any{
		"language": language,
		"code":     code,
	})
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}
	response, err := ca.chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	return response.Content, nil

}

func main() {
	assistant, err := NewCodeAssistant("YOUR_DEEPSEEK_API_KEY")
	if err != nil {
		log.Fatalf("初始化代码助手失败: %v", err)
	}

	code := `func getData() []int {
	data := []int{}
	for i := 0; i < 1000; i++ {
		data = append(data, i)
	}
	return data
}`

	// 解释代码
	fmt.Println("=== 代码解释 ===")
	explanation, err := assistant.ExplainCode(context.Background(), code, "Go")
	if err != nil {
		log.Fatalf("代码解释失败: %v", err)
	}
	fmt.Printf("代码解释：\\n%s\\n", explanation)

	// 优化代码
	fmt.Println("=== 代码优化 ===")
	optimizedCode, err := assistant.OptimizeCode(context.Background(), code, "Go")
	if err != nil {
		log.Fatalf("代码优化失败: %v", err)
	}
	fmt.Printf("优化后的代码：\\n%s\\n", optimizedCode)

	// 解释优化后的代码
	fmt.Println("=== 优化后代码解释 ===")
	explanationOptimized, err := assistant.ExplainCode(context.Background(), optimizedCode, "Go")
	if err != nil {
		log.Fatalf("优化后代码解释失败: %v", err)
	}
	fmt.Printf("优化后代码解释：\\n%s\\n", explanationOptimized)
}

/*
	本章节学习了：
	1. ChatTemplate 的创建与使用
	2. 变量的定义与格式化
	3. 模板复用和管理策略
	4. Few-Shot 学习与 Chain of Thought (CoT) 技巧
	5. 动态模板生成
	6. 提示词工程最佳实践
	7. 构建代码助手应用
*/

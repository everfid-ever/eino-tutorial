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

// PromptTemplate 提示词管理模板
type PromptTemplate struct{}

func (p *PromptTemplate) Translator(sourceLang, targetLang string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一个专业的翻译助手。请将%s翻译成%s。\\n"+
				"要求：\\n"+
				"1. 保持原文的语义和风格。\\n"+
				"2. 使用地道的表达方式。\\n"+
				"3. 只返回结果，不要添加解释。",
			sourceLang, targetLang,
		)),
		schema.UserMessage("{text}"),
	)
}

// 代码审核模板
func (p *PromptTemplate) CodeReview(language string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一个专业的%s开发专家。请审核以下代码。\\n"+
				"要求：\\n"+
				"1. 检查代码的正确性和效率。\\n"+
				"2. 提出改进建议。\\n"+
				"3. 只返回结果，不要添加解释。",
			language,
		)),
		schema.UserMessage("请审核以下代码: \\n\n```{language}\\n{code}\\n```"),
	)
}

// 技术面试官模板
func (p *PromptTemplate) TechInterview(position, level string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一个%s职位的技术面试官，负责面试%s级别职位。\\n"+
				"要求：\\n"+
				"1. 提出与职位相关的技术问题。\\n"+
				"2. 根据回答进行深入追问。\\n"+
				"3. 只返回问题，不要添加解释。",
			position, level,
		)),
		schema.UserMessage("候选人回答: {answer}\\n\n请评估并追问。"),
	)
}

func main() {
	ctx := context.Background()

	// 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 创建提示词模板管理器
	templates := &PromptTemplate{}

	// 使用翻译模板
	fmt.Println("=== 翻译示例 ===")
	translatorTemplate := templates.Translator("中文", "英文")
	messages, _ := translatorTemplate.Format(ctx, map[string]any{
		"text": "你好，欢迎使用我们的翻译服务！",
	})
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	fmt.Printf("AI 回答：\\n%s\\n", response.Content)

	// 使用代码审核模板
	fmt.Println("=== 代码审核示例 ===")
	codeReviewTemplate := templates.CodeReview("Go")
	messages, _ = codeReviewTemplate.Format(ctx, map[string]any{
		"language": "Go",
		"code":     "package main\\n\\nfunc main() {\\n    println(\"Hello, World!\")\\n}",
	})
	response, err = chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	fmt.Printf("AI 回答：\\n%s\\n", response.Content)

	// 使用技术面试官模板
	fmt.Println("=== 技术面试官示例 ===")
	interviewTemplate := templates.TechInterview("后端开发工程师", "中级")
	messages, _ = interviewTemplate.Format(ctx, map[string]any{
		"answer": "我有3年的Go语言开发经验，熟悉微服务架构。",
	})
	response, err = chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}
	fmt.Printf("AI 回答：\\n%s\\n", response.Content)
}

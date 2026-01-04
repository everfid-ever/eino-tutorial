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

	// Chain of Thought (CoT) 提示词模板
	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个逻辑推理专家。请按照以下步骤回答问题："+
			"1. 理解问题；复述问题的要求"+
			"2. 分析已知信息；列出相关事实"+
			"3. 制定解决方案；描述解决问题的方法"+
			"4. 逐步推理；详细说明每一步的逻辑"+
			"5. 得出结论；给出最终答案"),
		schema.UserMessage("{problem}"),
	)

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 要解决的问题
	problem := "如果一个火车以每小时60英里的速度行驶，另一辆火车以每小时80英里的速度行驶，它们相距240英里。问两辆火车多久会相遇？"

	messages, err := template.Format(ctx, map[string]any{
		"problem": problem,
	})
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("AI 回答：\\n%s\\n", response.Content)
}

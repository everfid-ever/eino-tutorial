package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

func main() {
	cxt := context.Background()

	chatModel, err := deepseek.NewChatModel(cxt, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("CHAT_MODEL_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 创建并行节点
	parallel := compose.NewParallel()

	// 任务1: 提取关键词
	parallel.AddLambda("keyword", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("任务1: 提取关键词")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("请提取文本中的关键词，以逗号分隔。"),
				schema.UserMessage("{text}"),
			)

			message, _ := template.Format(ctx, input)
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			return response.Content, nil
		},
	))

	// 任务2: 情感分析
	parallel.AddLambda("sentiment", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("任务2: 情感分析")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("请对文本进行情感分析，判断其是正面、负面还是中性。"),
				schema.UserMessage("{text}"),
			)
			message, _ := template.Format(ctx, input)
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			return response.Content, nil
		},
	))

	// 任务3: 摘要生成
	parallel.AddLambda("summary", compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (string, error) {
			fmt.Println("任务3: 摘要生成")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("请为以下文本生成一个简短的摘要。"),
				schema.UserMessage("{text}"),
			)
			message, _ := template.Format(ctx, input)
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			return response.Content, nil
		},
	))

	// 创建主链
	chain := compose.NewChain[string, map[string]string]()

	chain.
		// 准备输入
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, text string) (map[string]any, error) {
			return map[string]any{"text": text}, nil
		})).
		// 执行并行任务
		AppendParallel(parallel).
		// 处理结果
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, results map[string]any) (map[string]any, error) {
			fmt.Println("\\n=== 并行任务结果 ===")
			return results, nil
		}))

	runnable, err := chain.Compile(cxt)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	text := `Eino 是一个强大的 AI 开发框架，支持构建复杂的多步骤处理链。通过将数据清洗、格式转换、AI 分析和结果提取等步骤串联在一起，开发者可以轻松实现端到端的 AI 解决方案。`
	result, err := runnable.Invoke(cxt, text)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	fmt.Printf("\\n关键词: %s\\n", result["keyword"])
	fmt.Printf("情感分析: %s\\n", result["sentiment"])
	fmt.Printf("摘要: %s\\n", result["summary"])
}))
}

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

type ArticleRequest struct {
	Topic    string
	Keywords []string
	Length   int // 目标字数
}

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("CHAT_MODEL_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 构建文章生成流水线
	chain := compose.NewChain[ArticleRequest, string]()
	chain.
		// 步骤1: 生成文章大纲
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, req ArticleRequest) (string, error) {
			fmt.Println("=== 步骤1: 生成文章大纲 ===")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("你是一个专业的内容策划师。请根据主题和关键词生成文章大纲。"),
				schema.UserMessage("主题: {topic}\\n关键词: {keywords}\\n\n请生成一个包含主要章节和小节的文章大纲。"),
			)

			message, _ := template.Format(ctx, map[string]any{
				"topic":    req.Topic,
				"keywords": req.Keywords,
			})
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			fmt.Printf("生成的大纲:\\n%s\\n\n", response.Content)

			return response.Content, nil
		})).

		// 步骤2: 扩写内容
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, outline string) (string, error) {
			fmt.Println("=== 步骤2: 扩写内容 ===")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("你是一个专业的内容写作专家。请根据提供的大纲扩写成完整的文章。"),
				schema.UserMessage("大纲: {outline}\\n\n请根据大纲撰写一篇详细的文章，目标字数为800字。"),
			)

			message, _ := template.Format(ctx, map[string]any{
				"outline": outline,
			})
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			fmt.Printf("初稿完成, 字数: %d\\n\n", len(response.Content))

			return response.Content, nil
		})).

		// 步骤3: 修改润色
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, draft string) (string, error) {
			fmt.Println("=== 步骤3: 修改润色 ===")

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("你是一个专业的编辑。请对文章进行修改和润色，使其更流畅易读。"),
				schema.UserMessage("文章初稿: {draft}\\n\n请对文章进行修改和润色。"),
			)

			message, _ := template.Format(ctx, map[string]any{
				"draft": draft,
			})
			response, err := chatModel.Generate(ctx, message)
			if err != nil {
				return "", err
			}
			fmt.Printf("润色完成, 字数: %d\\n\n", len(response.Content))

			return response.Content, nil
		})).

		// 步骤4: 格式化输出
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, article string) (string, error) {
			fmt.Println("=== 步骤4: 格式化输出 ===")

			// 添加 Markdown 格式
			formatted := fmt.Sprintf("# 文章生成结果\\n\\n%s", article)
			return formatted, nil
		}))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}

	request := ArticleRequest{
		Topic:    "人工智能在现代社会的应用",
		Keywords: []string{"人工智能", "机器学习", "自动化", "社会影响"},
		Length:   800,
	}

	// 执行文章生成流水线
	result, err := runnable.Invoke(ctx, request)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	fmt.Printf("=== 最终文章输出 ===\\n%s\\n", result)
}

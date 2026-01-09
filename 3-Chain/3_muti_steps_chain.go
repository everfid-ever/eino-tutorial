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
	"strings"
)

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

	// 构建处理链
	chain := compose.NewChain[string, string]()

	chain.
		// Step 1: 数据清洗
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, rawText string) (string, error) {
			fmt.Printf("=== 步骤一: 数据清洗 ===")
			// 去除多余空格和换行
			cleaned := strings.TrimSpace(rawText)
			cleaned = strings.ReplaceAll(cleaned, "\\n\n", "\\n")
			fmt.Printf("清洗后: %s\\n\n", cleaned)
			return cleaned, nil
		})).

		// Step 2: 转换为 AI 分析输入
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, text string) (map[string]any, error) {
			fmt.Printf("=== 步骤2: 准备分析 ===\\n")
			return map[string]any{
				"text": text,
			}, nil
		})).

		// Step 3: AI 进行分析
		AppendGraph(func() *compose.Chain[map[string]any, *schema.Message] {
			analysisChain := compose.NewChain[map[string]any, *schema.Message]()

			template := prompt.FromMessages(
				schema.FString,
				schema.SystemMessage("你是一个专业的数据分析师。请根据提供的文本进行分析，并给出见解。"),
				schema.UserMessage("请分析以下文本内容：\\n{text}"),
			)

			analysisChain.AppendChatTemplate(template).AppendChatModel(chatModel)
			return analysisChain
		}()).

		// Step 4: 提取 AI 分析结果
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (string, error) {
			fmt.Printf("=== 步骤3: 提取结果 ===\\n")
			return msg.Content, nil
		}))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译 Chain 失败: %v", err)
	}
	rawInput := `
    	Eino 是一个强大的 AI 开发框架，支持构建复杂的多步骤处理链。
		通过将数据清洗、格式转换、AI 分析和结果提取等步骤串联在一起，
		开发者可以轻松实现端到端的 AI 解决方案。
		这种模块化设计不仅提高了代码的可维护性，还增强了系统的灵活性和扩展性。
		无论是处理文本、图像还是其他类型的数据，Eino 都能帮助你高效地构建智能应用。
	`
	result, err := runnable.Invoke(ctx, rawInput)
	if err != nil {
		log.Fatalf("运行 Chain 失败: %v", err)
	}

	fmt.Printf("=== 最终分析结果 ===\\n%s\\n", result)
}

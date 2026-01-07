package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"log"
	"strings"
)

func main() {
	ctx := context.Background()

	// 定义分支定义
	branchCondition := func(ctx context.Context, input map[string]any) (string, error) {
		language := input["language"].(string)
		language = strings.ToLower(language)

		fmt.Printf("检测到的语言: %s\\n", language)

		// 根据语言选择分支
		if language == "go" || language == "golang" {
			return "go_branch", nil
		} else if language == "python" {
			return "python_branch", nil
		}
		return "other_branch", nil
	}

	// Go 分支处理
	goBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("执行 Go 分支处理")
		input["advice"] = "推荐使用 Eino 框架进行 AI 开发"
		input["features"] = []string{"高并发", "并发安全", "类型安全"}
		return input, nil
	})

	// Python 分支处理
	pythonBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("执行 Python 分支处理")
		input["advice"] = "推荐使用 LangChain 进行快速原型开发"
		input["features"] = []string{"易用性", "丰富的生态", "快速迭代"}
		return input, nil
	})

	// 其他语言分支处理
	otherBranch := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("执行其他语言分支处理")
		input["advice"] = "建议学习 Go 或 Python 以利用现有 AI 框架"
		input["features"] = []string{"社区支持", "丰富的资源"}
		return input, nil
	})

	// 创建 Chain
	chain := compose.NewChain[map[string]any, map[string]any]()

	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		fmt.Println("=== 开始处理 ===")
		return input, nil
	})).AppendBranch(compose.NewChainBranch(branchCondition).AddLambda("go_branch", goBranch).AddLambda("python_branch", pythonBranch).AddLambda("other_branch", otherBranch)).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			fmt.Println("=== 处理结束 ===")
			return input, nil
		}))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		panic(fmt.Sprintf("编译 Chain 失败: %v", err))
	}

	// 测试输入
	testCases := []map[string]any{
		{"language": "Go", "task": "构建高性能 AI 应用"},
		{"language": "Python", "task": "快速原型开发 AI 模型"},
		{"language": "Java", "task": "企业级 AI 解决方案"},
	}

	for i, testCase := range testCases {
		fmt.Printf("\\n--- 测试用例 %d ---\\n", i+1)
		result, err := runnable.Invoke(ctx, testCase)
		if err != nil {
			log.Printf("运行 Chain 失败: %v", err)
			continue
		}
		fmt.Printf("建议: %s\\n", result["advice"])
		fmt.Printf("特点: %v\\n", result["features"])
	}
}

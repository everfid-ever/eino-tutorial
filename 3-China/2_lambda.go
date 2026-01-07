package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"strings"
)

func main() {
	ctx := context.Background()

	// 创建一个简单的 Chain：输入字符串 -> 转大写 -> 添加前缀 -> 输出字符串
	chain := compose.NewChain[string, string]()

	chain.
		// Lambda 1：转大写
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
			fmt.Printf("步骤1: 输入 = %s\\n", input)
			result := strings.ToUpper(input)
			fmt.Printf("步骤1: 输出 = %s\\n", result)
			return result, nil
		})).
		// Lambda 2：添加前缀
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
			fmt.Printf("步骤2: 输入 = %s\\n", input)
			result := "处理结果: " + input
			fmt.Printf("步骤2: 输出 = %s\\n", result)
			return result, nil
		}))

	// 编译 Chain
	runnable, err := chain.Compile(ctx)
	if err != nil {
		panic(fmt.Sprintf("编译 Chain 失败: %v", err))
	}

	// 运行 Chain
	input := "hello eino"
	output, err := runnable.Invoke(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("运行 Chain 失败: %v", err))
	}

	fmt.Printf("最终输出: %s\\n", output)
}

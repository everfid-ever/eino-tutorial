package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"time"
)

type MyTool struct{}

type MyParams struct {
	Required string `json:"required"`
	Count    int    `json:"count"`
}

func (t *MyTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// =========================
	// 1. 参数验证（Parameter Validation）
	// =========================

	var params MyParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("invalid arguments json: %w", err)
	}

	if params.Required == "" {
		return "", errors.New("required field must not be empty")
	}

	if params.Count < 0 || params.Count > 100 {
		return "", errors.New("count must be in range [0,100]")
	}

	// =========================
	// 2. 超时控制（Timeout Control）
	// =========================

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// =========================
	// 3. 执行业务 + 错误处理（Execution & Error Handling）
	// =========================

	result, err := someOperation(ctx, params)
	if err != nil {
		errorResult := map[string]string{
			"error":   "操作失败",
			"message": err.Error(),
			"code":    "OPERATION_FAILED",
		}
		errJson, _ := json.Marshal(errorResult)
		if errors.Is(err, context.DeadlineExceeded) {
			return "", errors.New(string(errJson))
		}
		return string(errJson), nil
	}

	// =========================
	// 4. 成功返回（Structured Output）
	// =========================

	resp := struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}{
		Success: true,
		Data:    result,
	}

	out, _ := json.Marshal(resp)
	return string(out), nil
}

func someOperation(ctx context.Context, p MyParams) (string, error) {
	select {
	case <-time.After(time.Duration(p.Count) * time.Millisecond):
		return fmt.Sprintf("processed: %s", p.Required), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

/*
	本章要点总结：
		1. Tool 的概念和接口定义
		2. 两种创建 Tool 的方式
		3. 实现 Tool 的关键步骤和注意事项
		4. ToolsNode 的使用
		5. 最佳实践建议
*/

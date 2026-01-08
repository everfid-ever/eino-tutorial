package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"log"
	"time"
)

type TimeParams struct {
	Format string `json:"format"`
}

type TimeResult struct {
	CurrentTime string `json:"current_time"`
}

func GetCurrentTime(ctx context.Context, params *TimeParams) (*TimeResult, error) {
	now := time.Now()
	var result string
	switch params.Format {
	case "date":
		result = now.Format("2006-01-02")
	case "time":
		result = now.Format("15:04:05")
	default:
		result = now.Format("2006-01-02 15:04:05")
	}
	return &TimeResult{CurrentTime: result}, nil
}

func main() {
	ctx := context.Background()

	// 使用 utils.NewTool 将函数封装成工具
	timeTool := utils.NewTool(&schema.ToolInfo{
		Name: "get_current_time",
		Desc: "获取当前时间，支持不同格式（date, time, datetime）。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"format": {
				Type:     "string",
				Desc:     "时间格式: date(日期), time(时间), datetime(完整时间)",
				Required: false,
			},
		}),
	}, GetCurrentTime)

	// 测试工具
	testFormats := []string{"date", "time", "datetime", ""}
	for _, format := range testFormats {
		params := &TimeParams{Format: format}
		b, _ := json.Marshal(params)
		// 调用工具
		outputJSON, err := timeTool.InvokableRun(ctx, string(b))
		if err != nil {
			log.Printf("调用工具失败: %v", err)
			continue
		}
		fmt.Printf("格式=%s, 输出=%s\n", format, outputJSON)
	}
}

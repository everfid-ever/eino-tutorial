package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"os"
)

// FileReaderTool 文件读取工具
type FileReaderTool struct{}

// Info 返回工具信息
func (f *FileReaderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_reader",
		Desc: "读取指定路径的文件内容。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"file_path": {
				Type:     "string",
				Desc:     "文件路径",
				Required: true,
			},
		}),
	}, nil
}

type FileReaderParams struct {
	FilePath string `json:"file_path"`
}

type FileReaderResult struct {
	Content string `json:"content"`
	Error   error  `json:"err"`
}

// InvokableRun 读取文件内容
func (f *FileReaderTool) InvokableRun(ctx context.Context, argumentsInJSON string, ops ...tool.Option) (string, error) {
	// 1. 解析参数
	var params FileReaderParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("unmarshal json fail: %w", err)
	}

	// 2. 读取文件内容
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		result := FileReaderResult{
			Error: fmt.Errorf("read file fail: %w", err),
		}
		resultBytes, _ := json.Marshal(result)
		return string(resultBytes), nil
	}

	// 3. 返回结果
	result := FileReaderResult{
		Content: string(content),
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result fail: %w", err)
	}
	return string(resultBytes), nil
}

// FileWriterTool 文件写入工具
type FileWriterTool struct{}

// Info 返回工具信息
func (f *FileWriterTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_writer",
		Desc: "将内容写入指定路径的文件。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"file_path": {
				Type:     "string",
				Desc:     "文件路径",
				Required: true,
			},
			"content": {
				Type:     "string",
				Desc:     "要写入文件的内容",
				Required: true,
			},
		}),
	}, nil
}

type FileWriterParams struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

type FileWriterResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// InvokableRun 写入文件内容
func (f *FileWriterTool) InvokableRun(ctx context.Context, argumentsInJSON string, ops ...tool.Option) (string, error) {
	// 1. 解析参数
	var params FileWriterParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("unmarshal json fail: %w", err)
	}

	// 2. 写入文件内容
	err := os.WriteFile(params.FilePath, []byte(params.Content), 0644)
	if err != nil {
		result := FileWriterResult{
			Success: false,
			Message: fmt.Sprintf("write file fail: %v", err),
		}
		resultBytes, _ := json.Marshal(result)
		return string(resultBytes), nil
	}

	// 3. 返回结果
	result := FileWriterResult{
		Success: true,
		Message: "文件写入成功",
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result fail: %w", err)
	}
	return string(resultBytes), nil
}

func main() {
	ctx := context.Background()

	// 测试文件写入
	writer := &FileWriterTool{}
	writeParams := FileWriterParams{
		FilePath: "test.txt",
		Content:  "Hello, Eino!",
	}
	writeParamsJSON, _ := json.Marshal(writeParams)
	writeResult, err := writer.InvokableRun(ctx, string(writeParamsJSON))
	if err != nil {
		fmt.Printf("文件写入失败: %v\n", err)
	} else {
		fmt.Printf("文件写入结果: %s\n", writeResult)
	}

	// 测试文件读取
	reader := &FileReaderTool{}
	readParams := FileReaderParams{
		FilePath: "test.txt",
	}
	readParamsJSON, _ := json.Marshal(readParams)
	readResult, err := reader.InvokableRun(ctx, string(readParamsJSON))
	if err != nil {
		fmt.Printf("文件读取失败: %v\n", err)
	} else {
		fmt.Printf("文件读取结果: %s\n", readResult)
	}

	// 清理测试文件
	os.Remove("test.txt")
}

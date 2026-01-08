package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DatabaseQueryTool 数据库查询工具
type DatabaseQueryTool struct {
	// 模拟用户数据
	user []User
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func NewDatabaseQueryTool() *DatabaseQueryTool {
	return &DatabaseQueryTool{
		user: []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
			{ID: 3, Name: "Charlie", Email: "charlie@example.com", Age: 35},
		},
	}
}

func (t *DatabaseQueryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "database_query",
		Desc: "查询用户数据库，支持按ID或姓名查询用户信息。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type: "integer",
				Desc: "用户ID",
			},
			"name": {
				Type: "string",
				Desc: "用户姓名",
			},
		}),
	}, nil
}

type QueryUserParams struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}

func (t *DatabaseQueryTool) InvokableRun(ctx context.Context, argumentsInJSON string, ops ...tool.Option) (string, error) {
	var params QueryUserParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("unmarshal json fail: %w", err)
	}

	var results []User

	// 根据条件查询用户
	for _, user := range t.user {
		if params.UserID != 0 && user.ID == params.UserID {
			results = append(results, user)
			break
		}
		if params.Name != "" && user.Name == params.Name {
			results = append(results, user)
		}
	}

	if len(results) == 0 {
		resultJSON, _ := json.Marshal(map[string]string{
			"message": "未找到匹配的用户信息。",
		})
		return string(resultJSON), nil
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("marshal result fail: %w", err)
	}
	return string(resultJSON), nil
}

func main() {
	ctx := context.Background()
	dbTool := NewDatabaseQueryTool()

	testCases := []QueryUserParams{
		{UserID: 1},
		{Name: "Bob"},
		{Name: "David"},
	}

	for _, tc := range testCases {
		paramsJSON, _ := json.Marshal(tc)
		result, err := dbTool.InvokableRun(ctx, string(paramsJSON))
		if err != nil {
			fmt.Printf("调用数据库查询工具失败: %v\n", err)
			continue
		}
		fmt.Printf("查询参数: %s, 结果: %s\n", string(paramsJSON), result)
	}
}

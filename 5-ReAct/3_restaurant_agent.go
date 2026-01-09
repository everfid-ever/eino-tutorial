package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

// Restaurant 餐厅数据结构
type Restaurant struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Location string   `json:"location"`
	Cuisine  string   `json:"cuisine"`
	Rating   string   `json:"rating"`
	Tags     []string `json:"tags"`
}

type Dish struct {
	Name  string `json:"name"`
	Price string `json:"price"`
	Desc  string `json:"desc"`
	Spicy bool   `json:"spicy"`
}

// 模拟餐厅数据库
var restaurantDB = []Restaurant{
	{
		ID:       "r1",
		Name:     "川香阁",
		Location: "北京市",
		Cuisine:  "川菜",
		Rating:   "4.8",
		Tags:     []string{"辣, 正宗, 环境好"},
	},
	{
		ID:       "r2",
		Name:     "粤味轩",
		Location: "上海市",
		Cuisine:  "粤菜",
		Rating:   "4.5",
		Tags:     []string{"清淡, 口味正宗, 服务好"},
	},
	{
		ID:       "r3",
		Name:     "湘聚楼",
		Location: "广州市",
		Cuisine:  "湘菜",
		Rating:   "4.6",
		Tags:     []string{"辣, 份量足, 价格实惠"},
	},
}

var dishDB = map[string][]Dish{
	"r1": {
		{Name: "水煮鱼", Price: "68元", Desc: "鲜嫩的鱼片配上麻辣汤底", Spicy: true},
		{Name: "宫保鸡丁", Price: "45元", Desc: "经典川菜，鸡丁配花生米", Spicy: true},
	},
	"r2": {
		{Name: "白切鸡", Price: "58元", Desc: "嫩滑的白切鸡，蘸酱更美味", Spicy: false},
		{Name: "清蒸石斑鱼", Price: "88元", Desc: "鲜美的石斑鱼，清蒸保留原汁原味", Spicy: false},
	},
	"r3": {
		{Name: "剁椒鱼头", Price: "72元", Desc: "鱼头配上剁椒，鲜辣开胃", Spicy: true},
		{Name: "湘西外婆菜", Price: "38元", Desc: "传统湘菜，口味独特", Spicy: false},
	},
}

func main() {
	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 工具1 - 查询餐厅信息
	restaurantTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "query_restaurants",
			Desc: "根据条件查询餐厅信息",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"location": {
					Type:     "string",
					Desc:     "餐厅所在城市",
					Required: false,
				},
				"cuisine": {
					Type:     "string",
					Desc:     "餐厅菜系",
					Required: false,
				},
				"spicy": {
					Type:     "boolean",
					Desc:     "是否要求辣味",
					Required: false,
				},
			}),
		}, func(ctx context.Context, params map[string]any) (string, error) {
			fmt.Printf("\\n[工具执行] query_restaurantss\\n")

			var results []Restaurant
			location, _ := params["location"].(string)
			cuisine, _ := params["cuisine"].(string)
			spicy, _ := params["spicy"].(bool)

			for _, r := range restaurantDB {
				match := true
				if location != "" && r.Location != location {
					match = false
				}
				if cuisine != "" && r.Cuisine != cuisine {
					match = false
				}
				if spicy {
					hasSpicy := false
					for _, tag := range r.Tags {
						if tag == "辣" {
							hasSpicy = true
							break
						}
					}
					if !hasSpicy {
						match = false
					}
				}
				if match {
					results = append(results, r)
				}
			}

			resultJSON, _ := json.Marshal(results)
			return string(resultJSON), nil
		},
	)

	// 工具2 - 查询菜品信息
	dishTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "query_dishes",
			Desc: "查询指定餐厅的菜品信息",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"restaurant_id": {
					Type:     "string",
					Desc:     "餐厅ID",
					Required: true,
				},
			}),
		}, func(ctx context.Context, params map[string]any) (string, error) {
			restaurantID, _ := params["restaurant_id"].(string)
			fmt.Printf("\\n[工具执行] query_dishes(restautant_id = %s)\\n", restaurantID)

			dishes, exists := dishDB[restaurantID]
			if !exists {
				return `{"error": "餐厅不存在"}`, nil
			}

			resultJSON, _ := json.Marshal(dishes)
			return string(resultJSON), nil
		},
	)

	// 创建 ReAct Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{restaurantTool, dishTool},
		},
	})
	if err != nil {
		log.Fatalf("创建 ReAct Agent 失败: %v", err)
	}

	// 使用 Agent 处理用户请求
	message := []*schema.Message{
		schema.UserMessage("我想找一家北京的川菜馆，最好有辣味的菜。"),
	}

	fmt.Printf("\\n用户: 我想找一家北京的川菜馆，最好有辣味的菜。\\n")
	response, err := agent.Generate(ctx, message)
	if err != nil {
		log.Fatalf("Agent 生成响应失败: %v", err)
	}

	fmt.Printf("AI: %s\\n", response.Content)
}

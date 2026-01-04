package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type UserProfile struct {
	Name      string
	Age       int
	Interests []string
	VIPLevel  int
}

func main() {
	ctx := context.Background()

	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个个性化推荐系统"),
		schema.UserMessage("用户信息：姓名：{name}，年龄：{age}，兴趣：{interests}，VIP等级：{vip_level}。请根据这些信息推荐三本书籍。"),
	)

	//准备用户数据
	user := UserProfile{
		Name:      "张伟",
		Age:       28,
		Interests: []string{"科技", "历史", "旅行"},
		VIPLevel:  3,
	}

	variables := map[string]any{
		"name":      user.Name,
		"age":       user.Age,
		"interests": fmt.Sprintf("%v", user.Interests),
		"vip_level": user.VIPLevel,
	}

	message, err := template.Format(ctx, variables)
	if err != nil {
		log.Fatalf("template format err: %v", err)
	}

	for _, msg := range message {
		fmt.Printf("[%s] %s\\n", msg.Role, msg.Content)
	}
}

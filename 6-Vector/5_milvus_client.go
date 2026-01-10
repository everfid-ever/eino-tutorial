package main

import (
	"context"
	"log"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

func main() {
	ctx := context.Background()

	// 创建 Milvus 客户端
	addr := os.Getenv("MILVUS_ADDRESS")
	if addr == "" {
		addr = "127.0.0.1:19530" // 默认地址
	}

	cli, err := client.NewClient(ctx, client.Config{
		Address: addr,
	})
	if err != nil {
		log.Fatalf("连接 Milvus 失败: " + err.Error())
	}
	defer cli.Close()

	log.Println("成功连接到 Milvus!")
}

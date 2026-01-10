package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func main() {
	// 获取环境变量
	addr := os.Getenv("MILVUS_ADDRESS")
	if addr == "" {
		addr = "192.168.179.138.19530" // 默认地址
	}

	// 创建客户端
	ctx := context.Background()
	cli, err := client.NewClient(ctx, client.Config{
		Address: addr,
	})
	if err != nil {
		panic("连接 Milvus 失败: " + err.Error())
	}
	defer cli.Close()

	// 检查集合是否存在
	collectionName := "eino_example"
	hasCollection, err := cli.HasCollection(ctx, collectionName)
	if err != nil {
		panic("检查集合失败: " + err.Error())
	}
	if !hasCollection {
		panic("集合不存在: " + collectionName)
	}
	log.Printf("集合存在: %s, 开始检索", collectionName)

	// 创建 Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ARK Embedding 模型失败: %v", err)
	}

	// 创建 Milvus Retriever
	searchParams, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		log.Fatalf("创建搜索参数失败: %v", err)
	}

	retriever, err := milvusretriever.NewRetriever(ctx, &milvusretriever.RetrieverConfig{
		Client:       cli,
		Collection:   collectionName,
		Embedding:    embedder,
		TopK:         3, // 检索 TopK 个结果
		MetricType:   entity.COSINE,
		OutputFields: []string{"id", "content", "metadata"}, // 输出字段
		Sp:           searchParams,                          // 搜索参数
	})
	if err != nil {
		log.Fatalf("创建 Milvus Retriever 失败: %v", err)
	}

	// 执行检索
	query := "What is EINO?"
	docs, err := retriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("检索失败: %v", err)
	}

	// 输出检索结果
	fmt.Printf("检索到 %d 个文档:\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("文档 %d:\n", i+1)
		fmt.Printf("ID: %s\n", doc.ID)
		fmt.Printf("内容: %s\n", doc.Content)
		fmt.Printf("元数据: %v\n", doc.MetaData)
		fmt.Println("-----")
	}
}

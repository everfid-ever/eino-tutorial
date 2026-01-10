package main

import (
	"context"
	"fmt"

	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
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

	// 删除旧集合 (如果存在), 避免 schema 冲突
	collectionName := "eino_example"
	hasCollection, err := cli.HasCollection(ctx, collectionName)
	if err == nil && hasCollection {
		log.Printf("删除旧集合: %s", collectionName)
		if err := cli.DropCollection(ctx, collectionName); err != nil {
			log.Print("删除集合失败: " + err.Error())
		}
	}

	// 2. 创建 Embedding 模型
	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("MILVUS_API_KEY"),
		Model:  os.Getenv("MILVUS_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ARK Embedding 模型失败: %v", err)
	}

	// 创建 Indexer
	// 注意: 默认配置使用二进制向量 (81920维度， HAMMING 度量), 不适用于 ARK Embedding 的浮点向量
	// 因此需要自定义 Fields 使用浮点向量类型
	// 首先需要获取 Embedding 维度, 通过 Embedding 一个测试文档来检测
	testVector, err := emb.EmbedStrings(ctx, []string{"test"})
	if err != nil {
		log.Fatalf("获取 Embedding 维度失败: %v", err)
	}
	if len(testVector) == 0 || len(testVector[0]) == 0 {
		log.Fatalf("获取 Embedding 维度失败: 返回空向量")
		return
	}
	dim := int64(len(testVector[0]))
	log.Printf("检测到 Embedding 维度: %d", dim)

	// 定义自定义字段: 浮点向量字段
	fields := []*entity.Field{
		entity.NewField().WithName("id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(255).WithIsPrimaryKey(true),
		entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(1024),
		entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(dim),
		entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON),
	}

	// 自定义 DocumentConverter: 将 float64 向量转换为 float32 向量
	// 注意: 必须确保返回的数据格式与 Fields 定义的字段顺序和类型一致
	documentConverter := func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
		rows := make([]interface{}, len(docs))
		for i, doc := range docs {
			// 将 float64 向量转换为 float32 向量
			if i >= len(vectors) {
				return nil, fmt.Errorf("向量数量不匹配: docs=%d, vectors=%d", len(docs), len(vectors))
			}

			vector32 := make([]float32, len(vectors[i]))
			for j, v := range vectors[i] {
				vector32[j] = float32(v)
			}

			// 构建行数据
			// 注意: 字段顺序必须与 Fields 定义一致
			row := map[string]interface{}{
				"id":       fmt.Sprintf("doc_%d", i+1),
				"content":  doc.Content,
				"vector":   vector32,
				"metadata": doc.MetaData,
			}
			rows[i] = row
		}
		return rows, nil
	}

	// 创建 Milvus Indexer
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            cli,
		Collection:        collectionName,
		Embedding:         emb,
		Fields:            fields,
		MetricType:        milvus.CONSINE,
		DocumentConverter: documentConverter,
	})
	if err != nil {
		log.Fatalf("创建 Milvus Indexer 失败: %v", err)
	}
	log.Printf("成功创建 Milvus Indexer")

	// 准备文档
	docs := []*schema.Document{
		{
			ID:      "doc_1",
			Content: "Go 语言是一门开源的编程语言，具有并发编程的优势。",
			MetaData: map[string]any{
				"source": "go_intro",
				"type":   "programming",
			},
		},
		{
			ID:      "doc_2",
			Content: "深圳是中国的科技创新中心，拥有众多高科技企业。",
			MetaData: map[string]any{
				"source": "shenzhen_info",
				"type":   "city",
			},
		},
	}

	// 索引文档
	fmt.Println("开始索引文档...")
	indexedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("索引文档失败: %v", err)
	}
	
	fmt.Printf("成功索引文档，文档ID: %v\n", indexedIDs)
}

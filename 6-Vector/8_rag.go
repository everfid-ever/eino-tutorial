package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	milvusindexer "github.com/cloudwego/eino-ext/components/indexer/milvus"
	milvusretriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	collectionName = "eino_rag_example"
)

func main() {
	ctx := context.Background()

	addr := os.Getenv("MILVUS_ADDRESS")
	if addr == "" {
		addr = "192.168.179.138.19530" // 默认地址
	}

	// 创建客户端
	cli, err := client.NewClient(ctx, client.Config{
		Address: addr,
	})
	if err != nil {
		panic("连接 Milvus 失败: " + err.Error())
	}
	defer cli.Close()

	// 2. 创建 Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ARK Embedding 模型失败: %v", err)
	}

	// 3. 检测向量维度
	testVector, err := embedder.EmbedStrings(ctx, []string{"test"})
	if err != nil {
		log.Fatalf("获取 Embedding 维度失败: %v", err)
	}
	if len(testVector) == 0 || len(testVector[0]) == 0 {
		log.Fatalf("获取到的测试向量无效")
	}
	dim := int64(len(testVector[0]))
	log.Printf("检测到 Embedding 维度: %d", dim)

	// 4. 使用自定义字段: 使用浮点向量
	fields := []*entity.Field{
		entity.NewField().WithName("id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(255).WithIsPrimaryKey(true),
		entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(1024),
		entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(dim),
		entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON),
	}

	// 5. 自定义 DocumentConverter: 将 float64 向量转化为 float32
	documentConverter := func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
		rows := make([]interface{}, len(docs))
		for i, doc := range docs {
			if i >= len(vectors) {
				return nil, fmt.Errorf("向量数量不匹配: docs=%d, vectors=%d", len(docs), len(vectors))
			}

			vector32 := make([]float32, len(vectors[i]))
			for j, v := range vectors[i] {
				vector32[j] = float32(v)
			}

			rows[i] = map[string]interface{}{
				"id":       doc.ID,
				"content":  doc.Content,
				"vector":   vector32,
				"metadata": doc.MetaData,
			}
		}
		return rows, nil
	}

	// 6. 创建 Milvus Indexer
	indexer, err := milvusindexer.NewIndexer(ctx, &milvusindexer.IndexerConfig{
		Client:            cli,
		Collection:        collectionName,
		Embedding:         embedder,
		Fields:            fields,
		DocumentConverter: documentConverter,
	})
	if err != nil {
		log.Fatalf("创建 Milvus Indexer 失败: %v", err)
	}

	// 准备知识库文档
	docs := []*schema.Document{
		{
			ID:      "eino_1",
			Content: "Eino 是一个高性能的开源 AI 平台，旨在简化 AI 模型的集成和部署。",
			MetaData: map[string]any{
				"source":   "eino_intro",
				"category": "framework",
			},
		},
		{
			ID:      "eino_2",
			Content: "Eino 支持多种 AI 模型，包括语言模型和嵌入模型，方便开发者构建智能应用。",
			MetaData: map[string]any{
				"source":   "eino_components",
				"category": "framework",
			},
		},
		{
			ID:      "eino_3",
			Content: "ReAct Agent 结合了反思（Reflection）和行动（Action）两种能力，能够更有效地解决复杂任务。",
			MetaData: map[string]any{
				"source":   "eino_agent",
				"category": "agent",
			},
		},
		{
			ID:      "eino_4",
			Content: "ElasticSearch 是一个分布式搜索和分析引擎，广泛应用于日志分析和全文搜索等场景。",
			MetaData: map[string]any{
				"source":   "es_intro",
				"category": "database",
			},
		},
	}

	fmt.Println("=== 开始索引文档 ===")
	indexedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("索引文档失败: %v", err)
	}

	fmt.Printf("成功索引文档，文档ID: %v\n", indexedIDs)
	fmt.Println("=== 索引完成 ===")

	// 7. 创建 Milvus Retriever
	searchParam, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		log.Fatalf("创建搜索参数失败: %v", err)
	}

	retriever, err := milvusretriever.NewRetriever(ctx, &milvusretriever.RetrieverConfig{
		Client:       cli,
		Collection:   collectionName,
		Embedding:    embedder,
		TopK:         3,
		MetricType:   entity.COSINE,
		OutputFields: []string{"id", "content", "metadata"},
		Sp:           searchParam,
	})
	if err != nil {
		log.Fatalf("创建 Milvus Retriever 失败: %v", err)
	}

	// 8. 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("ARK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.ai",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 9. RAG 流程: 检索 + 生成
	userQuery := "Go 语言有哪些主要的特性？"
	fmt.Printf("=== 用户查询: %s ===\n", userQuery)

	// 9.1 检索相关文档
	fmt.Println("步骤1: 检索相关文档...")
	retrieveDocs, err := retriever.Retrieve(ctx, userQuery)
	if err != nil {
		log.Fatalf("检索失败: %v", err)
	}
	fmt.Printf("检索到 %d 个相关文档:\n", len(retrieveDocs))
	for i, doc := range retrieveDocs {
		score := doc.Score()
		if score != 0 {
			fmt.Printf("文档 %d (相似度: %.4f): %s\n", i+1, score, doc.Content)
		} else {
			fmt.Printf("文档 %d: %s\n", i+1, doc.Content)
		}
	}
	fmt.Println()

	// 9.2 构建上下文
	contextText := "相关文档内容:\n\n"
	for i, doc := range retrieveDocs {
		contextText += fmt.Sprintf("%d: %s\n", i+1, doc.Content)
	}

	// 9.3 生成回答
	fmt.Println("步骤2: 生成回答...")
	message := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf(`你是一个知识丰富的 AI 助手。请根据以下提供的相关文档内容，回答用户的问题。如果文档中没有相关信息，请如实告知用户你无法回答该问题。%s`, contextText)),
		schema.UserMessage(fmt.Sprintf(userQuery)),
	}
	response, err := chatModel.Generate(ctx, message)
	if err != nil {
		log.Fatalf("生成回答失败: %v", err)
	}
	fmt.Println("=== AI回答: ===")
	fmt.Println(response.Content)
}

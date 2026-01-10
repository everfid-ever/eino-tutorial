package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	es8indexer "github.com/cloudwego/eino-ext/components/indexer/es8"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func main() {
	ctx := context.Background()

	// 1. 创建 ElasticSearch 客户端
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("连接 ElasticSearch 失败: %v", err)
	}

	// 2. 创建 Embedding 模型
	// 创建 ARK Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ARK Embedding 模型失败: %v", err)
	}

	// 3. 创建 Indexer
	indexer, err := es8indexer.NewIndexer(ctx, &es8indexer.IndexerConfig{
		Client:    client,
		Index:     indexName,
		Embedding: embedder,
		DocumentToFields: func(ctx context.Context, doc *schema.Document) (map[string]es8indexer.FieldValue, error) {
			fields := make(map[string]es8indexer.FieldValue)
			// 文本内容字段, 设置 EmbedKey 以便自动向量化
			fields[filedContent] = es8indexer.FieldValue{
				Value:    doc.Content,
				EmbedKey: fieldContentVector, // 对文档内容进行向量化并保存到 content_vector 字段
			}
			// 将元数据也存储为字段
			for k, v := range doc.MetaData {
				fields[k] = es8indexer.FieldValue{
					Value: v,
				}
			}
			return fields, nil
		},
	})

	if err != nil {
		log.Fatalf("创建 ElasticSearch Indexer 失败: %v", err)
	}

	// 准备知识库文档
	docs := []*schema.Document{
		{
			Content: "Eino 是一个高性能的开源 AI 平台，旨在简化 AI 模型的集成和部署。",
			MetaData: map[string]any{
				"source":   "eino_intro",
				"category": "framework",
			},
		},
		{
			Content: "Eino 支持多种 AI 模型，包括语言模型和嵌入模型，方便开发者构建智能应用。",
			MetaData: map[string]any{
				"source":   "eino_components",
				"category": "framework",
			},
		},
		{
			Content: "ReAct Agent 结合了反思（Reflection）和行动（Action）两种能力，能够更有效地解决复杂任务。",
			MetaData: map[string]any{
				"source":   "eino_agent",
				"category": "agent",
			},
		},
		{
			Content: "ElasticSearch 是一个分布式搜索和分析引擎，广泛应用于日志分析和全文搜索等场景。",
			MetaData: map[string]any{
				"source":   "es_intro",
				"category": "database",
			},
		},
	}

	fmt.Println("=== 开始索引文档 ===")
	// 5. 索引文档
	indexedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("索引文档失败: %v", err)
	}

	fmt.Printf("成功索引文档，文档ID: %v\n", indexedIDs)
	for i, id := range indexedIDs {
		fmt.Printf("文档 %d ID: %s, 内容: %s\n", i+1, id, docs[i].Content)
	}
	fmt.Println("=== 索引完成 ===")

	// 手动刷新索引, 确保数据立即可搜索
	res, err := client.Indices.Refresh(client.Indices.Refresh.WithIndex(indexName))
	if err != nil {
		log.Fatalf("刷新索引失败: %v", err)
	} else {
		res.Body.Close()
		fmt.Println("索引已刷新，文档可搜索。")
	}

	// 6. 创建检索器 (近似搜索)
	retriever, err := es8.NewRetriever(ctx, &es8.RetrieverConfig{
		Client: client,
		Index:  indexName,
		TopK:   5,
		SearchMode: search_mode.SearchModeApproximate(&search_mode.ApproximateConfig{
			QueryFieldName:  filedContent,
			VectorFieldName: fieldContentVector,
			K:               intPtr(5),  // 返回最相似的 5 条结果
			NumCandidates:   intPtr(10), // 预选 10 条候选
			Hybrid:          true,       // 启用混合搜索
			RRF:             false,      // 不启用 RRF 重排序
		}),
		ResultParser: func(ctx context.Context, hit types.Hit) (doc *schema.Document, err error) {
			doc = &schema.Document{
				ID:       *hit.Id_,
				Content:  "",
				MetaData: map[string]any{},
			}

			var src map[string]any
			if err = json.Unmarshal(hit.Source_, &src); err != nil {
				return nil, err
			}

			// 解析字段
			for field, val := range src {
				switch field {
				case filedContent:
					doc.Content = val.(string)
				case fieldContentVector:
					// 向量字段 (可选)
					var v []float64
					for _, item := range val.([]interface{}) {
						v = append(v, item.(float64))
					}
					doc.WithDenseVector(v)
				}
			}

			// 添加相似度分数
			if hit.Source_ != nil {
				doc.WithScore(float64(*hit.Score_))
			}

			return doc, nil
		},
		Embedding: embedder,
	})
	if err != nil {
		log.Fatalf("创建 ElasticSearch 检索器失败: %v", err)
	}

	// 7. 创建检索
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("CHAT_MODEL_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 8. RAG 流程: 检索 + 生成
	userQuery := "Eino 框架有哪些主要的组件"
	fmt.Printf("=== 用户查询: %s ===\n", userQuery)

	// 执行检索
	fmt.Println("步骤1: 检索相关文档...")
	retrieveDocs, err := retriever.Retrieve(ctx, userQuery)
	if err != nil {
		log.Fatalf("检索失败: %v", err)
	}
	fmt.Printf("检索到 %d 条相关文档。\n", len(retrieveDocs))
	for i, doc := range retrieveDocs {
		score := doc.Score()
		if score != 0 {
			fmt.Printf("文档 %d (相似度: %.4f): %s\n", i+1, score, doc.Content)
		} else {
			fmt.Printf("文档 %d: %s\n", i+1, doc.Content)
		}
		fmt.Println()

		// 构建上下文
		context := "相关文档内容:\\n\n"
		for i, doc := range retrieveDocs {
			context += fmt.Sprintf("%d: %s\\n", i+1, doc.Content)
		}

		// 使用 LLM 生成回答
		fmt.Println("步骤2: 生成回答...")
		messages := []*schema.Message{
			schema.SystemMessage(fmt.Sprintf(`你是一个知识丰富的 AI 助手。请根据以下提供的相关文档内容，回答用户的问题。如果文档中没有相关信息，请如实告知用户你无法回答该问题。%s`, context)),
			schema.UserMessage(fmt.Sprintf(userQuery)),
		}
		response, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Fatalf("生成回答失败: %v", err)
		}
		fmt.Println("=== AI回答: ===")
		fmt.Println(response.Content)
	}
}

func intPtr(n int) *int {
	return &n
}

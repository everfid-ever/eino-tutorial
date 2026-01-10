package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/es8"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
)

const (
	indexName          = "eino_example"
	filedContent       = "content"
	fieldContentVector = "content_vector"
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
	indexer, err := es8.NewIndexer(ctx, &es8.IndexerConfig{
		Client:    client,
		Index:     indexName,
		Embedding: embedder,
		DocumentToFields: func(ctx context.Context, doc *schema.Document) (map[string]es8.FieldValue, error) {
			fields := make(map[string]es8.FieldValue)
			// 文本内容字段, 设置 EmbedKey 以便自动向量化
			fields[filedContent] = es8.FieldValue{
				Value:    doc.Content,
				EmbedKey: fieldContentVector, // 对文档内容进行向量化并保存到 content_vector 字段
			}
			// 将元数据也存储为字段
			for k, v := range doc.MetaData {
				fields[k] = es8.FieldValue{
					Value: v,
				}
			}
			return fields, nil
		},
	})

	if err != nil {
		log.Fatalf("创建 ElasticSearch Indexer 失败: %v", err)
	}

	// 4. 准备文档
	docs := []*schema.Document{
		{
			Content: "Go 语言是一门开源的编程语言，具有并发编程的优势。",
			MetaData: map[string]any{
				"source": "go_intro",
				"type":   "programming",
			},
		},
		{
			Content: "深圳是中国的科技创新中心，拥有众多高科技企业。",
			MetaData: map[string]any{
				"source": "shenzhen_info",
				"type":   "city",
			},
		},
	}

	// 5. 索引文档
	fmt.Println("开始索引文档...")
	indexedIDs, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("索引文档失败: %v", err)
	}

	fmt.Printf("成功索引文档，文档ID: %v\n", indexedIDs)
	for i, id := range indexedIDs {
		fmt.Printf("文档 %d ID: %s, 内容: %s\n", i+1, id, docs[i].Content)
	}
}

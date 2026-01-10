package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

func main() {
	ctx := context.Background()

	// 创建 ARK Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("EINO_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 ARK Embedding 模型失败: %v", err)
	}

	// 待向量化的文本
	texts := []string{
		"Go 是一门编程语言",
		"Python 是一门编程语言",
		"今天的天气很好",
	}

	// 生成文本嵌入向量
	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		log.Fatalf("生成文本嵌入向量失败: %v", err)
	}

	// 输出结果
	for i, text := range texts {
		fmt.Printf("文本 %d: %s\\n", i+1, text)
		fmt.Printf(" 向量维度: %d\\n", len(vectors[i]))
		fmt.Printf(" 前5个维度: %v\\n", vectors[i][:5])
	}

	// 计算两个文本的相似度（余弦相似度）
	similarity12 := cosineSimilarity(vectors[0], vectors[1])
	similarity13 := cosineSimilarity(vectors[0], vectors[2])

	fmt.Printf("文本1和文本2的相似度: %.4f\\n", similarity12)
	fmt.Printf("文本1和文本3的相似度: %.4f\\n", similarity13)
}

// 计算余弦相似度
func cosineSimilarity(vecA, vecB []float64) float64 {
	if len(vecA) != len(vecB) {
		return 0.0
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(vecA); i++ {
		dotProduct += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

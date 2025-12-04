package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/soaringjerry/AnyQA/backend/config"
	"github.com/soaringjerry/AnyQA/backend/models"
)

// ChunkWithSimilarity 用于存储文档块及其与问题的相似度
type ChunkWithSimilarity struct {
	Chunk      models.DocumentChunk
	Similarity float64
}

// RetrieveRelevantChunks 根据问题检索最相关的文档块
func RetrieveRelevantChunks(db *sql.DB, cfg *config.Config, question string, sessionId string, topK int) ([]models.DocumentChunk, error) {
	if question == "" || sessionId == "" {
		return nil, fmt.Errorf("question and sessionId cannot be empty")
	}
	if topK <= 0 {
		topK = 5 // 默认检索5个最相关的块
	}

	// 1. 获取问题的嵌入向量
	openaiClient := NewOpenAIClient(cfg)
	questionEmbeddings, err := openaiClient.GetEmbeddings([]string{question})
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding for question: %w", err)
	}
	if len(questionEmbeddings) == 0 || len(questionEmbeddings[0]) == 0 {
		return nil, fmt.Errorf("received empty embedding for question")
	}
	questionEmbedding := questionEmbeddings[0]
	fmt.Printf("问题向量获取成功 (维度: %d)\n", len(questionEmbedding))

	// 2. 从缓存获取文档块向量（避免每次查询都解析 JSON）
	cache := GetVectorCache()
	cachedChunks, err := cache.GetSessionChunks(db, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached chunks for session %s: %w", sessionId, err)
	}

	var chunksWithSimilarity []ChunkWithSimilarity

	// 3. 计算相似度（直接使用缓存的向量，无需反序列化）
	fmt.Println("开始计算相似度...")
	for _, cached := range cachedChunks {
		similarity, err := cosineSimilarity(questionEmbedding, cached.Embedding)
		if err != nil {
			fmt.Printf("警告：计算文档块 %d 的相似度失败: %v\n", cached.ID, err)
			continue
		}

		chunk := models.DocumentChunk{
			ID:         cached.ID,
			DocumentID: cached.DocumentID,
			Content:    cached.Content,
			ChunkIndex: cached.ChunkIndex,
		}
		chunksWithSimilarity = append(chunksWithSimilarity, ChunkWithSimilarity{
			Chunk:      chunk,
			Similarity: similarity,
		})
	}
	fmt.Printf("相似度计算完成，处理了 %d 个有效块。\n", len(chunksWithSimilarity))

	if len(chunksWithSimilarity) == 0 {
		fmt.Println("没有找到可比较的文档块。")
		return []models.DocumentChunk{}, nil // 返回空切片，表示没有找到相关内容
	}

	// 4. 按相似度降序排序
	sort.Slice(chunksWithSimilarity, func(i, j int) bool {
		return chunksWithSimilarity[i].Similarity > chunksWithSimilarity[j].Similarity
	})

	// 5. 提取 topK 个块
	numToReturn := min(topK, len(chunksWithSimilarity))
	relevantChunks := make([]models.DocumentChunk, numToReturn)
	fmt.Printf("检索到 Top %d 相关块:\n", numToReturn)
	for i := 0; i < numToReturn; i++ {
		relevantChunks[i] = chunksWithSimilarity[i].Chunk
		fmt.Printf("  - 块 ID: %d, 相似度: %.4f, 内容: %s...\n",
			relevantChunks[i].ID,
			chunksWithSimilarity[i].Similarity,
			relevantChunks[i].Content[:min(50, len(relevantChunks[i].Content))])
	}

	return relevantChunks, nil
}

// cosineSimilarity 计算两个 float32 切片的余弦相似度
func cosineSimilarity(vecA, vecB []float32) (float64, error) {
	if len(vecA) != len(vecB) {
		return 0, fmt.Errorf("vector dimensions mismatch: %d != %d", len(vecA), len(vecB))
	}
	if len(vecA) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct float64
	var normA, normB float64

	for i := 0; i < len(vecA); i++ {
		dotProduct += float64(vecA[i] * vecB[i])
		normA += float64(vecA[i] * vecA[i])
		normB += float64(vecB[i] * vecB[i])
	}

	// 检查零向量
	if normA == 0 || normB == 0 {
		// 如果其中一个是零向量，相似度未定义或为0，取决于具体场景
		// 这里返回0并打印警告
		fmt.Println("警告：计算余弦相似度时遇到零向量。")
		return 0, nil // 或者返回错误 fmt.Errorf("zero vector encountered")
	}

	magnitude := math.Sqrt(normA) * math.Sqrt(normB)
	if magnitude == 0 {
		fmt.Println("警告：向量幅度为零。")
		return 0, nil
	}

	similarity := dotProduct / magnitude

	// 处理浮点数精度问题，确保结果在 [-1, 1] 范围内
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}

	return similarity, nil
}

// min 返回两个整数中较小的一个 (如果 document_processor.go 中没有，可以在这里也定义一个)
// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

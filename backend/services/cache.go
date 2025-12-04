package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CachedChunk 缓存的文档块（已解析向量）
type CachedChunk struct {
	ID         int
	DocumentID int
	Content    string
	ChunkIndex int
	Embedding  []float32
}

// SessionCache 单个 session 的向量缓存
type SessionCache struct {
	Chunks    []CachedChunk
	UpdatedAt time.Time
}

// VectorCache 向量缓存管理器
type VectorCache struct {
	mu       sync.RWMutex
	sessions map[string]*SessionCache
	ttl      time.Duration
}

var (
	globalCache *VectorCache
	cacheOnce   sync.Once
)

// GetVectorCache 获取全局向量缓存实例
func GetVectorCache() *VectorCache {
	cacheOnce.Do(func() {
		globalCache = &VectorCache{
			sessions: make(map[string]*SessionCache),
			ttl:      30 * time.Minute, // 缓存 30 分钟
		}
	})
	return globalCache
}

// GetSessionChunks 获取 session 的缓存向量，如果没有或过期则从数据库加载
func (vc *VectorCache) GetSessionChunks(db *sql.DB, sessionId string) ([]CachedChunk, error) {
	vc.mu.RLock()
	cache, exists := vc.sessions[sessionId]
	if exists && time.Since(cache.UpdatedAt) < vc.ttl {
		chunks := cache.Chunks
		vc.mu.RUnlock()
		return chunks, nil
	}
	vc.mu.RUnlock()

	// 缓存不存在或过期，从数据库加载
	return vc.refreshSessionCache(db, sessionId)
}

// refreshSessionCache 从数据库刷新缓存
func (vc *VectorCache) refreshSessionCache(db *sql.DB, sessionId string) ([]CachedChunk, error) {
	query := `
		SELECT dc.id, dc.document_id, dc.content, dc.chunk_index, dc.embedding
		FROM document_chunks dc
		JOIN documents d ON dc.document_id = d.id
		WHERE d.session_id = ? AND dc.embedding IS NOT NULL AND dc.embedding != ''
	`
	rows, err := db.Query(query, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks for cache: %w", err)
	}
	defer rows.Close()

	var chunks []CachedChunk
	for rows.Next() {
		var chunk CachedChunk
		var embeddingJSON string

		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &embeddingJSON); err != nil {
			fmt.Printf("警告：缓存扫描文档块失败: %v\n", err)
			continue
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &chunk.Embedding); err != nil {
			fmt.Printf("警告：缓存解析向量失败 (chunk %d): %v\n", chunk.ID, err)
			continue
		}

		if len(chunk.Embedding) > 0 {
			chunks = append(chunks, chunk)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chunk rows for cache: %w", err)
	}

	// 更新缓存
	vc.mu.Lock()
	vc.sessions[sessionId] = &SessionCache{
		Chunks:    chunks,
		UpdatedAt: time.Now(),
	}
	vc.mu.Unlock()

	fmt.Printf("缓存刷新: session %s, %d 个向量块\n", sessionId, len(chunks))
	return chunks, nil
}

// InvalidateSession 使 session 缓存失效（文档上传/删除时调用）
func (vc *VectorCache) InvalidateSession(sessionId string) {
	vc.mu.Lock()
	delete(vc.sessions, sessionId)
	vc.mu.Unlock()
	fmt.Printf("缓存失效: session %s\n", sessionId)
}

// InvalidateDocument 使包含特定文档的缓存失效
func (vc *VectorCache) InvalidateDocument(db *sql.DB, docId int) {
	// 查询文档所属的 session
	var sessionId string
	err := db.QueryRow("SELECT session_id FROM documents WHERE id = ?", docId).Scan(&sessionId)
	if err != nil {
		return
	}
	vc.InvalidateSession(sessionId)
}

// GetStats 获取缓存统计信息
func (vc *VectorCache) GetStats() map[string]interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	totalChunks := 0
	for _, cache := range vc.sessions {
		totalChunks += len(cache.Chunks)
	}

	return map[string]interface{}{
		"sessions":     len(vc.sessions),
		"total_chunks": totalChunks,
		"ttl_minutes":  vc.ttl.Minutes(),
	}
}

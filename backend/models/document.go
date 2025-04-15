package models

import "time"

// Document 对应数据库中的 documents 表
type Document struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"sessionId"`
	Title      string    `json:"title"`
	FilePath   string    `json:"filePath"`
	FileType   string    `json:"fileType"`
	UploadTime time.Time `json:"uploadTime"`
}

// DocumentChunk 对应数据库中的 document_chunks 表
type DocumentChunk struct {
	ID         int    `json:"id"`
	DocumentID int    `json:"documentId"`
	Content    string `json:"content"`
	ChunkIndex int    `json:"chunkIndex"`
	Embedding  string `json:"embedding"` // 存储为 JSON 字符串
}

package models

import "time"

// Question 对应数据库中的 questions 表
type Question struct {
	ID           int       `json:"id"`
	SessionID    string    `json:"sessionId"`
	Content      string    `json:"content"`
	Status       string    `json:"status"` // 'pending', 'showing', 'answered', 'finished'
	AiSuggestion string    `json:"aiSuggestion"`
	KbSuggestion string    `json:"kbSuggestion"` // 新增：知识库回答建议
	CreatedAt    time.Time `json:"createdAt"`
}

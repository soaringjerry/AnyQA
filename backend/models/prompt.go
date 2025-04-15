package models

import "time"

// SessionPrompt 对应数据库中的 session_prompts 表
type SessionPrompt struct {
	SessionID     string    `json:"sessionId" db:"session_id"`
	GenericPrompt *string   `json:"genericPrompt" db:"generic_prompt"` // 使用指针以区分 NULL 和空字符串
	KbPrompt      *string   `json:"kbPrompt" db:"kb_prompt"`           // 使用指针
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

package handlers

import (
	"database/sql"
	"net/http"
	"time" // 导入 time 包

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/AnyQA/backend/config" // 确保路径正确
	"github.com/soaringjerry/AnyQA/backend/models" // 确保路径正确
)

// GetSessionPrompts 获取指定会话的自定义提示词，如果不存在则返回默认值
// GET /api/prompts/:sessionId
func GetSessionPrompts(c *gin.Context, db *sql.DB, cfg *config.Config) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	var prompt models.SessionPrompt
	err := db.QueryRow(`SELECT session_id, generic_prompt, kb_prompt, updated_at FROM session_prompts WHERE session_id = ?`, sessionId).Scan(
		&prompt.SessionID, &prompt.GenericPrompt, &prompt.KbPrompt, &prompt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有找到自定义提示词，返回配置中的默认值
			defaultPrompt := models.SessionPrompt{
				SessionID:     sessionId,
				GenericPrompt: &cfg.GenericSystemPrompt,       // 返回默认值指针
				KbPrompt:      &cfg.KnowledgeBaseSystemPrompt, // 返回默认值指针
				UpdatedAt:     time.Time{},                    // 表示未使用自定义
			}
			c.JSON(http.StatusOK, defaultPrompt)
			return
		}
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query session prompts: " + err.Error()})
		return
	}

	// 如果数据库中的字段为 NULL，Scan 会将指针设为 nil，我们需要处理这种情况
	// 如果用户从未设置过某个提示词，我们应该返回默认值
	if prompt.GenericPrompt == nil {
		prompt.GenericPrompt = &cfg.GenericSystemPrompt
	}
	if prompt.KbPrompt == nil {
		prompt.KbPrompt = &cfg.KnowledgeBaseSystemPrompt
	}

	c.JSON(http.StatusOK, prompt)
}

// UpdateSessionPrompts 更新或创建指定会话的自定义提示词
// POST /api/prompts/:sessionId
func UpdateSessionPrompts(c *gin.Context, db *sql.DB) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	var req models.SessionPrompt
	// 只绑定需要更新的字段，SessionID 从 URL 获取
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	req.SessionID = sessionId // 确保 SessionID 正确

	// 使用 INSERT ... ON DUPLICATE KEY UPDATE 来简化插入或更新逻辑
	query := `
        INSERT INTO session_prompts (session_id, generic_prompt, kb_prompt, updated_at)
        VALUES (?, ?, ?, NOW())
        ON DUPLICATE KEY UPDATE
            generic_prompt = VALUES(generic_prompt),
            kb_prompt = VALUES(kb_prompt),
            updated_at = NOW()
    `
	_, err := db.Exec(query, req.SessionID, req.GenericPrompt, req.KbPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session prompts: " + err.Error()})
		return
	}

	// 返回更新后的提示词（可选，或者只返回成功状态）
	// 为了简单起见，只返回成功状态
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "prompts updated successfully"})
}

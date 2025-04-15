package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/soaringjerry/AnyQA/backend/config"   // 确保路径正确
	"github.com/soaringjerry/AnyQA/backend/services" // 确保路径正确
)

// HandleQuestion 处理新问题的提交
// POST /api/question
func HandleQuestion(c *gin.Context, db *sql.DB, cfg *config.Config) { // 添加 cfg 参数
	var question struct {
		SessionID string `json:"sessionId"`
		Content   string `json:"content"`
	}
	if err := c.BindJSON(&question); err != nil {
		fmt.Printf("JSON绑定错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 首先插入问题内容到数据库，AI和知识库回答先留空
	result, err := db.Exec(`INSERT INTO questions (session_id, content, ai_suggestion, kb_suggestion) VALUES (?, ?, ?, ?)`,
		question.SessionID, question.Content, "", "") // 初始化 kb_suggestion 为空
	if err != nil {
		fmt.Printf("SQL执行错误: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	fmt.Printf("插入成功，ID: %d\n", id)

	// 立即返回给用户成功响应，不阻塞用户请求
	c.JSON(http.StatusOK, gin.H{"status": "success", "questionId": id})

	// 异步处理AI回复和知识库检索逻辑
	go func(questionID int64, qSessionID string, qContent string) {
		// 1. 获取通用 AI 建议
		openaiClient := services.NewOpenAIClient(cfg) // 创建 OpenAI 客户端
		// 传递 db 和 qSessionID 以获取自定义或默认提示词
		aiResponse, err := openaiClient.GetGenericAIResponse(db, cfg, qSessionID, qContent)
		if err != nil {
			fmt.Printf("获取通用 AI 建议错误 (问题ID %d): %v\n", questionID, err)
			aiResponse = "" // 出错则为空
		} else {
			fmt.Printf("问题ID %d 的通用 AI 建议获取成功。\n", questionID)
		}

		// 2. 从知识库检索相关文档块
		var kbSuggestion string = "" // 初始化知识库建议为空
		topK := 3                    // 检索最相关的3个块
		relevantChunks, err := services.RetrieveRelevantChunks(db, cfg, qContent, qSessionID, topK)
		if err != nil {
			fmt.Printf("知识库检索错误 (问题ID %d): %v\n", questionID, err)
			// 检索失败，kbSuggestion 保持为空
		} else if len(relevantChunks) > 0 {
			fmt.Printf("问题ID %d 检索到 %d 个相关文档块。\n", questionID, len(relevantChunks))
			// 使用检索到的块生成知识库回答
			// 注意：openaiClient 已经在上面创建过了，可以直接复用
			// 传递 db 和 qSessionID 以获取自定义或默认提示词
			generatedKbAnswer, genErr := openaiClient.GenerateAnswerWithContext(db, cfg, qSessionID, qContent, relevantChunks)
			if genErr != nil {
				fmt.Printf("生成知识库回答错误 (问题ID %d): %v\n", questionID, genErr)
				// 生成失败，可以提供一个默认提示或之前的简单拼接
				kbSuggestion = "【知识库参考】:\n（生成回答时出错，仅列出部分参考）\n"
				for _, chunk := range relevantChunks {
					// 需要一个 min 函数
					kbSuggestion += fmt.Sprintf("- %s...\n", chunk.Content[:minLocal(100, len(chunk.Content))])
				}
			} else {
				kbSuggestion = generatedKbAnswer // 使用 AI 生成的回答
				fmt.Printf("问题ID %d 的知识库回答生成成功。\n", questionID)
			}
		} else {
			fmt.Printf("问题ID %d 未在知识库中检索到相关内容。\n", questionID)
			kbSuggestion = "" // 如果未检索到内容，则知识库建议为空
		}

		// 3. 更新数据库记录，同时写入 AI 建议和知识库建议
		_, updateErr := db.Exec(`UPDATE questions SET ai_suggestion = ?, kb_suggestion = ? WHERE id = ?`,
			aiResponse, kbSuggestion, questionID)
		if updateErr != nil {
			fmt.Printf("更新问题 %d 的 AI 和知识库建议时出错: %v\n", questionID, updateErr)
		} else {
			fmt.Printf("问题 %d 的 AI 和知识库建议已更新。\n", questionID)
		}

	}(id, question.SessionID, question.Content) // 传递 SessionID 和 Content
}

// GetQuestions 获取指定会话的所有问题
// GET /api/questions/:sessionId
func GetQuestions(c *gin.Context, db *sql.DB) {
	sessionId := c.Param("sessionId")
	// 更新查询以包含 kb_suggestion
	rows, err := db.Query(`
	   SELECT id, content, status, ai_suggestion, kb_suggestion, created_at
	   FROM questions
	   WHERE session_id = ?
	   ORDER BY created_at DESC
	`, sessionId)
	if err != nil {
		fmt.Printf("查询错误: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var questions []map[string]interface{}
	for rows.Next() {
		q := make(map[string]interface{})
		var id int
		var content, status, createdAt string
		var aiSuggestion, kbSuggestion sql.NullString                                                       // 添加 kbSuggestion
		if err := rows.Scan(&id, &content, &status, &aiSuggestion, &kbSuggestion, &createdAt); err != nil { // 更新 Scan
			fmt.Printf("Scan错误: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		q["id"] = id
		q["content"] = content
		q["status"] = status
		q["ai_suggestion"] = aiSuggestion.String
		q["kb_suggestion"] = kbSuggestion.String // 添加 kb_suggestion 到响应
		q["created_at"] = createdAt
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error iterating question rows: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, questions)
}

// UpdateQuestionStatus 更新问题的状态
// POST /api/question/status
func UpdateQuestionStatus(c *gin.Context, db *sql.DB) {
	var req struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果设置showing状态，先把其他showing的改为finished
	if req.Status == "showing" {
		_, err := db.Exec("UPDATE questions SET status = 'finished' WHERE status = 'showing'")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_, err := db.Exec("UPDATE questions SET status = ? WHERE id = ?", req.Status, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// minLocal 返回两个整数中较小的一个 (本地辅助函数)
func minLocal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

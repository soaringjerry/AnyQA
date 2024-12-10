package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/soaringjerry/AnyQA/backend/config" // 替换为实际项目中的导入路径

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var db *sql.DB
var cfg *config.Config // 全局配置变量
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	// 初始化配置
	cfg = config.NewConfig()

	var err error
	db, err = sql.Open("mysql", cfg.GetDBDSN())
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}
	fmt.Println("数据库连接成功!")
}

func main() {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	r.POST("/api/question", handleQuestion)
	r.GET("/api/questions/:sessionId", getQuestions)
	r.GET("/api/ws", handleWebSocket)
	r.POST("/api/question/status", updateQuestionStatus)
	r.DELETE("/api/question/:id", func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM questions WHERE id = ?", id)
		if err != nil {
			fmt.Printf("删除错误: %v\n", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
	})

	r.Run(cfg.ServerPort)
}

func handleQuestion(c *gin.Context) {
	var question struct {
		SessionID string `json:"sessionId"`
		Content   string `json:"content"`
	}
	if err := c.BindJSON(&question); err != nil {
		fmt.Printf("JSON绑定错误: %v\n", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 首先插入问题内容到数据库，AI回答先留空
	result, err := db.Exec(`INSERT INTO questions (session_id, content, ai_suggestion) VALUES (?, ?, ?)`,
		question.SessionID, question.Content, "")
	if err != nil {
		fmt.Printf("SQL执行错误: %v\n", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	fmt.Printf("插入成功，ID: %d\n", id)

	// 立即返回给用户成功响应，不阻塞用户请求
	c.JSON(200, gin.H{"status": "success", "questionId": id})

	// 异步处理AI回复逻辑
	go func(questionID int64, content string) {
		aiResponse, err := getAIResponse(content)
		if err != nil {
			fmt.Printf("AI响应错误: %v\n", err)
			aiResponse = ""
		}
		fmt.Printf("AI回复: %s\n", aiResponse)

		// 更新数据库记录的AI回复字段
		if _, err := db.Exec(`UPDATE questions SET ai_suggestion = ? WHERE id = ?`, aiResponse, questionID); err != nil {
			fmt.Printf("更新数据库错误: %v\n", err)
		}
	}(id, question.Content)
}

func getQuestions(c *gin.Context) {
	sessionId := c.Param("sessionId")
	rows, err := db.Query(`
    SELECT id, content, status, ai_suggestion, created_at 
    FROM questions 
    WHERE session_id = ?
    ORDER BY created_at DESC
`, sessionId)
	if err != nil {
		fmt.Printf("查询错误: %v\n", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var questions []map[string]interface{}
	for rows.Next() {
		q := make(map[string]interface{})
		var id int
		var content, status, createdAt string
		var aiSuggestion sql.NullString
		if err := rows.Scan(&id, &content, &status, &aiSuggestion, &createdAt); err != nil {
			fmt.Printf("Scan错误: %v\n", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		q["id"] = id
		q["content"] = content
		q["status"] = status
		q["ai_suggestion"] = aiSuggestion.String
		q["created_at"] = createdAt
		questions = append(questions, q)
	}

	c.JSON(200, questions)
}

func updateQuestionStatus(c *gin.Context) {
	var req struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 如果设置showing状态，先把其他showing的改为finished
	if req.Status == "showing" {
		_, err := db.Exec("UPDATE questions SET status = 'finished' WHERE status = 'showing'")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	_, err := db.Exec("UPDATE questions SET status = ? WHERE id = ?", req.Status, req.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func handleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		ws.WriteMessage(websocket.TextMessage, msg)
	}
}

func getAIResponse(question string) (string, error) {
	requestBody := struct {
		Model    string        `json:"model"`
		Messages []ChatMessage `json:"messages"`
	}{
		Model: cfg.OpenAIModel,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a bilingual presentation assistant AI. Provide responses in markdown format to help Chinese speakers deliver English presentations. Always structure your response in two main sections: 1. Chinese Understanding: Provide core message, key terms (both Chinese and English), and key points for the speaker's understanding. 2. English Delivery: Give a quick answer, key speaking points, one clear example (marked as real or hypothetical), and useful phrases. Use headings (#), bold (**), and bullet points (*) in markdown format. Keep responses concise, use simple English, and clearly distinguish between real and hypothetical examples. Your goal is to help speakers understand the content in Chinese while preparing them to deliver confidently in English.",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", cfg.OpenAIAPIUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI")
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

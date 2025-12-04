package main

import (
	// "bytes" // 移除未使用的导入
	"database/sql"
	// "encoding/json" // 移除未使用的导入
	"fmt"
	"net/http"

	"github.com/soaringjerry/AnyQA/backend/config"   // 替换为实际项目中的导入路径
	"github.com/soaringjerry/AnyQA/backend/handlers" // 导入 handlers 包

	// "github.com/soaringjerry/AnyQA/backend/services" // 移除未使用的导入

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

	/* // CORS - 已移至外部 Nginx 处理
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE") // 注意：这里重复设置了 Allow-Methods

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	*/

	// API routes
	// 使用 handlers 包中的函数，并传递 db 和 cfg
	r.POST("/api/question", func(c *gin.Context) { handlers.HandleQuestion(c, db, cfg) })
	r.GET("/api/questions/:sessionId", func(c *gin.Context) { handlers.GetQuestions(c, db) })
	r.GET("/api/ws", handleWebSocket) // WebSocket 暂时保留在 main.go
	r.POST("/api/question/status", func(c *gin.Context) { handlers.UpdateQuestionStatus(c, db) })
	r.DELETE("/api/question/:id", func(c *gin.Context) { // 删除问题的逻辑比较简单，暂时保留匿名函数
		id := c.Param("id")
		_, err := db.Exec("DELETE FROM questions WHERE id = ?", id)
		if err != nil {
			fmt.Printf("删除错误: %v\n", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{})
	})
	// 新增：文档上传路由
	r.POST("/api/documents", func(c *gin.Context) { handlers.HandleDocumentUpload(c, db, cfg) })
	// 新增：获取文档列表路由
	r.GET("/api/documents/:sessionId", func(c *gin.Context) { handlers.GetSessionDocuments(c, db) })
	// 新增：删除文档路由
	r.DELETE("/api/document/:id", func(c *gin.Context) { handlers.DeleteDocument(c, db) })
	// 新增：获取会话提示词路由
	r.GET("/api/prompts/:sessionId", func(c *gin.Context) { handlers.GetSessionPrompts(c, db, cfg) })
	// 新增：更新会话提示词路由
	r.POST("/api/prompts/:sessionId", func(c *gin.Context) { handlers.UpdateSessionPrompts(c, db) })

	r.Run(cfg.ServerPort)
}

// WebSocket 处理逻辑 (暂时保留在 main.go)
func handleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer ws.Close()

	fmt.Println("WebSocket client connected")
	// TODO: 实现更复杂的 WebSocket 逻辑，例如广播消息等
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("WebSocket read error: %v\n", err)
			break // 连接关闭或出错
		}
		fmt.Printf("Received WebSocket message: %s\n", string(msg))
		// 简单地将消息回显给客户端
		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			fmt.Printf("WebSocket write error: %v\n", err)
			break
		}
	}
	fmt.Println("WebSocket client disconnected")
}

// 注意：getAIResponse 函数现在应该在 services/openai_service.go 中实现或调用
// 注意：ChatMessage 和 OpenAIResponse 结构体也应该移到相应的位置（例如 models 或 services）
// 注意：min 函数如果只在 handlers/question.go 中使用，可以移到那里或保持为本地函数

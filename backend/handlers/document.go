package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/soaringjerry/AnyQA/backend/config"   // 导入 config 包
	"github.com/soaringjerry/AnyQA/backend/models"   // 确保路径正确
	"github.com/soaringjerry/AnyQA/backend/services" // 导入 services 包
)

// HandleDocumentUpload 处理文档上传请求
// POST /api/documents
func HandleDocumentUpload(c *gin.Context, db *sql.DB) {
	// 1. 从表单获取会话ID
	sessionId := c.PostForm("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	// 2. 从表单获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file upload error: " + err.Error()})
		return
	}

	// 3. 创建上传目录（如果不存在）
	// 注意：在生产环境中，上传目录应配置化管理
	uploadDir := filepath.Join(".", "uploads", sessionId) // 按会话ID分子目录
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory: " + err.Error()})
		return
	}

	// 4. 生成唯一文件名，保留原始扩展名
	ext := filepath.Ext(file.Filename)
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, uniqueFilename)

	// 5. 保存文件到服务器
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file: " + err.Error()})
		return
	}

	// 6. 将文档元数据存入数据库
	doc := models.Document{
		SessionID:  sessionId,
		Title:      strings.TrimSuffix(file.Filename, ext), // 使用原始文件名作为标题（去除扩展名）
		FilePath:   filePath,                               // 存储相对路径
		FileType:   strings.ToLower(strings.TrimPrefix(ext, ".")),
		UploadTime: time.Now(),
	}

	result, err := db.Exec(`INSERT INTO documents (session_id, title, file_path, file_type, upload_time) VALUES (?, ?, ?, ?, ?)`,
		doc.SessionID, doc.Title, doc.FilePath, doc.FileType, doc.UploadTime)
	if err != nil {
		// 如果数据库插入失败，尝试删除已保存的文件
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save document metadata: " + err.Error()})
		return
	}

	docId, _ := result.LastInsertId()
	doc.ID = int(docId) // 获取插入的ID

	// 异步触发文档处理
	// 需要访问全局配置 cfg
	// 在实际应用中，考虑依赖注入或其他方式传递配置和数据库连接
	appConfig := config.NewConfig() // 重新加载配置，或者从 main 传递进来

	go func(docID int, filePath string, dbConn *sql.DB, cfgInstance *config.Config) { // 添加 cfgInstance 参数
		fmt.Printf("开始异步处理文档 ID: %d, Path: %s\n", docID, filePath)
		// 传递配置给处理函数
		err := services.ProcessUploadedDocument(dbConn, cfgInstance, docID, filePath)
		if err != nil {
			// 记录错误，实际应用中可能需要更健壮的错误处理机制
			fmt.Printf("异步处理文档 ID %d 失败: %v\n", docID, err)
		} else {
			fmt.Printf("异步处理文档 ID %d 完成\n", docID)
		}
	}(doc.ID, doc.FilePath, db, appConfig) // 传递数据库连接和配置实例

	c.JSON(http.StatusOK, gin.H{"status": "success", "document": doc})
}

// GetSessionDocuments 获取指定会话的所有已上传文档
// GET /api/documents/:sessionId
func GetSessionDocuments(c *gin.Context, db *sql.DB) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	rows, err := db.Query(`SELECT id, session_id, title, file_path, file_type, upload_time FROM documents WHERE session_id = ? ORDER BY upload_time DESC`, sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query documents: " + err.Error()})
		return
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.SessionID, &doc.Title, &doc.FilePath, &doc.FileType, &doc.UploadTime); err != nil {
			fmt.Printf("扫描文档行错误: %v\n", err)
			// 可以选择继续处理其他行或直接返回错误
			continue
		}
		documents = append(documents, doc)
	}
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error iterating document rows: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// DeleteDocument 删除指定的文档及其关联数据和文件
// DELETE /api/document/:id
func DeleteDocument(c *gin.Context, db *sql.DB) {
	docId := c.Param("id")
	if docId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document id is required"})
		return
	}

	// 1. 开始事务
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction: " + err.Error()})
		return
	}
	defer tx.Rollback() // 保证出错时回滚

	// 2. 获取文件路径以便稍后删除文件
	var filePath string
	err = tx.QueryRow(`SELECT file_path FROM documents WHERE id = ?`, docId).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query document file path: " + err.Error()})
		}
		return
	}

	// 3. 从数据库删除文档记录 (关联的 chunks 会级联删除)
	_, err = tx.Exec(`DELETE FROM documents WHERE id = ?`, docId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document record: " + err.Error()})
		return
	}

	// 4. 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction: " + err.Error()})
		return
	}

	// 5. 事务成功后，尝试删除物理文件
	if filePath != "" {
		err := os.Remove(filePath)
		if err != nil {
			// 文件删除失败通常不应阻止API成功响应，但应记录错误
			fmt.Printf("警告：成功删除数据库记录，但删除文件 %s 失败: %v\n", filePath, err)
		} else {
			fmt.Printf("成功删除文件: %s\n", filePath)
		}
	} else {
		fmt.Printf("警告：文档 %s 的文件路径为空，无法删除物理文件。\n", docId)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "document deleted successfully"})
}

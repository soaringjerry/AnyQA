package services

import (
	"database/sql"
	"encoding/json" // 用于将向量序列化为JSON字符串
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// "unicode/utf8" // 移除未使用的导入

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/soaringjerry/AnyQA/backend/config" // 导入 config 包
	// "github.com/soaringjerry/AnyQA/backend/models" // 移除未使用的导入
)

// ProcessUploadedDocument 是处理上传文档的主函数
// 它会提取文本、分块、向量化并存储
func ProcessUploadedDocument(db *sql.DB, cfg *config.Config, docID int, filePath string) error { // 添加 cfg 参数
	// 1. 提取文本
	textContent, err := ExtractTextFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to extract text for doc %d: %w", docID, err)
	}
	// 使用 TrimSpace 检查是否只有空白字符
	if strings.TrimSpace(textContent) == "" {
		fmt.Printf("文档 %d (%s) 内容为空或无法提取，跳过处理。\n", docID, filePath)
		return nil // 内容为空不是致命错误，但需要记录
	}

	// 2. 文本分块
	// TODO: 使 chunkSize 和 chunkOverlap 可配置
	chunkSize := 1000   // 示例分块大小（字符数）
	chunkOverlap := 100 // 示例重叠大小（字符数）
	chunks := chunkText(textContent, chunkSize, chunkOverlap)
	fmt.Printf("文档 %d 分块完成，共 %d 块。\n", docID, len(chunks))

	if len(chunks) == 0 {
		fmt.Printf("文档 %d 没有有效的文本块，处理结束。\n", docID)
		return nil
	}

	// 3. 向量化每个块 (调用OpenAI API)
	openaiClient := NewOpenAIClient(cfg) // 创建 OpenAI 客户端
	embeddings, err := openaiClient.GetEmbeddings(chunks)
	if err != nil {
		return fmt.Errorf("failed to get embeddings for doc %d: %w", docID, err)
	}

	if len(embeddings) != len(chunks) {
		return fmt.Errorf("embedding count mismatch for doc %d: expected %d, got %d", docID, len(chunks), len(embeddings))
	}

	fmt.Printf("文档 %d 向量化完成，获得 %d 个向量。\n", docID, len(embeddings))

	// 4. 将块和向量存储到 document_chunks 表
	// 使用事务确保原子性
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for doc %d: %w", docID, err)
	}
	defer tx.Rollback() // 如果后续出错，回滚事务

	stmt, err := tx.Prepare(`INSERT INTO document_chunks (document_id, content, chunk_index, embedding) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for doc %d: %w", docID, err)
	}
	defer stmt.Close()

	for i, chunk := range chunks {
		if i >= len(embeddings) || len(embeddings[i]) == 0 {
			fmt.Printf("警告：文档 %d 的块 %d 没有有效的嵌入向量，跳过存储。\n", docID, i)
			continue // 跳过没有有效向量的块
		}

		// 将 float32 切片序列化为 JSON 字符串
		embeddingJSON, err := json.Marshal(embeddings[i])
		if err != nil {
			// 这个错误比较严重，可能需要中断处理或记录详细日志
			fmt.Printf("错误：无法序列化文档 %d 块 %d 的向量: %v\n", docID, i, err)
			continue // 或者 return err
		}

		_, err = stmt.Exec(docID, chunk, i, string(embeddingJSON))
		if err != nil {
			return fmt.Errorf("failed to insert chunk %d for doc %d: %w", i, docID, err)
		}
		fmt.Printf("  文档 %d 块 %d 已存储。\n", docID, i)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for doc %d: %w", docID, err)
	}

	fmt.Printf("文档 %d 所有块和向量存储完成。\n", docID)
	return nil
}

// ExtractTextFromFile 根据文件路径和类型提取文本内容
func ExtractTextFromFile(filePath string) (string, error) {
	fileType := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	switch fileType {
	case "pdf":
		return extractTextFromPDF(filePath)
	case "docx":
		return extractTextFromDOCX(filePath)
	case "pptx":
		// TODO: 实现PPTX文本提取
		return "", fmt.Errorf("pptx text extraction not implemented yet")
	case "txt":
		contentBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read txt file %s: %w", filePath, err)
		}
		return string(contentBytes), nil
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// extractTextFromPDF 从PDF文件中提取文本
func extractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	// 记得关闭文件
	defer f.Close()
	if err != nil {
		return "", fmt.Errorf("failed to open pdf file %s: %w", filePath, err)
	}

	var buf strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			// 尝试继续处理其他页面
			fmt.Printf("Warning: failed to get text from page %d of %s: %v\n", pageIndex, filePath, err)
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n") // 添加换行符分隔页面内容
	}

	return buf.String(), nil
}

// extractTextFromDOCX 从DOCX文件中提取文本
func extractTextFromDOCX(filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read docx file %s: %w", filePath, err)
	}
	defer r.Close()

	content := r.Editable().GetContent()
	// DOCX库可能返回XML格式的文本，需要进一步清理或解析
	// 这里的实现比较基础，可能需要根据实际效果调整
	// 简单替换掉XML标签（非常粗糙的方式）
	// re := regexp.MustCompile(`<[^>]*>`)
	// plainText := re.ReplaceAllString(content, "")
	// return plainText, nil
	// 暂时直接返回原始内容，后续可能需要更精细的处理
	return content, nil
}

// chunkText 将文本按指定大小和重叠进行分块
// 注意：这是一个简单的基于字符数的分块实现，更高级的实现会考虑句子边界等。
func chunkText(text string, chunkSize int, chunkOverlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000 // 默认值
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = 100 // 默认或修正重叠值
	}

	var chunks []string
	textRunes := []rune(text) // 使用 rune 处理多字节字符
	textLen := len(textRunes)
	start := 0

	for start < textLen {
		end := start + chunkSize
		if end > textLen {
			end = textLen
		}

		chunks = append(chunks, string(textRunes[start:end]))

		// 计算下一个块的起始位置
		start += chunkSize - chunkOverlap
		// 如果重叠导致 start 没有前进，强制前进一小步避免死循环
		if start <= (end-chunkSize) && start < textLen {
			start = end - chunkSize + 1
		}
		// 确保 start 不会因为重叠计算而回退太多
		if start < end-chunkSize+chunkOverlap && start > 0 {
			start = end - chunkSize + chunkOverlap
		}
		// 避免 start 越界
		if start >= textLen {
			break
		}
	}

	// 过滤掉可能产生的空块
	var nonEmptyChunks []string
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) != "" {
			nonEmptyChunks = append(nonEmptyChunks, chunk)
		}
	}

	return nonEmptyChunks
}

// min 返回两个整数中较小的一个
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TODO: 添加文本向量化函数 (调用OpenAI API)
// TODO: 添加存储文档块和向量的函数

package services

import (
	"bytes"
	"database/sql" // 确保导入 database/sql
	"encoding/json"
	"fmt"
	"net/http"
	"strings" // 确保导入 strings

	"github.com/soaringjerry/AnyQA/backend/config" // 确保路径正确
	"github.com/soaringjerry/AnyQA/backend/models" // 导入 models 包
)

// ChatMessage 定义了聊天消息的结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatCompletionRequest 定义了调用 Chat Completions API 的请求体结构
type OpenAIChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// OpenAIChatCompletionResponse 定义了 Chat Completions API 响应体的主要结构
type OpenAIChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage UsageData `json:"usage"`
}

// OpenAIEmbeddingRequest 定义了调用 OpenAI Embeddings API 的请求体结构
type OpenAIEmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

// OpenAIEmbeddingResponse 定义了 OpenAI Embeddings API 响应体的主要结构
type OpenAIEmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  UsageData       `json:"usage"`
}

// EmbeddingData 包含单个输入的嵌入向量信息
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// UsageData 包含 API 使用情况信息
type UsageData struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// OpenAIClient 封装了与 OpenAI API 的交互
type OpenAIClient struct {
	apiKey          string
	chatAPIURL      string
	embeddingAPIURL string
	chatModel       string
	embeddingModel  string
}

// NewOpenAIClient 创建一个新的 OpenAIClient 实例
func NewOpenAIClient(cfg *config.Config) *OpenAIClient {
	embeddingsURL := "https://api.openai.com/v1/embeddings"
	return &OpenAIClient{
		apiKey:          cfg.OpenAIAPIKey,
		chatAPIURL:      cfg.OpenAIAPIUrl,
		embeddingAPIURL: embeddingsURL,
		chatModel:       cfg.OpenAIModel,
		embeddingModel:  cfg.OpenAIEmbeddingModel,
	}
}

// GetEmbeddings 获取一批文本的嵌入向量
func (client *OpenAIClient) GetEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("input texts cannot be empty")
	}
	if client.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	requestBody := OpenAIEmbeddingRequest{
		Input:          texts,
		Model:          client.embeddingModel,
		EncodingFormat: "float",
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", client.embeddingAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for embeddings: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.apiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send embeddings request to OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return nil, fmt.Errorf("OpenAI Embeddings API request failed with status %d: %v", resp.StatusCode, errorResponse)
		}
		return nil, fmt.Errorf("OpenAI Embeddings API request failed with status %d", resp.StatusCode)
	}

	var result OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI Embeddings API response: %w", err)
	}

	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf("mismatch between input texts (%d) and embeddings received (%d)", len(texts), len(result.Data))
	}

	embeddings := make([][]float32, len(texts))
	for _, data := range result.Data {
		if data.Index >= 0 && data.Index < len(texts) {
			embeddings[data.Index] = data.Embedding
		} else {
			return nil, fmt.Errorf("received invalid index %d from OpenAI API", data.Index)
		}
	}
	for i, emb := range embeddings {
		if len(emb) == 0 {
			fmt.Printf("Warning: Received empty embedding for input index %d\n", i)
		}
	}

	fmt.Printf("Successfully retrieved %d embeddings using model %s. Usage: %d prompt tokens, %d total tokens.\n",
		len(embeddings), result.Model, result.Usage.PromptTokens, result.Usage.TotalTokens)
	return embeddings, nil
}

// GenerateAnswerWithContext 使用检索到的上下文和指定的系统提示生成回答
func (client *OpenAIClient) GenerateAnswerWithContext(db *sql.DB, cfg *config.Config, sessionId string, question string, chunks []models.DocumentChunk) (string, error) {
	if client.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}
	if len(chunks) == 0 {
		return "知识库中没有找到相关信息来回答这个问题。", nil
	}

	contextStr := ""
	for i, chunk := range chunks {
		contextStr += fmt.Sprintf("相关信息片段 %d:\n\"%s\"\n\n", i+1, chunk.Content)
	}

	kbPromptTemplate := getSessionPromptOrDefault(db, sessionId, "kb", cfg.KnowledgeBaseSystemPrompt)
	systemPrompt := fmt.Sprintf(kbPromptTemplate, contextStr) // Insert context

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	requestBody := OpenAIChatCompletionRequest{
		Model:    client.chatModel,
		Messages: messages,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chat request body: %w", err)
	}

	req, err := http.NewRequest("POST", client.chatAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create http request for chat completion: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.apiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send chat completion request to OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return "", fmt.Errorf("OpenAI Chat API request failed with status %d: %v", resp.StatusCode, errorResponse)
		}
		return "", fmt.Errorf("OpenAI Chat API request failed with status %d", resp.StatusCode)
	}

	var result OpenAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI Chat API response: %w", err)
	}

	if len(result.Choices) > 0 && result.Choices[0].Message.Content != "" {
		fmt.Printf("Chat completion successful. Usage: %d prompt tokens, %d total tokens.\n",
			result.Usage.PromptTokens, result.Usage.TotalTokens)
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response content received from OpenAI Chat API")
}

// GetGenericAIResponse 获取通用的 AI 回答建议
func (client *OpenAIClient) GetGenericAIResponse(db *sql.DB, cfg *config.Config, sessionId string, question string) (string, error) {
	if client.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not configured")
	}

	systemPrompt := getSessionPromptOrDefault(db, sessionId, "generic", cfg.GenericSystemPrompt)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	requestBody := OpenAIChatCompletionRequest{
		Model:    client.chatModel,
		Messages: messages,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generic chat request body: %w", err)
	}

	req, err := http.NewRequest("POST", client.chatAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create http request for generic chat completion: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.apiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send generic chat completion request to OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return "", fmt.Errorf("OpenAI Generic Chat API request failed with status %d: %v", resp.StatusCode, errorResponse)
		}
		return "", fmt.Errorf("OpenAI Generic Chat API request failed with status %d", resp.StatusCode)
	}

	var result OpenAIChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI Generic Chat API response: %w", err)
	}

	if len(result.Choices) > 0 && result.Choices[0].Message.Content != "" {
		fmt.Printf("Generic chat completion successful. Usage: %d prompt tokens, %d total tokens.\n",
			result.Usage.PromptTokens, result.Usage.TotalTokens)
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response content received from OpenAI Generic Chat API")
}

// getSessionPromptOrDefault 尝试从数据库获取会话的特定提示词，如果失败或为空则返回默认值
func getSessionPromptOrDefault(db *sql.DB, sessionId string, promptType string, defaultValue string) string {
	var promptValue sql.NullString
	var query string

	switch promptType {
	case "generic":
		query = `SELECT generic_prompt FROM session_prompts WHERE session_id = ?`
	case "kb":
		query = `SELECT kb_prompt FROM session_prompts WHERE session_id = ?`
	default:
		fmt.Printf("警告：无效的提示词类型 '%s'，将使用默认值。\n", promptType)
		return defaultValue
	}

	err := db.QueryRow(query, sessionId).Scan(&promptValue)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Printf("警告：查询会话 %s 的 %s 提示词失败: %v。将使用默认值。\n", sessionId, promptType, err)
		}
		return defaultValue
	}

	if promptValue.Valid && strings.TrimSpace(promptValue.String) != "" {
		fmt.Printf("会话 %s 使用自定义 %s 提示词。\n", sessionId, promptType)
		return promptValue.String
	}

	return defaultValue
}

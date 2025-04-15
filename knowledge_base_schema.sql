-- 确保使用正确的数据库
USE aiqa;

-- 创建文档表
CREATE TABLE IF NOT EXISTS documents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session (session_id)
);

-- 创建文档片段表
CREATE TABLE IF NOT EXISTS document_chunks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    document_id INT NOT NULL,
    content TEXT NOT NULL,
    chunk_index INT NOT NULL,
    embedding LONGTEXT, -- 存储OpenAI嵌入向量的JSON字符串
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    INDEX idx_document (document_id)
);

-- 修改现有questions表，添加知识库回答字段
-- 首先检查列是否存在，避免重复添加
SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'questions' AND column_name = 'kb_suggestion');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE questions ADD COLUMN kb_suggestion TEXT AFTER ai_suggestion;', 'SELECT "Column kb_suggestion already exists.";');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 创建会话自定义提示词表
CREATE TABLE IF NOT EXISTS session_prompts (
    session_id VARCHAR(50) PRIMARY KEY,
    generic_prompt TEXT,
    kb_prompt TEXT, -- 存储知识库问答的提示词模板（可能包含 %s）
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

SELECT '数据库表结构更新完成（如果需要）。';
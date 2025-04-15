-- 确保使用正确的数据库
-- USE aiqa;

-- 创建文档表
CREATE TABLE IF NOT EXISTS documents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    title VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    file_path VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    file_type VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建文档片段表
CREATE TABLE IF NOT EXISTS document_chunks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    document_id INT NOT NULL,
     content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
     chunk_index INT NOT NULL,
     embedding LONGTEXT, -- 存储OpenAI嵌入向量的JSON字符串
     FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
     INDEX idx_document (document_id)
 );

-- 修改现有 questions 表以支持 utf8mb4
-- 检查表是否存在
SET @table_exists = (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'questions');
SET @sql_alter_questions = IF(@table_exists > 0,
   'ALTER TABLE questions CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;',
   'SELECT "Table questions does not exist, skipping alteration.";'
);
PREPARE stmt_alter_q FROM @sql_alter_questions;
EXECUTE stmt_alter_q;
DEALLOCATE PREPARE stmt_alter_q;

-- 修改现有 questions 表的列（如果表存在）
-- 检查列 content 是否存在并修改
SET @col_content_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'questions' AND column_name = 'content');
SET @sql_alter_content = IF(@col_content_exists > 0 AND @table_exists > 0,
   'ALTER TABLE questions MODIFY content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;',
   'SELECT "Column content in questions does not exist or table does not exist, skipping alteration.";'
);
PREPARE stmt_alter_content FROM @sql_alter_content;
EXECUTE stmt_alter_content;
DEALLOCATE PREPARE stmt_alter_content;

-- 检查列 ai_suggestion 是否存在并修改
SET @col_ai_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'questions' AND column_name = 'ai_suggestion');
SET @sql_alter_ai = IF(@col_ai_exists > 0 AND @table_exists > 0,
   'ALTER TABLE questions MODIFY ai_suggestion TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;',
   'SELECT "Column ai_suggestion in questions does not exist or table does not exist, skipping alteration.";'
);
PREPARE stmt_alter_ai FROM @sql_alter_ai;
EXECUTE stmt_alter_ai;
DEALLOCATE PREPARE stmt_alter_ai;


-- 修改现有questions表，添加知识库回答字段
-- 首先检查列是否存在，避免重复添加
SET @col_kb_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'questions' AND column_name = 'kb_suggestion');
SET @sql_add_kb = IF(@col_kb_exists = 0 AND @table_exists > 0, -- 只有在表存在且列不存在时才添加
   'ALTER TABLE questions ADD COLUMN kb_suggestion TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci AFTER ai_suggestion;',
   IF(@col_kb_exists > 0, 'SELECT "Column kb_suggestion already exists.";', 'SELECT "Table questions does not exist, cannot add kb_suggestion.";')
);
PREPARE stmt_add_kb FROM @sql_add_kb;
EXECUTE stmt_add_kb;
DEALLOCATE PREPARE stmt_add_kb;

-- 创建会话自定义提示词表
CREATE TABLE IF NOT EXISTS session_prompts (
    session_id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci PRIMARY KEY,
    generic_prompt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    kb_prompt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci, -- 存储知识库问答的提示词模板（可能包含 %s）
     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SELECT '数据库表结构更新完成（如果需要）。';
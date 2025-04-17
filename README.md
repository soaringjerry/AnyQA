# 演讲互动问答系统 (增强版)

一个支持实时提问、AI辅助回答、知识库增强和自定义提示词的演讲互动问答系统。

## ✨ 功能特点

*   **观众实时提问**: 观众通过网页或扫描二维码轻松提交问题。
*   **AI 辅助回答**:
    *   **通用建议**: 基于强大的 AI 模型生成通用的回答建议。
    *   **知识库问答 (新!)**: 演讲者可预先上传相关文档 (PDF, DOCX, TXT)，系统能基于文档内容生成更精准的回答。
*   **自定义提示词 (新!)**: 演讲者可在控制台为当前会话自定义通用 AI 和知识库问答的系统提示词 (System Prompt)。
*   **文档管理 (新!)**: 演讲者可在控制台上传、查看和删除知识库文档。
*   **实时互动**: 演讲者在控制台管理问题，控制大屏显示。
*   **大屏展示**: 将选中的问题清晰地展示在大屏幕上。
*   **多语言支持**: 前端界面支持多种语言切换。

## 🛠️ 技术栈

*   **前端 (主)**: Vue 3 + Vite + Vue Router + Vuetify + Marked (Markdown渲染) + vue-i18n (国际化)
*   **前端 (旧版 - 已弃用)**: 原生 HTML/CSS/JavaScript
*   **后端**: Go (Gin 框架) + Gorilla WebSocket
*   **数据库**: MySQL
*   **AI**: OpenAI API (Chat Completions & Embeddings)
*   **文档处理**: Go PDF/DOCX 文本提取库

## 🚀 快速开始

**强烈推荐使用 Vue 前端 (`frontend-vue`)**

### 1. 配置数据库

*   确保你有一个正在运行的 MySQL 服务。
*   创建一个数据库 (例如 `aiqa`)。
*   执行项目根目录下的 `database.sql` 和 `knowledge_base_schema.sql` 脚本来创建所需的表结构。
    ```bash
    # 示例命令 (需要安装 mysql 客户端)
    mysql -u <your_user> -p <your_password> < database.sql
    mysql -u <your_user> -p <your_password> aiqa < knowledge_base_schema.sql
    ```
    *   `database.sql`: 创建初始的 `questions` 表。
    *   `knowledge_base_schema.sql`: 添加 `documents`, `document_chunks`, `session_prompts` 表，并修改 `questions` 表以支持知识库功能。

### 2. 配置后端 (`backend/`)

*   **环境变量**: 后端配置通过环境变量加载。请参考 `backend/config/config.go.example` 文件，设置必要的环境变量，至少包括：
    *   `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`: 数据库连接信息。
    *   `OPENAI_API_KEYs`: 你的 OpenAI API 密钥。
    *   `OPENAI_MODEL` (可选): 指定聊天模型，默认为 `gpt-4o-latest`。
    *   `OPENAI_EMBEDDING_MODEL` (可选): 指定嵌入模型，默认为 `text-embedding-3-small`。
    *   `GENERIC_SYSTEM_PROMPT` (可选): 自定义通用 AI 建议的默认系统提示词。
    *   `KB_SYSTEM_PROMPT` (可选): 自定义知识库问答的默认系统提示词模板 (必须包含 `%s` 用于插入上下文)。
    *   `SERVER_PORT` (可选): 后端服务监听端口，默认为 `:8080`。
*   **安装依赖**:
    ```bash
    cd backend
    go mod tidy # 或者 go get ./... 来获取所有依赖
    ```
*   **运行后端**:
    ```bash
    # 开发模式
    go run main.go
    # 或者编译后运行
    # go build -o backend_app
    # ./backend_app
    ```

### 3. 配置并运行 Vue 前端 (`frontend-vue/`)

*   **配置文件**: 检查 `frontend-vue/src/config/config.yaml` 文件，确保 `api.endpoint` 指向你运行的后端地址 (例如 `http://localhost:8080/api`)。
*   **安装依赖**:
    ```bash
    cd frontend-vue
    npm install # 或者 yarn install
    ```
*   **运行开发服务器**:
    ```bash
    npm run dev # 或者 yarn dev
    ```
    终端会显示前端访问地址，通常是 `http://localhost:5173` 或类似地址。

### 4. 访问系统

*   **会话创建/设置页面**: 访问 Vue 前端地址 (例如 `http://localhost:5173`)，点击按钮生成新的会话 ID。
*   **观众提问页面**: 使用生成的链接或扫描二维码访问 (例如 `http://localhost:5173/index?sessionId=YOUR_SESSION_ID`)。
*   **演讲者控制台**: 使用生成的链接访问 (例如 `http://localhost:5173/presenter?sessionId=YOUR_SESSION_ID`)。
*   **大屏展示页面**: 使用生成的链接访问 (例如 `http://localhost:5173/display?sessionId=YOUR_SESSION_ID`)。

## 🚀 自动化部署 (CI/CD)

本项目配置了 GitHub Actions 以实现自动化构建和部署。当代码推送到 `main` 分支时，会自动构建 Docker 镜像并部署到目标服务器。

**详细的 CI/CD 配置步骤和要求，请参考：[CI/CD 设置指南 (CICD_SETUP.md)](./CICD_SETUP.md)**

##  页面/组件说明 (Vue 前端)

*   **`IndexPage.vue`**: 观众提问页面。
*   **`PresenterPage.vue`**: 演讲者控制台，用于：
    *   查看和管理实时问题。
    *   查看 AI 生成的通用建议和基于知识库的回答建议。
    *   控制问题在大屏上的显示状态。
    *   上传、查看和删除知识库文档。
    *   查看和修改当前会话的自定义系统提示词。
*   **`DisplayPage.vue`**: 大屏展示页面，实时显示演讲者选中的问题。
*   **`SessionSetup.vue`**: 用于生成新会话 ID 和访问链接/二维码的起始页面。

## 🔄 使用流程

1.  **创建会话**: 访问前端主页，生成一个新的会话 ID 和链接。
2.  **分享链接/二维码**: 将观众提问页面的链接或二维码分享给观众。
3.  **准备知识库 (可选)**: 演讲者访问控制台页面，上传相关的 PDF/DOCX/TXT 文档。
4.  **自定义提示词 (可选)**: 演讲者在控制台修改通用提示词或知识库问答提示词模板。
5.  **开始互动**:
    *   观众通过链接或扫码进入提问页面提交问题。
    *   演讲者在控制台监控问题列表，查看 AI 建议和知识库回答。
    *   大屏幕打开展示页面 (`DisplayPage`)。
    *   演讲者选择问题，点击“显示问题”将其推送到大屏幕。
    *   演讲者可以标记问题状态或删除不合适的问题。

## ⚙️ 配置详解

后端服务依赖以下环境变量进行配置 (详见 `backend/config/config.go.example`):

*   **数据库**: `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`
*   **OpenAI**: `OPENAI_API_KEYs`, `OPENAI_API_URL` (Chat API), `OPENAI_MODEL`, `OPENAI_EMBEDDING_MODEL`
*   **默认提示词**: `GENERIC_SYSTEM_PROMPT`, `KB_SYSTEM_PROMPT`
*   **服务端口**: `SERVER_PORT`

## 🧠 知识库工作原理

1.  **上传与处理**: 演讲者上传文档后，后端会异步提取文本内容，将其分割成较小的文本块 (Chunks)。
2.  **向量化**: 每个文本块通过 OpenAI Embeddings API 转换成向量 (Embedding)，这是一种能代表文本语义的数字表示。
3.  **存储**: 文本块内容和对应的向量存储在数据库的 `document_chunks` 表中。
4.  **检索**: 当收到新问题时，后端同样将问题向量化。然后，通过计算问题向量与数据库中所有文档块向量的余弦相似度，找出与问题最相关的几个文本块。
5.  **生成回答**: 将原始问题和检索到的最相关文本块一起发送给 OpenAI Chat Completions API，并使用特定的系统提示词（优先使用会话自定义提示词，否则使用默认知识库提示词）指导模型生成基于这些信息的回答。

## ⚠️ 注意事项

*   请确保已正确执行数据库脚本 (`database.sql` 和 `knowledge_base_schema.sql`)。
*   需要有效的 OpenAI API Key，并确保账户有足够余额。
*   知识库功能依赖 Embeddings API 和 Chat Completions API。
*   文档处理（特别是向量化）可能需要一些时间，在文档上传后稍等片刻才能用于问答。
*   建议在本地或可控网络环境下部署和使用。

## 🚀 未来改进

*   [ ] 用户认证系统
*   [ ] 问题点赞/排序功能
*   [ ] 更细致的知识库管理（例如按文档查看/管理块）
*   [ ] 支持更多文档格式 (如 Markdown)
*   [ ] 优化向量检索性能 (例如使用专门的向量数据库或索引)
*   [ ] 允许调整 RAG 参数 (如检索的块数量 topK)
*   [ ] 历史记录查询
*   [ ] 多会话管理界面

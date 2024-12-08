# AnyQA Vue 版本快速启动指南

## 1. 配置数据库

与原版相同,首先需要创建并配置 MySQL 数据库:

```sql
CREATE DATABASE IF NOT EXISTS aiqa;
USE aiqa;
CREATE TABLE IF NOT EXISTS questions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    status ENUM('pending', 'showing', 'answered', 'finished') DEFAULT 'pending',
    ai_suggestion TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 2. 配置后端

后端配置保持不变:

```bash
# 安装依赖
go mod init github.com/soaringjerry/AnyQA
go get github.com/gin-gonic/gin
go get github.com/go-sql-driver/mysql
go get github.com/gorilla/websocket
```

## 3. 前端配置 (Vue 版本)

### 3.1 创建 Vue 项目

```bash
# 使用 Vue CLI 创建项目
npm install -g @vue/cli
vue create anyqa-frontend

# 选择配置
- 选择 Vue 3
- 勾选:
  - Babel
  - Router
  - Vuex
  - CSS Pre-processors (Sass/SCSS)
  - Linter / Formatter
```

### 3.2 安装依赖

```bash
cd anyqa-frontend
npm install axios vuex@next vue-router@4
npm install element-plus    # UI 组件库
npm install socket.io-client # WebSocket 客户端
```

### 3.3 项目结构

```
anyqa-frontend/
├── src/
│   ├── pages/
│   │   ├── DisplayPage.vue    # 观众大屏页面
│   │   ├── IndexPage.vue      # 提问页面
│   │   ├── SessionSetup.vue   # 开始创建页面
│   │   └── PresenterPage.vue  # 演讲者控制台
│   ├── router/
│   │   └── index.js           # 路由配置
│   ├── components/            # 可复用组件
│   └── config/               # 配置文件
```

### 3.4 配置环境变量

创建 `.env.development` 和 `.env.production`:

```env
VUE_APP_API_URL=http://localhost:8080
VUE_APP_WS_URL=ws://localhost:8080/ws
```

## 4. 运行系统

```bash
# 启动后端
cd backend
go run main.go

# 启动前端开发服务器
cd anyqa-frontend
npm run serve

# 访问地址
- 观众页面: http://localhost:8080/#/
- 演讲者控制台: http://localhost:8080/#/presenter
- 大屏显示: http://localhost:8080/#/display
- 开始页面: http://localhost:8080/#/setup
```

## 5. 生产环境部署

```bash
# 构建前端
cd anyqa-frontend
npm run build

# 部署
将 dist 目录下的文件部署到 Web 服务器
配置 nginx 或其他 Web 服务器处理路由
```

## 注意事项

1. 确保后端 API 已开启 CORS 支持
2. WebSocket 连接需要在 Vue 组件内正确管理生命周期
3. 路由模式建议使用 history 模式,需要相应配置服务器
4. 生产环境部署时注意配置环境变量

# 演讲互动问答系统

一个简单的演讲互动问答系统，支持观众实时提问、AI辅助回答和大屏展示。

## 功能特点

- 观众扫码提问：无需下载APP，扫码即可提交问题
- AI辅助回答：自动生成答案建议，帮助演讲者准备回答
- 实时互动：演讲者可以控制问题显示顺序
- 大屏展示：支持问题在大屏幕上展示

## 技术栈

- 前端：原生HTML/CSS/JavaScript
- Vue重制版前端（目前功能更多，但并不稳定，请尽量先使用原生版本）：Vue 3 + Vite + vue-router + Vuetify + js-yaml（用于解析 YAML 配置文件）
- 后端：Go (Gin框架)
- 数据库：MySQL
- AI：OpenAI API

## 快速开始

1. 配置数据库
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

2. 配置后端及Config
```bash
# 安装依赖
go mod init github.com/soaringjerry/AnyQA
go get github.com/gin-gonic/gin
go get github.com/go-sql-driver/mysql
go get github.com/gorilla/websocket

# 配置frontend和backend里面的config
```

3. 运行系统
```bash
# 启动后端
go run main.go

# 打开前端页面
# 观众页面：index.html
# 演讲者控制台：presenter.html
# 大屏显示：display.html
```

## 页面说明

- index.html: 观众提问页面
- presenter.html: 演讲者控制台，用于管理问题和查看AI建议
- display.html: 大屏显示页面，展示当前问题

## 使用流程

1. 演讲开始前生成二维码（指向index.html）
2. 观众扫码进入提问页面
3. 演讲者打开控制台(presenter.html)监控问题
4. 大屏幕打开display.html展示问题
5. 演讲者可以：
   - 查看问题和AI建议
   - 控制问题显示/隐藏
   - 标记问题状态
   - 删除不合适的问题

## 注意事项

- 需要正确配置数据库连接信息
- 需要有效的OpenAI API Key
- 建议在本地网络环境下使用

## 未来改进

- [ ] 用户认证系统
- [ ] 问题点赞功能
- [ ] 问题分类管理
- [ ] 历史记录查询
- [ ] 多会话支持

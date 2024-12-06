演讲互动问答系统

一个轻量级的实时互动问答系统，观众可通过扫描二维码向演讲者提问，并由 AI 提供回答建议，让演讲更高效有趣。

目录
功能特点
技术栈
快速开始
数据库配置
后端配置
运行系统
页面说明
使用流程
注意事项
未来改进
功能特点
观众扫码提问：无需下载安装，移动端浏览器扫码即可提交问题。
AI辅助回答：利用 OpenAI 接口，为演讲者提供答案建议，节省准备时间。
实时互动：演讲者可在控制台实时查看新问题，决定展示顺序。
大屏展示：将问题投放到大屏幕上，增加互动氛围。
技术栈
前端：原生 HTML / CSS / JavaScript
后端：Go (Gin 框架)
数据库：MySQL
AI：OpenAI API
快速开始
数据库配置
请先在 MySQL 中创建数据库和数据表：

sql
复制代码
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
后端配置
安装 Go 依赖：

bash
复制代码
go mod init qa-system
go get github.com/gin-gonic/gin
go get github.com/go-sql-driver/mysql
go get github.com/gorilla/websocket
设置 OpenAI API Key（在启动服务前通过环境变量 OPENAI_API_KEY 设置，或在代码中进行配置）：

bash
复制代码
export OPENAI_API_KEY="YOUR-OPENAI-API-KEY"
运行系统
bash
复制代码
# 启动后端服务
go run main.go

# 打开前端页面（在浏览器中访问）
# 观众提问页面：index.html
# 演讲者控制台：presenter.html
# 大屏展示页面：display.html
页面说明
index.html：观众提问界面
presenter.html：演讲者控制台，可查看问题列表、AI建议并控制展示顺序
display.html：大屏幕显示当前问题，为现场观众提供可视化的问答互动界面
使用流程
在演讲开始前生成二维码（指向 index.html），观众扫码进入提问界面。
观众实时提交问题，问题将存入数据库。
演讲者在控制台（presenter.html）查看新问题及 AI 建议，并进行状态管理（展示、隐藏、标记等）。
大屏幕使用 display.html 来显示当前精选问题，使现场互动更加直观。
演讲者可根据需要标记问题状态（例如：回答完成后标记为 answered，或将处理完毕的问题标记为 finished）。
注意事项
确保已正确配置数据库连接信息（在代码或环境变量中设置）。
提前在本地或服务器中设置 OPENAI_API_KEY。
建议在内网或特定环境中使用，确保网络与安全策略合规。
未来改进
 用户认证系统（限制提问者或管理员权限）
 问题点赞与排序功能（让观众参与度更高）
 问题分类与筛选（便于管理和回顾）
 历史记录查询和导出（方便后期分析与总结）
 多会话支持（适用于多个演讲场景）
# AnyQA CI/CD 设置指南

本文档详细介绍了如何为 AnyQA 项目（或其 fork）配置和运行持续集成与持续部署 (CI/CD) 流程。该流程使用 GitHub Actions 实现自动化，将应用程序构建为 Docker 镜像，推送到 GitHub Container Registry (GHCR)，并通过 SSH 部署到目标服务器。

## 流程概述

1.  **触发：** 当代码被推送到 `main` 分支时，GitHub Actions 工作流 (`.github/workflows/go.yml`) 会自动触发。
2.  **构建：**
    *   Go 后端代码被编译。
    *   Vue 前端代码被构建。
    *   使用各自的 Dockerfile (`backend/Dockerfile`, `frontend-vue/Dockerfile`) 构建后端和前端的 Docker 镜像。
    *   镜像被打上基于 commit SHA 的标签 (例如 `sha-xxxxxxx`)。
3.  **推送：** 构建好的 Docker 镜像被推送到 GitHub Container Registry (GHCR)。镜像仓库地址基于触发工作流的 GitHub 用户名或组织名 (例如 `ghcr.io/your-username/anyqa_backend:sha-xxxxxxx`)。
4.  **部署：**
    *   工作流通过 SSH 连接到您指定的目标服务器。
    *   确保服务器上的部署目录 (`/root/anyqa_app`) 存在。
    *   将项目根目录下的 `docker-compose.yml` 文件复制到服务器的部署目录。
    *   在服务器上登录 GHCR 以获取拉取私有镜像的权限。
    *   在服务器的部署目录下执行 `docker compose up -d` 命令。
    *   `docker compose` 会自动加载部署目录下的 `.env` 文件获取配置，并使用从 GHCR 拉取的最新镜像启动或更新后端和前端服务。

## 前提条件

### 1. 目标服务器

您需要准备一台可以通过 SSH 访问的目标服务器，并满足以下条件：

*   **SSH 访问权限：** 您需要拥有服务器的 IP 地址或域名、SSH 用户名和密码（或配置 SSH 密钥）。
*   **Docker 安装：** 服务器上必须安装 Docker 引擎。请参考 [Docker 官方文档](https://docs.docker.com/engine/install/) 进行安装。
*   **Docker Compose 安装：** 服务器上必须安装 Docker Compose。请参考 [Docker 官方文档](https://docs.docker.com/compose/install/) 进行安装。
*   **部署目录：** CI/CD 脚本默认使用 `/root/anyqa_app` 作为部署目录。您可以根据需要修改 `.github/workflows/go.yml` 文件中的路径。脚本会自动创建此目录（如果不存在）。
*   **`.env` 文件：** 您需要在服务器的部署目录下（例如 `/root/anyqa_app/`) **手动创建并配置**一个名为 `.env` 的文件。此文件包含应用程序运行所需的敏感配置，**不会**被提交到代码仓库。文件内容应类似：
    ```dotenv
    # .env file on the server (/root/anyqa_app/.env)

    # --- Database Configuration ---
    DB_USER=your_db_user
    DB_PASSWORD=your_db_password
    DB_HOST=your_db_host # Can be localhost if DB runs on the same server, or external IP/domain
    DB_PORT=3306
    DB_NAME=aiqa

    # --- OpenAI Configuration ---
    OPENAI_API_KEYs=your_openai_api_key
    # Optional: Override default OpenAI settings if needed
    # OPENAI_API_URL=...
    # OPENAI_MODEL=...
    # OPENAI_EMBEDDING_MODEL=...

    # --- System Prompts (Optional Overrides) ---
    # GENERIC_SYSTEM_PROMPT="..."
    # KB_SYSTEM_PROMPT="..."

    # --- Server Configuration ---
    # SERVER_PORT is typically managed by docker-compose ports, but can be set if needed internally
    ```
    **重要：** 请确保将上述示例中的占位符替换为您自己的真实配置。

### 2. GitHub Secrets

您需要在您的 GitHub 仓库（或 fork 后的仓库）中配置以下 Secrets，以便 CI/CD 工作流能够安全地访问您的服务器：

*   `SERVER_HOST`: 您的目标服务器的 IP 地址或域名。
*   `SERVER_USER`: 用于 SSH 连接的用户名。
*   `SSH_PASSWORD`: 用于 SSH 连接的密码。 (注意：为了更高安全性，推荐使用 SSH 密钥对进行认证，并相应修改 `.github/workflows/go.yml` 文件中的 `ssh-action` 和 `scp-action` 配置。)

**如何配置 Secrets:**

1.  进入您的 GitHub 仓库页面。
2.  点击 "Settings" 标签页。
3.  在左侧菜单中，选择 "Secrets and variables" -> "Actions"。
4.  点击 "New repository secret" 按钮。
5.  逐个添加上述三个 Secret（`SERVER_HOST`, `SERVER_USER`, `SSH_PASSWORD`）及其对应的值。

## 运行 CI/CD

完成上述前提条件配置后：

1.  对代码进行修改。
2.  将修改提交 (commit) 并推送 (push) 到您的仓库的 `main` 分支。
3.  GitHub Actions 将自动检测到推送，并开始运行 CI/CD 工作流。
4.  您可以在仓库的 "Actions" 标签页下查看工作流的运行状态和日志。

## 访问部署的应用

部署成功后：

*   **后端服务：** 默认监听在服务器的 `8080` 端口。
*   **前端服务：** 默认监听在服务器的 `8081` 端口（因为 80 端口可能已被占用）。您可以通过 `http://<您的服务器IP或域名>:8081` 在浏览器中访问前端界面。

---

## 外部反向代理配置 (可选，如果使用域名)

如果您希望通过**单个域名**（例如 `https://vueqa.jerryspace.one`）来访问前端界面并让前端调用后端 API，您需要在您的 Docker 服务**前面**设置一个**外部反向代理**（例如服务器上的主 Nginx、Cloudflare 等）。

这是因为浏览器访问该域名时，外部代理需要根据请求路径将流量转发到不同的内部服务：
*   访问前端界面的请求（路径 `/`, `/presenter` 等）需要转发到 **前端 Docker 服务**（默认映射到宿主机的 `11451` 端口）。
*   前端发出的 API 请求（路径 `/api/...`）需要转发到 **后端 Docker 服务**（默认映射到宿主机的 `18080` 端口）。

**重要配置：**

您需要在**外部反向代理**（处理您域名的那个）上配置类似以下的规则（以 Nginx 为例）：

```nginx
# 外部 Nginx 配置示例 (vueqa.jerryspace.one.conf)

# 优先处理 API 请求，转发到后端服务 (端口 18080)
location /api/ {
    proxy_pass http://127.0.0.1:18080; # 指向后端服务端口
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket 支持
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}

# 处理其他所有请求，转发到前端服务 (端口 11451)
location / {
    proxy_pass http://127.0.0.1:11451; # 指向前端服务端口
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

**注意：如果您使用网站管理面板（如宝塔面板）：**

*   面板可能不允许在 UI 上同时设置 `/` 和 `/api` 的代理。
*   **推荐方法**：
    1.  在面板 UI 中**只设置**根路径 `/` 的反向代理，指向前端端口 `11451`。
    2.  找到面板中该站点的“配置文件”或“自定义配置”编辑区域。
    3.  将上面示例中的 `location /api/ { ... }` 块**完整粘贴**到自定义配置区域。
    4.  保存面板配置。面板会自动处理并重载 Nginx。
*   **备选方法（不推荐，易被覆盖）**：在面板设置 `/` 代理后，手动编辑面板生成的 Nginx 配置文件，在 `location /` 块**上方**插入 `location /api/` 块。

**请务必根据您实际使用的外部反向代理软件（Nginx, Cloudflare, Apache, Caddy 等）调整具体配置语法。**

---
祝您配置顺利！
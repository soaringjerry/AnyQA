# GitHub Actions CI/CD 计划

## 1. 目标

自动化测试 Go 后端和 Vue 前端，构建整个应用的 Docker 镜像，并在代码合并到 `main` 分支时自动部署到你的服务器。

## 2. 触发条件

- `push` 到 `main` 分支。

## 3. CI (持续集成) 阶段

1.  **检出代码:** 获取最新的代码 (`actions/checkout@v4`)。
2.  **设置环境:**
    -   配置 Go 环境 (`actions/setup-go@v4`)。
    -   配置 Node.js 环境 (`actions/setup-node@v4`)。
3.  **后端测试:** 运行 Go 项目的测试 (`cd backend && go test ./...`)。
4.  **(可选) 前端测试:**
    -   安装前端依赖 (`cd frontend-vue && npm install`)。
    -   运行 Vue 项目的测试 (`cd frontend-vue && npm run test`)。*(待实施阶段确认)*
5.  **构建前端:** 打包 Vue 应用 (`cd frontend-vue && npm run build`)。
6.  **构建后端:** 编译 Go 应用 (`cd backend && go build ...`)。
7.  **构建 Docker 镜像:**
    -   登录 Docker Registry (GHCR/Docker Hub) (`docker/login-action@v3`)。
    -   使用 `docker-compose.yml` 或专门的 `Dockerfile` 构建包含前后端的应用镜像 (`docker/build-push-action@v5`)。*(具体实现细节将在 Code 模式下处理)*
8.  **推送 Docker 镜像:** 将构建好的镜像推送到容器镜像仓库。

## 4. CD (持续部署) 阶段

1.  **SSH 连接服务器:** 使用 `appleboy/ssh-action@master` 或类似 Action，配合 GitHub Secrets (`SERVER_HOST`, `SERVER_USER`, `SSH_PRIVATE_KEY`) 安全连接。
2.  **服务器部署 (通过 SSH 执行):**
    ```bash
    cd /path/to/your/app  # 进入服务器上存放 docker-compose.yml 的目录
    docker-compose pull   # 拉取 CI 阶段推送的最新镜像
    docker-compose up -d --remove-orphans # 使用新镜像重新启动服务
    docker image prune -f # (可选) 清理旧的无用镜像
    ```

## 5. 关键准备工作

-   **GitHub Secrets 配置:**
    -   `SERVER_HOST`: 服务器 IP 或域名。
    -   `SERVER_USER`: SSH 用户名。
    -   `SSH_PRIVATE_KEY`: SSH 私钥。
    -   `GHCR_TOKEN` (推荐) 或 `DOCKER_USERNAME` / `DOCKER_PASSWORD`: 镜像仓库凭证。
-   **Dockerfile/Compose 文件:** 检查并优化 `docker-compose.yml` 或创建 `Dockerfile` 以便进行生产环境的镜像构建 *(将在 Code 模式下进行)*。
-   **服务器环境:** 确保目标服务器已安装 Docker 和 Docker Compose。

## 6. 可视化流程 (Mermaid)

```mermaid
graph LR
    A[代码合并到 main 分支] --> B{GitHub Actions 触发};
    B --> C[CI: 检出代码];
    C --> D[CI: 设置 Go/Node 环境];
    D --> E[CI: 运行后端测试];
    D --> F[CI: (可选) 运行前端测试];
    E --> G[CI: 构建后端];
    F --> H[CI: 构建前端];
    G & H --> I[CI: 构建 Docker 镜像];
    I --> J[CI: 推送镜像到 GHCR/DockerHub];
    J --> K{CD: SSH 连接服务器};
    K --> L[CD: 服务器拉取最新镜像];
    L --> M[CD: 服务器重启 Docker Compose 服务];
    M --> N[部署完成];

    subgraph CI 阶段
        direction LR
        C; D; E; F; G; H; I; J;
    end

    subgraph CD 阶段
        direction LR
        K; L; M;
    end
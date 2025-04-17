# frontend-vue/Dockerfile

# ---- Build Stage ----
# 使用 Node.js Alpine 镜像作为构建环境 (选择一个合适的 LTS 版本)
FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package.json 和 package-lock.json (或 yarn.lock)
COPY package*.json ./

# 安装依赖项。利用 Docker 缓存
RUN npm install

# 复制前端源代码
COPY . .

# 执行构建命令
RUN npm run build

# ---- Run Stage ----
# 使用 Nginx Alpine 镜像作为运行环境
FROM nginx:stable-alpine

# 复制自定义的 Nginx 配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从构建阶段复制构建好的静态文件到 Nginx 的 web 根目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 暴露 Nginx 默认监听的端口
EXPOSE 80

# Nginx 镜像默认会启动 Nginx 服务，无需指定 ENTRYPOINT 或 CMD
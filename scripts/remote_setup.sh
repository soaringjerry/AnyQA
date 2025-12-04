#!/usr/bin/env bash
#
# AnyQA 一键部署脚本
# 用法: bash -c "$(curl -fsSL https://raw.githubusercontent.com/soaringjerry/AnyQA/main/scripts/remote_setup.sh)"
#
set -euo pipefail

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
REPO_URL="https://github.com/soaringjerry/AnyQA.git"
REPO_RAW="https://raw.githubusercontent.com/soaringjerry/AnyQA/main"
DEPLOY_DIR="${DEPLOY_DIR:-/root/anyqa_app}"
GHCR_OWNER="${GHCR_OWNER:-soaringjerry}"

log_info()  { echo -e "${BLUE}[INFO]${NC} $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# 检测系统
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    fi
    echo "$OS"
}

# 检测包管理器
detect_pkg_manager() {
    if command -v apt-get &>/dev/null; then
        echo "apt"
    elif command -v yum &>/dev/null; then
        echo "yum"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    elif command -v apk &>/dev/null; then
        echo "apk"
    else
        echo "unknown"
    fi
}

# 安装 Docker
install_docker() {
    if command -v docker &>/dev/null; then
        log_ok "Docker 已安装: $(docker --version)"
        return 0
    fi

    log_info "安装 Docker..."
    local pkg_mgr=$(detect_pkg_manager)

    case "$pkg_mgr" in
        apt)
            apt-get update -qq
            apt-get install -y -qq curl
            curl -fsSL https://get.docker.com | sh
            ;;
        yum|dnf)
            $pkg_mgr install -y curl
            curl -fsSL https://get.docker.com | sh
            ;;
        *)
            log_error "不支持的包管理器，请手动安装 Docker"
            exit 1
            ;;
    esac

    systemctl enable docker
    systemctl start docker
    log_ok "Docker 安装完成"
}

# 检测 docker compose
check_compose() {
    if docker compose version &>/dev/null 2>&1; then
        echo "docker compose"
    elif command -v docker-compose &>/dev/null; then
        echo "docker-compose"
    else
        log_error "Docker Compose 未安装"
        exit 1
    fi
}

# 交互输入（带默认值）
prompt() {
    local prompt="$1"
    local default="$2"
    local var
    read -r -p "$prompt [$default]: " var
    echo "${var:-$default}"
}

# 密码输入
prompt_secret() {
    local prompt="$1"
    local var
    read -r -s -p "$prompt: " var
    echo
    echo "$var"
}

# 创建目录结构
setup_dirs() {
    log_info "创建目录结构..."
    mkdir -p "$DEPLOY_DIR"/{config/frontend,uploads,logs}
    log_ok "目录创建完成: $DEPLOY_DIR"
}

# 下载配置文件
download_files() {
    log_info "下载配置文件..."

    # docker-compose.yml
    curl -fsSL "$REPO_RAW/docker-compose.yml" -o "$DEPLOY_DIR/docker-compose.yml"

    # schema.sql
    curl -fsSL "$REPO_RAW/schema.sql" -o "$DEPLOY_DIR/schema.sql"

    log_ok "配置文件下载完成"
}

# 登录 GHCR
ghcr_login() {
    log_info "登录 GitHub Container Registry..."

    if docker pull "ghcr.io/$GHCR_OWNER/anyqa_backend:latest" &>/dev/null 2>&1; then
        log_ok "GHCR 已可访问（公开镜像）"
        return 0
    fi

    echo "如果镜像是私有的，需要登录 GHCR"
    local need_login=$(prompt "是否需要登录 GHCR? (y/n)" "n")

    if [[ "$need_login" == "y" ]]; then
        local ghcr_user=$(prompt "GitHub 用户名" "")
        local ghcr_token=$(prompt_secret "GitHub Token (需要 read:packages 权限)")

        if [[ -n "$ghcr_user" && -n "$ghcr_token" ]]; then
            echo "$ghcr_token" | docker login ghcr.io -u "$ghcr_user" --password-stdin
            log_ok "GHCR 登录成功"
        fi
    fi
}

# 配置环境变量
configure_env() {
    log_info "配置环境变量..."

    local env_file="$DEPLOY_DIR/.env"

    # 如果已存在，询问是否覆盖
    if [[ -f "$env_file" ]]; then
        local overwrite=$(prompt "检测到已有配置，是否覆盖? (y/n)" "n")
        if [[ "$overwrite" != "y" ]]; then
            log_info "保留现有配置"
            return 0
        fi
    fi

    echo ""
    echo "=== 镜像配置 ==="
    local image_tag=$(prompt "镜像 Tag (latest 或 sha-xxx)" "latest")

    echo ""
    echo "=== 数据库配置 ==="
    local db_host=$(prompt "数据库地址" "host.docker.internal")
    local db_port=$(prompt "数据库端口" "3306")
    local db_user=$(prompt "数据库用户" "root")
    local db_pass=$(prompt_secret "数据库密码")
    local db_name=$(prompt "数据库名" "aiqa")

    echo ""
    echo "=== OpenAI 配置 ==="
    local openai_key=$(prompt_secret "OpenAI API Key")
    local openai_url=$(prompt "OpenAI API URL" "https://api.openai.com/v1/chat/completions")
    local openai_model=$(prompt "Chat 模型" "gpt-4o")
    local openai_embed=$(prompt "Embedding 模型" "text-embedding-3-large")

    echo ""
    echo "=== 前端配置 ==="
    local public_host=$(prompt "公网访问地址 (域名或IP)" "localhost")
    local api_port=$(prompt "后端 API 端口" "18082")
    local fe_port=$(prompt "前端端口" "11451")
    local protocol=$(prompt "协议 (http/https)" "http")

    # 写入 .env
    cat > "$env_file" <<EOF
# AnyQA 配置文件 - 自动生成于 $(date)

# 镜像
BACKEND_IMAGE=ghcr.io/${GHCR_OWNER}/anyqa_backend:${image_tag}
FRONTEND_IMAGE=ghcr.io/${GHCR_OWNER}/anyqa_frontend:${image_tag}

# 数据库
DB_HOST=${db_host}
DB_PORT=${db_port}
DB_USER=${db_user}
DB_PASSWORD=${db_pass}
DB_NAME=${db_name}

# OpenAI
OPENAI_API_KEYs=${openai_key}
OPENAI_API_URL=${openai_url}
OPENAI_MODEL=${openai_model}
OPENAI_EMBEDDING_MODEL=${openai_embed}

# 服务端口
SERVER_PORT=:8080
EOF

    # 前端配置
    local ws_protocol="ws"
    [[ "$protocol" == "https" ]] && ws_protocol="wss"

    cat > "$DEPLOY_DIR/config/frontend/config.json" <<EOF
{
  "api": { "endpoint": "${protocol}://${public_host}:${api_port}/api" },
  "session": { "id": "default" },
  "ws": { "endpoint": "${ws_protocol}://${public_host}:${api_port}/api/ws" }
}
EOF

    log_ok "配置文件已生成"
}

# 拉取镜像
pull_images() {
    log_info "拉取 Docker 镜像..."
    local compose_cmd=$(check_compose)
    (cd "$DEPLOY_DIR" && $compose_cmd pull)
    log_ok "镜像拉取完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    local compose_cmd=$(check_compose)
    (cd "$DEPLOY_DIR" && $compose_cmd up -d --remove-orphans)
    log_ok "服务已启动"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    local compose_cmd=$(check_compose)
    (cd "$DEPLOY_DIR" && $compose_cmd down)
    log_ok "服务已停止"
}

# 查看状态
show_status() {
    local compose_cmd=$(check_compose)
    echo ""
    echo "=== 容器状态 ==="
    (cd "$DEPLOY_DIR" && $compose_cmd ps)
    echo ""
    echo "=== 最近日志 ==="
    (cd "$DEPLOY_DIR" && $compose_cmd logs --tail=10)
}

# 查看日志
show_logs() {
    local compose_cmd=$(check_compose)
    (cd "$DEPLOY_DIR" && $compose_cmd logs -f --tail=100)
}

# 升级
upgrade() {
    log_info "升级 AnyQA..."

    # 下载最新配置
    download_files

    # 拉取最新镜像
    pull_images

    # 重启服务
    local compose_cmd=$(check_compose)
    (cd "$DEPLOY_DIR" && $compose_cmd up -d --remove-orphans)

    # 清理旧镜像
    docker image prune -f

    log_ok "升级完成"
}

# 修复常见问题
repair() {
    log_info "开始诊断和修复..."

    local issues_found=0

    # 检查 Docker
    if ! command -v docker &>/dev/null; then
        log_warn "Docker 未安装，正在安装..."
        install_docker
        ((issues_found++))
    fi

    # 检查目录
    if [[ ! -d "$DEPLOY_DIR" ]]; then
        log_warn "部署目录不存在，正在创建..."
        setup_dirs
        ((issues_found++))
    fi

    # 检查配置文件
    if [[ ! -f "$DEPLOY_DIR/.env" ]]; then
        log_warn "配置文件不存在"
        ((issues_found++))
    fi

    if [[ ! -f "$DEPLOY_DIR/docker-compose.yml" ]]; then
        log_warn "docker-compose.yml 不存在，正在下载..."
        curl -fsSL "$REPO_RAW/docker-compose.yml" -o "$DEPLOY_DIR/docker-compose.yml"
        ((issues_found++))
    fi

    if [[ ! -f "$DEPLOY_DIR/config/frontend/config.json" ]]; then
        log_warn "前端配置不存在"
        ((issues_found++))
    fi

    # 检查容器状态
    if docker ps | grep -q anyqa_backend; then
        log_ok "后端容器运行中"
    else
        log_warn "后端容器未运行"
        ((issues_found++))
    fi

    if docker ps | grep -q anyqa_frontend; then
        log_ok "前端容器运行中"
    else
        log_warn "前端容器未运行"
        ((issues_found++))
    fi

    # 检查端口
    if command -v lsof &>/dev/null; then
        if lsof -i :18082 &>/dev/null; then
            log_ok "后端端口 18082 已监听"
        else
            log_warn "后端端口 18082 未监听"
        fi

        if lsof -i :11451 &>/dev/null; then
            log_ok "前端端口 11451 已监听"
        else
            log_warn "前端端口 11451 未监听"
        fi
    fi

    echo ""
    if [[ $issues_found -eq 0 ]]; then
        log_ok "未发现问题"
    else
        log_warn "发现 $issues_found 个问题"
        echo ""
        local do_fix=$(prompt "是否尝试自动修复? (y/n)" "y")
        if [[ "$do_fix" == "y" ]]; then
            if [[ ! -f "$DEPLOY_DIR/.env" ]] || [[ ! -f "$DEPLOY_DIR/config/frontend/config.json" ]]; then
                configure_env
            fi
            pull_images
            start_services
            log_ok "修复完成"
        fi
    fi
}

# 完全卸载
uninstall() {
    log_warn "即将卸载 AnyQA..."
    local confirm=$(prompt "确定要卸载吗? 输入 'yes' 确认" "no")

    if [[ "$confirm" != "yes" ]]; then
        log_info "取消卸载"
        return 0
    fi

    stop_services 2>/dev/null || true

    # 删除容器和镜像
    docker rm -f anyqa_backend anyqa_frontend 2>/dev/null || true
    docker rmi ghcr.io/$GHCR_OWNER/anyqa_backend:latest 2>/dev/null || true
    docker rmi ghcr.io/$GHCR_OWNER/anyqa_frontend:latest 2>/dev/null || true

    # 询问是否删除数据
    local del_data=$(prompt "是否删除配置和数据? (y/n)" "n")
    if [[ "$del_data" == "y" ]]; then
        rm -rf "$DEPLOY_DIR"
        log_ok "数据已删除"
    fi

    log_ok "卸载完成"
}

# 显示数据库初始化命令
show_db_init() {
    echo ""
    echo "=== 数据库初始化 ==="
    echo ""
    echo "如果数据库表不存在，请执行以下命令："
    echo ""
    echo "  mysql -u root -p $DB_NAME < $DEPLOY_DIR/schema.sql"
    echo ""
    echo "或者手动执行 schema.sql 中的 SQL 语句"
    echo ""
}

# 显示帮助
show_help() {
    cat <<EOF
AnyQA 一键部署脚本

用法:
  bash -c "\$(curl -fsSL $REPO_RAW/scripts/remote_setup.sh)"
  bash -c "\$(curl -fsSL $REPO_RAW/scripts/remote_setup.sh)" -- [命令]

命令:
  install     完整安装（默认）
  config      仅配置环境变量
  upgrade     升级到最新版本
  repair      诊断并修复问题
  start       启动服务
  stop        停止服务
  restart     重启服务
  status      查看状态
  logs        查看日志
  uninstall   卸载
  help        显示帮助

环境变量:
  DEPLOY_DIR    部署目录 (默认: /root/anyqa_app)
  GHCR_OWNER    GHCR 用户名 (默认: soaringjerry)

示例:
  # 完整安装
  curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash

  # 升级
  curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash -s -- upgrade

  # 修复
  curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash -s -- repair
EOF
}

# 完整安装
full_install() {
    echo ""
    echo "======================================"
    echo "       AnyQA 一键部署脚本"
    echo "======================================"
    echo ""

    install_docker
    setup_dirs
    download_files
    ghcr_login
    configure_env
    pull_images
    start_services

    echo ""
    echo "======================================"
    log_ok "安装完成!"
    echo "======================================"
    echo ""
    echo "前端地址: http://localhost:11451"
    echo "后端 API: http://localhost:18082/api"
    echo ""
    echo "配置目录: $DEPLOY_DIR"
    echo ""
    show_db_init
    echo "常用命令:"
    echo "  查看状态: curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash -s -- status"
    echo "  查看日志: curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash -s -- logs"
    echo "  升级:     curl -fsSL $REPO_RAW/scripts/remote_setup.sh | bash -s -- upgrade"
    echo ""
}

# 主入口
main() {
    local cmd="${1:-install}"

    case "$cmd" in
        install)    full_install ;;
        config)     configure_env ;;
        upgrade)    upgrade ;;
        repair)     repair ;;
        start)      start_services ;;
        stop)       stop_services ;;
        restart)    stop_services; start_services ;;
        status)     show_status ;;
        logs)       show_logs ;;
        uninstall)  uninstall ;;
        help|--help|-h) show_help ;;
        *)
            log_error "未知命令: $cmd"
            show_help
            exit 1
            ;;
    esac
}

main "$@"

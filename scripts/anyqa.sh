#!/usr/bin/env bash
set -euo pipefail

# AnyQA one-click helper
# - Configures backend env (.env)
# - Configures frontend runtime (frontend-vue/public/config.json)
# - Builds and runs via docker compose (local)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.local.yml"
ENV_FILE="${ROOT_DIR}/.env"
FRONTEND_CONFIG_JSON="${ROOT_DIR}/frontend-vue/public/config.json"

detect_compose() {
  if command -v docker &>/dev/null; then
    if docker compose version >/dev/null 2>&1; then
      echo "docker compose"
      return 0
    fi
  fi
  if command -v docker-compose &>/dev/null; then
    echo "docker-compose"
    return 0
  fi
  echo ""; return 1
}

require_tools() {
  local dcmd
  if ! command -v docker >/dev/null 2>&1; then
    echo "[ERROR] docker not found. Please install Docker first." >&2
    exit 1
  fi
  dcmd=$(detect_compose) || true
  if [[ -z "$dcmd" ]]; then
    echo "[ERROR] docker compose (or docker-compose) not found. Please install Docker Compose v2 or v1." >&2
    exit 1
  fi
}

prompt_default() {
  local prompt="$1"; shift
  local default="$1"; shift
  local var
  read -r -p "${prompt} [${default}]: " var || true
  if [[ -z "$var" ]]; then
    echo "$default"
  else
    echo "$var"
  fi
}

write_env_file() {
  local DB_USER="$1" DB_PASSWORD="$2" DB_HOST="$3" DB_PORT="$4" DB_NAME="$5"
  local OPENAI_API_KEYS="$6" OPENAI_API_URL="$7" OPENAI_MODEL="$8" OPENAI_EMBED_MODEL="$9"
  local GENERIC_PROMPT="${10}" KB_PROMPT="${11}" SERVER_PORT="${12}"

  # Quote values for .env (double-quoted, escape \ and ")
  quote_env() { local s="$1"; s="${s//\\/\\\\}"; s="${s//\"/\\\"}"; printf '"%s"' "$s"; }

  cat >"${ENV_FILE}" <<EOF
# AnyQA backend environment
DB_USER=$(quote_env "${DB_USER}")
DB_PASSWORD=$(quote_env "${DB_PASSWORD}")
DB_HOST=$(quote_env "${DB_HOST}")
DB_PORT=$(quote_env "${DB_PORT}")
DB_NAME=$(quote_env "${DB_NAME}")

OPENAI_API_KEYs=$(quote_env "${OPENAI_API_KEYS}")
OPENAI_API_URL=$(quote_env "${OPENAI_API_URL}")
OPENAI_MODEL=$(quote_env "${OPENAI_MODEL}")
OPENAI_EMBEDDING_MODEL=$(quote_env "${OPENAI_EMBED_MODEL}")

# Optional overrides for default prompts
GENERIC_SYSTEM_PROMPT=$(quote_env "${GENERIC_PROMPT}")
KB_SYSTEM_PROMPT=$(quote_env "${KB_PROMPT}")

# Internal server listen (container)
SERVER_PORT=$(quote_env "${SERVER_PORT}")
EOF
  echo "[OK] Wrote ${ENV_FILE}"
}

write_frontend_config() {
  local API_ENDPOINT="$1" WS_ENDPOINT="$2" DEFAULT_SESSION_ID="$3"
  mkdir -p "$(dirname "${FRONTEND_CONFIG_JSON}")"
  cat >"${FRONTEND_CONFIG_JSON}" <<EOF
{
  "api": {
    "endpoint": "${API_ENDPOINT}"
  },
  "session": {
    "id": "${DEFAULT_SESSION_ID}"
  },
  "ws": {
    "endpoint": "${WS_ENDPOINT}"
  }
}
EOF
  echo "[OK] Wrote ${FRONTEND_CONFIG_JSON}"
}

action_config() {
  echo "== AnyQA Configuration =="

  # Backend
  local DB_USER DB_PASSWORD DB_HOST DB_PORT DB_NAME
  DB_USER=$(prompt_default "DB user" "root")
  DB_PASSWORD=$(prompt_default "DB password" "password_placeholder")
  DB_HOST=$(prompt_default "DB host" "127.0.0.1")
  DB_PORT=$(prompt_default "DB port" "3306")
  DB_NAME=$(prompt_default "DB name" "aiqa")

  local OPENAI_API_KEYS OPENAI_API_URL OPENAI_MODEL OPENAI_EMBED_MODEL
  OPENAI_API_KEYS=$(prompt_default "OpenAI API key(s)" "sk-...")
  OPENAI_API_URL=$(prompt_default "OpenAI Chat API URL" "https://api.openai.com/v1/chat/completions")
  OPENAI_MODEL=$(prompt_default "OpenAI chat model" "chatgpt-4o-latest")
  OPENAI_EMBED_MODEL=$(prompt_default "OpenAI embedding model" "text-embedding-3-large")

  local GENERIC_PROMPT KB_PROMPT
  GENERIC_PROMPT=$(prompt_default "Override Generic System Prompt?" "")
  KB_PROMPT=$(prompt_default "Override KB System Prompt template? (MUST contain %s)" "")

  local SERVER_PORT
  SERVER_PORT=$(prompt_default "Backend container listen (keep :8080)" ":8080")

  # Frontend
  local PUBLIC_HOST API_PORT FE_PORT PROTO WS_PROTO
  PUBLIC_HOST=$(prompt_default "Public host (domain or IP) used by browser" "127.0.0.1")
  API_PORT=$(prompt_default "Public API port" "18080")
  FE_PORT=$(prompt_default "Public Frontend port" "11451")
  PROTO=$(prompt_default "HTTP scheme for API (http/https)" "http")
  WS_PROTO=$(prompt_default "WS scheme (ws/wss)" "ws")

  local API_ENDPOINT="${PROTO}://${PUBLIC_HOST}:${API_PORT}/api"
  local WS_ENDPOINT="${WS_PROTO}://${PUBLIC_HOST}:${API_PORT}/api/ws"
  local DEFAULT_SESSION_ID
  DEFAULT_SESSION_ID=$(prompt_default "Default session id (optional)" "alpha")

  write_env_file \
    "$DB_USER" "$DB_PASSWORD" "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    "$OPENAI_API_KEYS" "$OPENAI_API_URL" "$OPENAI_MODEL" "$OPENAI_EMBED_MODEL" \
    "$GENERIC_PROMPT" "$KB_PROMPT" "$SERVER_PORT"

  write_frontend_config "$API_ENDPOINT" "$WS_ENDPOINT" "$DEFAULT_SESSION_ID"
}

action_up() {
  require_tools
  local DCMD
  DCMD=$(detect_compose)
  echo "[INFO] Using compose: ${DCMD}"
  ${DCMD} -f "${COMPOSE_FILE}" up -d --build
  echo "[OK] Services started. Frontend: http://127.0.0.1:11451  Backend API: http://127.0.0.1:18080/api"
}

action_down() {
  require_tools
  local DCMD
  DCMD=$(detect_compose)
  ${DCMD} -f "${COMPOSE_FILE}" down
}

action_logs() {
  require_tools
  local DCMD
  DCMD=$(detect_compose)
  ${DCMD} -f "${COMPOSE_FILE}" logs -f --tail=200
}

action_restart() {
  action_down
  action_up
}

action_init() {
  action_config
  action_up
}

usage() {
  cat <<USAGE
AnyQA helper
Usage: $0 <command>

Commands:
  config    Interactive configuration (.env + frontend config.json)
  init      Run config then build & start containers
  up        Build & start containers (docker-compose.local.yml)
  down      Stop containers
  restart   Restart containers
  logs      Tail service logs

Examples:
  $0 init
  $0 config && $0 up
USAGE
}

main() {
  local cmd="${1:-}"; shift || true
  case "${cmd}" in
    config) action_config ;;
    init) action_init ;;
    up) action_up ;;
    down) action_down ;;
    restart) action_restart ;;
    logs) action_logs ;;
    *) usage; exit 1 ;;
  esac
}

main "$@"

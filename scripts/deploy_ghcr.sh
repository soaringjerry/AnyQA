#!/usr/bin/env bash
set -euo pipefail

# AnyQA GHCR-based deployment helper (server-side)
# - Uses GHCR images produced by CI/CD
# - Prepares deploy dir (default /root/anyqa_app)
# - Creates/updates .env (includes BACKEND_IMAGE/FRONTEND_IMAGE + secrets)
# - Ensures frontend config.json exists (for Nginx container)
# - docker login ghcr.io (optional) and docker compose up -d

DEPLOY_DIR=${DEPLOY_DIR:-/root/anyqa_app}
COMPOSE_SRC=${COMPOSE_SRC:-docker-compose.yml}
COMPOSE_DST="${DEPLOY_DIR}/docker-compose.yml"
ENV_FILE="${DEPLOY_DIR}/.env"
FE_CONF_DIR="${DEPLOY_DIR}/config/frontend"
FE_CONF_PATH="${FE_CONF_DIR}/config.json"

detect_compose() {
  if command -v docker &>/dev/null; then
    if docker compose version >/dev/null 2>&1; then
      echo "docker compose"; return 0
    fi
  fi
  if command -v docker-compose &>/dev/null; then
    echo "docker-compose"; return 0
  fi
  echo ""; return 1
}

require_tools() {
  local dcmd
  if ! command -v docker >/dev/null 2>&1; then
    echo "[ERROR] docker not found." >&2; exit 1
  fi
  dcmd=$(detect_compose) || true
  if [[ -z "$dcmd" ]]; then
    echo "[ERROR] docker compose not found (v2 or v1)." >&2; exit 1
  fi
}

prompt_default() {
  local prompt="$1"; shift
  local default="$1"; shift
  local var
  read -r -p "${prompt} [${default}]: " var || true
  if [[ -z "$var" ]]; then echo "$default"; else echo "$var"; fi
}

write_env() {
  local BACKEND_IMAGE="$1" FRONTEND_IMAGE="$2"
  local DB_USER="$3" DB_PASSWORD="$4" DB_HOST="$5" DB_PORT="$6" DB_NAME="$7"
  local OPENAI_API_KEYS="$8" OPENAI_API_URL="$9" OPENAI_MODEL="${10}" OPENAI_EMBED_MODEL="${11}"
  local GENERIC_PROMPT="${12}" KB_PROMPT="${13}" SERVER_PORT="${14}"

  mkdir -p "${DEPLOY_DIR}"

  quote_env() { local s="$1"; s="${s//\\/\\\\}"; s="${s//\"/\\\"}"; printf '"%s"' "$s"; }

  cat >"${ENV_FILE}" <<EOF
# Images (from GHCR)
BACKEND_IMAGE=${BACKEND_IMAGE}
FRONTEND_IMAGE=${FRONTEND_IMAGE}

# Backend environment
DB_USER=$(quote_env "${DB_USER}")
DB_PASSWORD=$(quote_env "${DB_PASSWORD}")
DB_HOST=$(quote_env "${DB_HOST}")
DB_PORT=$(quote_env "${DB_PORT}")
DB_NAME=$(quote_env "${DB_NAME}")

OPENAI_API_KEYs=$(quote_env "${OPENAI_API_KEYS}")
OPENAI_API_URL=$(quote_env "${OPENAI_API_URL}")
OPENAI_MODEL=$(quote_env "${OPENAI_MODEL}")
OPENAI_EMBEDDING_MODEL=$(quote_env "${OPENAI_EMBED_MODEL}")

GENERIC_SYSTEM_PROMPT=$(quote_env "${GENERIC_PROMPT}")
KB_SYSTEM_PROMPT=$(quote_env "${KB_PROMPT}")

SERVER_PORT=$(quote_env "${SERVER_PORT}")
EOF
  echo "[OK] Wrote ${ENV_FILE}"
}

write_frontend_config() {
  local API_ENDPOINT="$1" WS_ENDPOINT="$2" DEFAULT_SESSION_ID="$3"
  mkdir -p "${FE_CONF_DIR}"
  cat >"${FE_CONF_PATH}" <<EOF
{
  "api": { "endpoint": "${API_ENDPOINT}" },
  "session": { "id": "${DEFAULT_SESSION_ID}" },
  "ws": { "endpoint": "${WS_ENDPOINT}" }
}
EOF
  echo "[OK] Wrote ${FE_CONF_PATH}"
}

action_login() {
  require_tools
  local user token
  user=$(prompt_default "GHCR username (GitHub handle/org)" "")
  read -r -s -p "GHCR token (PAT with read:packages): " token || true; echo
  if [[ -z "$user" || -z "$token" ]]; then
    echo "[ERROR] username or token empty." >&2; exit 1
  fi
  echo "$token" | docker login ghcr.io -u "$user" --password-stdin
}

action_config() {
  echo "== Configure images and secrets =="
  local GHCR_OWNER BACKEND_REPO FRONTEND_REPO TAG
  GHCR_OWNER=$(prompt_default "GHCR owner (e.g., your-username)" "your-username")
  BACKEND_REPO=$(prompt_default "Backend image repo name" "anyqa_backend")
  FRONTEND_REPO=$(prompt_default "Frontend image repo name" "anyqa_frontend")
  TAG=$(prompt_default "Image tag (e.g., latest or sha-xxxxx)" "latest")

  local BACKEND_IMAGE="ghcr.io/${GHCR_OWNER}/${BACKEND_REPO}:${TAG}"
  local FRONTEND_IMAGE="ghcr.io/${GHCR_OWNER}/${FRONTEND_REPO}:${TAG}"

  echo "-- Backend image:  ${BACKEND_IMAGE}"
  echo "-- Frontend image: ${FRONTEND_IMAGE}"

  # Backend env
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

  local GENERIC_PROMPT KB_PROMPT SERVER_PORT
  GENERIC_PROMPT=$(prompt_default "Override Generic System Prompt?" "")
  KB_PROMPT=$(prompt_default "Override KB System Prompt template? (must contain %s)" "")
  SERVER_PORT=$(prompt_default "Backend container listen (keep :8080)" ":8080")

  # Frontend runtime config
  local PUBLIC_HOST API_PORT FE_PORT PROTO WS_PROTO DEFAULT_SESSION_ID
  PUBLIC_HOST=$(prompt_default "Public host (domain or IP) for browser" "your.domain.or.ip")
  API_PORT=$(prompt_default "Public API port" "18080")
  FE_PORT=$(prompt_default "Public Frontend port" "11451")
  PROTO=$(prompt_default "HTTP scheme (http/https) for API" "http")
  WS_PROTO=$(prompt_default "WS scheme (ws/wss)" "ws")
  DEFAULT_SESSION_ID=$(prompt_default "Default session id (optional)" "alpha")

  local API_ENDPOINT="${PROTO}://${PUBLIC_HOST}:${API_PORT}/api"
  local WS_ENDPOINT="${WS_PROTO}://${PUBLIC_HOST}:${API_PORT}/api/ws"

  mkdir -p "${DEPLOY_DIR}"
  cp -f "${COMPOSE_SRC}" "${COMPOSE_DST}"
  echo "[OK] Copied compose to ${COMPOSE_DST}"

  write_env \
    "$BACKEND_IMAGE" "$FRONTEND_IMAGE" \
    "$DB_USER" "$DB_PASSWORD" "$DB_HOST" "$DB_PORT" "$DB_NAME" \
    "$OPENAI_API_KEYS" "$OPENAI_API_URL" "$OPENAI_MODEL" "$OPENAI_EMBED_MODEL" \
    "$GENERIC_PROMPT" "$KB_PROMPT" "$SERVER_PORT"

  write_frontend_config "$API_ENDPOINT" "$WS_ENDPOINT" "$DEFAULT_SESSION_ID"

  echo "[OK] Config completed. Deploy dir: ${DEPLOY_DIR}"
}

action_up() {
  require_tools
  local DCMD; DCMD=$(detect_compose)
  (cd "${DEPLOY_DIR}" && ${DCMD} up -d)
}

action_down() {
  require_tools
  local DCMD; DCMD=$(detect_compose)
  (cd "${DEPLOY_DIR}" && ${DCMD} down)
}

action_pull() {
  require_tools
  local DCMD; DCMD=$(detect_compose)
  (cd "${DEPLOY_DIR}" && ${DCMD} pull)
}

action_update() {
  action_pull
  require_tools
  local DCMD; DCMD=$(detect_compose)
  (cd "${DEPLOY_DIR}" && ${DCMD} up -d --remove-orphans)
}

action_logs() {
  require_tools
  local DCMD; DCMD=$(detect_compose)
  (cd "${DEPLOY_DIR}" && ${DCMD} logs -f --tail=200)
}

usage() {
  cat <<USAGE
AnyQA GHCR deploy helper (server-side)
Environment (optional):
  DEPLOY_DIR=/root/anyqa_app   COMPOSE_SRC=docker-compose.yml

Commands:
  login     Login to ghcr.io (docker login)
  config    Interactive config (.env + frontend config.json + copy compose)
  up        Start services in background (compose up -d)
  down      Stop services
  pull      Pull latest images (from GHCR)
  update    Pull and restart with new images
  logs      Tail logs

Typical flow (on server):
  chmod +x scripts/deploy_ghcr.sh
  ./scripts/deploy_ghcr.sh login   # once per server (with GHCR token)
  ./scripts/deploy_ghcr.sh config  # set image tags + secrets
  ./scripts/deploy_ghcr.sh up      # start
  # for upgrades after CI builds new tag:
  ./scripts/deploy_ghcr.sh update
USAGE
}

main() {
  local cmd="${1:-}"; shift || true
  case "$cmd" in
    login) action_login ;;
    config) action_config ;;
    up) action_up ;;
    down) action_down ;;
    pull) action_pull ;;
    update) action_update ;;
    logs) action_logs ;;
    *) usage; exit 1 ;;
  esac
}

main "$@"


#!/usr/bin/env bash

set -Eeuo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
backend_root="$project_root/backend"
frontend_root="$project_root/frontend"
service="${1:-all}"
processes=()

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "缺少命令：$1" >&2
    exit 1
  fi
}

cleanup() {
  status=$?
  trap - EXIT INT TERM
  for process_id in "${processes[@]}"; do
    kill "$process_id" >/dev/null 2>&1 || true
  done
  for process_id in "${processes[@]}"; do
    wait "$process_id" >/dev/null 2>&1 || true
  done
  exit "$status"
}

start_backend() {
  if command -v air >/dev/null 2>&1; then
    echo "启动后端（Air 热更新）"
    (cd "$backend_root" && exec air -c .air.toml) &
  elif command -v go >/dev/null 2>&1; then
    echo "未检测到 Air，使用 go run 启动后端"
    (cd "$backend_root" && exec go run ./cmd) &
  elif command -v mise >/dev/null 2>&1; then
    echo "未检测到 Air，使用 mise 管理的 Go 启动后端"
    (cd "$backend_root" && exec mise exec -- go run ./cmd) &
  else
    echo "缺少命令：go（也未检测到 mise）" >&2
    exit 1
  fi
  processes+=("$!")
}

start_frontend() {
  require_command pnpm
  if [[ ! -d "$frontend_root/node_modules" ]]; then
    echo "安装前端依赖"
    (cd "$frontend_root" && pnpm install --frozen-lockfile)
  fi
  echo "启动前端（Vite）"
  (cd "$frontend_root" && exec pnpm dev) &
  processes+=("$!")
}

case "$service" in
  all)
    start_backend
    start_frontend
    ;;
  backend)
    start_backend
    ;;
  frontend)
    start_frontend
    ;;
  *)
    echo "用法：./scripts/dev-server.sh [all|backend|frontend]" >&2
    exit 2
    ;;
esac

trap cleanup EXIT INT TERM
echo "开发服务已启动，按 Ctrl+C 停止"

while true; do
  for process_id in "${processes[@]}"; do
    if ! kill -0 "$process_id" >/dev/null 2>&1; then
      set +e
      wait "$process_id"
      process_status=$?
      set -e
      exit "$process_status"
    fi
  done
  sleep 1
done

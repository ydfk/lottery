#!/usr/bin/env bash

set -Eeuo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
build_version="${BUILD_VERSION:-${APP_VERSION:-dev-local}}"

if ! command -v pnpm >/dev/null 2>&1; then
  echo "缺少命令：pnpm" >&2
  exit 1
fi

if command -v go >/dev/null 2>&1; then
  go_runner=(go)
elif command -v mise >/dev/null 2>&1; then
  go_runner=(mise exec -- go)
else
  echo "缺少命令：go（也未检测到 mise）" >&2
  exit 1
fi

echo "安装并构建前端"
(
  cd "$project_root/frontend"
  pnpm install --frozen-lockfile
  VITE_APP_VERSION="$build_version" pnpm build
)

binary_name="lottery"
if [[ "$(uname -s)" == MINGW* || "$(uname -s)" == MSYS* ]]; then
  binary_name="lottery.exe"
fi

echo "构建后端"
(
  cd "$project_root/backend"
  mkdir -p bin
  "${go_runner[@]}" mod download
  CGO_ENABLED=0 "${go_runner[@]}" build -o "bin/$binary_name" ./cmd
)

echo "构建完成：frontend/dist，backend/bin/$binary_name"

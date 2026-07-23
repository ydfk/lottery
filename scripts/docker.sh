#!/usr/bin/env bash

set -Eeuo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/.." && pwd)"
action="${1:-help}"
image_name="${DOCKER_IMAGE_NAME:-ydfk/lottery}"
image_tag="${DOCKER_IMAGE_TAG:-${BUILD_VERSION:-latest}}"

if [[ "$action" == "help" || "$action" == "-h" || "$action" == "--help" ]]; then
  echo "用法：./scripts/docker.sh <build|up|down|logs|push>"
  echo "环境变量：DOCKER_IMAGE_NAME、DOCKER_IMAGE_TAG、BUILD_VERSION"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "缺少命令：docker" >&2
  exit 1
fi

image_tags=("$image_name:$image_tag")
if [[ "$image_tag" != "latest" ]]; then
  image_tags+=("$image_name:latest")
fi

case "$action" in
  build)
    arguments=(build -f Dockerfile --build-arg "APP_VERSION=$image_tag")
    for image in "${image_tags[@]}"; do
      arguments+=(-t "$image")
    done
    (cd "$project_root" && docker "${arguments[@]}" .)
    ;;
  up)
    (cd "$project_root" && docker compose up -d --build)
    ;;
  down)
    (cd "$project_root" && docker compose down)
    ;;
  logs)
    (cd "$project_root" && docker compose logs -f --tail=100)
    ;;
  push)
    for image in "${image_tags[@]}"; do
      docker push "$image"
    done
    ;;
  *)
    echo "未知 Docker 操作：$action" >&2
    exit 2
    ;;
esac

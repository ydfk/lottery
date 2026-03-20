# 整个项目单镜像发布。

FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend
ARG APP_VERSION=dev-local
ENV VITE_APP_VERSION=$APP_VERSION

RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm build

FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=backend-builder /app/backend/main /app/main
COPY --from=backend-builder /app/backend/config /app/config
COPY --from=backend-builder /app/backend/docs /app/docs
COPY --from=frontend-builder /app/frontend/dist /app/web

RUN mkdir -p /app/config /app/data /app/log

ENV TZ=Asia/Shanghai

EXPOSE 25610

CMD ["./main"]

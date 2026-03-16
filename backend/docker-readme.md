# Docker 部署说明

本项目提供了完整的 Docker 部署配置，包括 Dockerfile、.dockerignore 和 docker-compose.yml 文件。

## 构建和运行

### 使用 Docker 构建并运行

```bash
# 构建镜像
docker build -t go-fiber-starter .

# 运行容器
docker run -d -p 25610:25610 --name go-fiber-api go-fiber-starter
```

### 使用 Docker Compose 构建并运行（推荐）

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 配置说明

- 应用默认监听 25610 端口
- 数据存储在 `/app/data` 目录下
- 日志存储在 `/app/log` 目录下
- 配置文件位于 `/app/config` 目录下

## 持久化数据

Docker Compose 配置中已经设置了持久化卷：

- `app-data`: 保存应用数据（如 SQLite 数据库）
- `app-logs`: 保存应用日志

如果需要备份这些数据，可以使用 Docker 卷备份命令。

## 环境变量

您可以在 docker-compose.yml 的 environment 部分添加环境变量来覆盖配置。

## 自定义配置

如果需要自定义配置，您可以：

1. 修改本地 config 目录下的配置文件，然后重新构建镜像
2. 或者通过卷挂载的方式直接替换容器中的配置文件：

```yaml
volumes:
  - ./custom-config/config.yaml:/app/config/config.yaml
```

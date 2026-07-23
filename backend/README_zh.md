# Go Fiber Starter

一个基于 [Fiber](https://github.com/gofiber/fiber) 框架的 Go 语言 API 项目启动模板，专为快速开发、高性能 API 服务而设计。

## 项目特点

- 🚀 基于 Go Fiber 框架，提供极快的 HTTP 性能
- 📝 集成 Swagger 文档，API 一目了然
- 🔐 内置 JWT 认证系统
- 📦 使用 SQLite 作为数据库，简单易用
- 🔄 自动数据库迁移功能
- 📊 优雅的日志处理机制
- 🛠️ 完整的错误处理中间件
- 🐳 Docker 支持，一键部署

## 项目结构

```
go-fiber-starter/
├── cmd/                     # 应用入口
│   ├── api.go               # API 服务配置
│   └── main.go              # 主程序入口
├── config/                  # 配置文件
│   └── config.yaml          # 应用配置
├── data/                    # 数据存储
│   └── db.sqlite            # SQLite 数据库文件
├── docs/                    # Swagger 文档
│   ├── docs.go              # 自动生成的文档代码
│   ├── swagger.json         # Swagger JSON 配置
│   └── swagger.yaml         # Swagger YAML 配置
├── internal/                # 内部应用代码
│   ├── api/                 # API 处理器
│   │   ├── auth/            # 认证相关API
│   │   │   ├── handler.go   # 认证处理函数
│   │   │   └── router.go    # 认证路由
│   │   └── response/        # 响应处理
│   │       └── response.go  # 响应工具函数
│   ├── middleware/          # 中间件
│   │   └── middleware.go    # 全局中间件
│   ├── model/               # 数据模型
│   │   ├── base/            # 基础模型
│   │   │   └── base.go      # 模型基类
│   │   └── user/            # 用户模型
│   │       └── user.go      # 用户结构体
│   └── service/             # 业务逻辑层
│       └── user.go          # 用户服务
├── log/                     # 日志文件
│   └── log.json             # JSON格式日志
├── pkg/                     # 公共包
│   ├── config/              # 配置处理
│   │   └── config.go        # 配置加载逻辑
│   ├── db/                  # 数据库操作
│   │   ├── db.go            # 数据库连接
│   │   ├── migrate.go       # 数据库迁移
│   │   └── user.go          # 用户数据库操作
│   ├── logger/              # 日志处理
│   │   └── logger.go        # 日志配置
│   └── util/                # 工具函数
│       └── file.go          # 文件操作工具
├── .dockerignore            # Docker忽略文件
├── docker-compose.yml       # Docker Compose配置
├── Dockerfile               # Docker构建文件
├── go.mod                   # Go模块文件
├── go.sum                   # Go依赖校验
└── README.md                # 项目说明
```

## 快速开始

### 准备工作

1. 安装 [Go](https://golang.org/dl/) (版本 1.24 或更高)
2. 克隆本仓库

```bash
git clone https://github.com/your-username/go-fiber-starter.git
cd go-fiber-starter
```

### 本地运行

1. 安装依赖

```bash
go mod download
```

2. 运行应用

```bash
go run ./cmd
```

3. 访问应用

API 服务默认运行在 `http://localhost:25610`

Swagger 文档可通过 `http://localhost:25610/swagger/` 访问

### 项目脚本

跨平台脚本统一放在仓库根目录的 `scripts/`：

```bash
./scripts/dev-server.sh backend
./scripts/build.sh
```

Windows PowerShell 在仓库根目录使用 `.\scripts\dev-server.ps1 backend` 和 `.\scripts\build.ps1`。

### 运行测试

```bash
go test ./...
```

认证相关的 HTTP 测试使用内存 SQLite，不会修改 `data/db.sqlite`。

### 使用 Docker 运行

1. 构建并启动容器

```bash
./scripts/docker.sh up
```

2. 访问应用

API 服务默认运行在 `http://localhost:25610`

## API 文档

本项目使用 Swagger 自动生成 API 文档。启动应用后，访问 `/swagger/` 路径即可查看完整的 API 文档。

## 主要 API 端点

- **认证相关**

  - `POST /register` - 用户注册
  - `POST /login` - 用户登录

- **用户相关**
  - `GET /api/user/profile` - 获取用户资料 (需要认证)

## 配置

配置文件位于 `config/config.yaml`，主要配置项包括：

```yaml
app:
  port: "25610" # 应用端口
  env: "development" # 环境设置 (development/production)
jwt:
  secret: "your-secret" # JWT密钥 (生产环境建议使用环境变量)
  expiration: 86400 # Token有效期(秒)
database:
  path: "data/db.sqlite" # SQLite数据库路径
```

## 目录结构说明

- `cmd/`: 应用入口点
- `config/`: 配置文件
- `docs/`: Swagger 文档
- `internal/`: 内部应用代码，不对外暴露
  - `api/`: API 处理器和路由
  - `middleware/`: 中间件
  - `model/`: 数据模型
  - `service/`: 业务逻辑
- `pkg/`: 公共包，可以被外部引用
  - `config/`: 配置处理
  - `db/`: 数据库操作
  - `logger/`: 日志处理
  - `util/`: 工具函数

## Docker 部署

项目提供了 Docker 部署相关文件：

- `Dockerfile`: 用于构建 Docker 镜像
- `docker-compose.yml`: 用于 Docker Compose 部署
- `.dockerignore`: 排除不需要的文件

详细的 Docker 部署说明请参考 [docker-readme.md](docker-readme.md)。

## 开发指南

### 添加新路由

1. 在 `internal/api` 下创建新的包
2. 实现处理函数
3. 在 `cmd/api.go` 中注册路由

### 添加新模型

1. 在 `internal/model` 下创建新的包和模型文件
2. 在 `pkg/db/migrate.go` 中添加模型到自动迁移列表

### 生成 Swagger 文档

使用 [swag](https://github.com/swaggo/swag) 工具更新 API 文档：

```bash
# 安装 swag 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/main.go
```

## 贡献指南

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

Copyright © 2025 ydfk.

## 联系方式

如有任何问题或建议，请通过以下方式联系：

- 项目维护者: ydfk
- 邮箱: [lyh6728326@gmail.com]

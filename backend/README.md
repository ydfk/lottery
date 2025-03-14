# 彩票智能推荐系统

这是一个基于Go语言开发的彩票智能推荐系统后端，使用AI技术来为用户提供彩票号码推荐。

## 项目特点

- 支持多种彩票类型（当前包括双色球和大乐透）
- 基于AI模型的号码推荐
- JWT认证和权限控制
- RESTful API接口
- 使用SQLite本地数据库
- 支持定时任务调度

## 技术栈

- Go 1.24+
- Fiber v2.52+ (Web框架)
- GORM v1.25+ (ORM库)
- SQLite3 (数据库)
- JWT v4 (认证)
- DeepSeek Chat (AI接口)

## 项目结构

```
backend/
├── cmd/                # 命令行应用
│   ├── api.go          # API服务实现
│   └── main.go         # 程序入口
├── config/             # 配置文件目录
│   ├── config.yaml     # 主配置文件
│   └── config.example.yaml # 示例配置文件
├── data/               # 数据文件目录
│   └── lottery.db      # SQLite数据库文件
├── docs/               # 文档目录
│   └── openapi.yaml    # API 文档
├── internal/           # 内部包
│   ├── config/         # 配置处理
│   ├── handlers/       # HTTP处理器
│   │   ├── audit.go    # 审计日志处理
│   │   ├── auth.go     # 认证处理
│   │   ├── common.go   # 通用函数
│   │   ├── draw_result.go # 开奖结果处理
│   │   ├── lottery_type.go # 彩票类型处理
│   │   └── recommendation.go # 推荐处理
│   ├── models/         # 数据模型
│   │   └── models.go   # 模型定义
│   ├── pkg/            # 内部工具包
│   │   ├── ai/         # AI客户端
│   │   ├── database/   # 数据库处理
│   │   ├── logger/     # 日志处理
│   │   ├── middleware/ # 中间件
│   │   └── scheduler/  # 任务调度器
│   └── services/       # 业务服务
│       └── draw.go     # 开奖服务
├── logs/               # 日志文件目录
└── scripts/            # 脚本目录
```

## 配置说明

## 如何编译

### 前提条件

- 安装Go 1.24或更高版本
- 设置好GOPATH和GOROOT环境变量

### 编译步骤

1. 克隆仓库到本地

```bash
git clone https://your-repo/lottery.git
cd lottery/backend
```

2. 安装依赖

```bash
go mod tidy
```

3. 编译程序

```bash
go build -o lottery-backend /cmd
```

## 如何运行

### 直接运行

```bash
go run ./cmd
```

### 运行编译后的程序

```bash
./lottery-backend    # Linux/Mac
lottery-backend.exe  # Windows
```

## 开发说明

- 项目使用Go Modules管理依赖
- 数据库会在首次运行时自动初始化并创建必要的表
- 初始用户和彩票类型会在启动时自动创建

## 注意事项

- 默认配置中的密钥和密码仅用于开发环境，生产环境请务必更改
- 开发环境建议使用DeepSeek Chat免费模型进行测试
- 系统会在启动时自动创建配置文件中定义的用户和彩票类型
- 请确保配置文件中的cron表达式与实际开奖时间相匹配
- 建议在生产环境中通过环境变量覆盖敏感配置项
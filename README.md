# 彩迹 Lottery

移动优先的彩票管理、录票和推荐系统，适合个人长期记录双色球、大乐透等彩票购买与中奖情况。

项目目标不是“预测开奖”，而是把推荐、录票、开奖同步、自动判奖、历史追踪整合到一个可持续维护的系统里。

英文文档见 [README.en.md](README.en.md)。

设计文档见 [docs/system-design.md](docs/system-design.md)。

## 功能概览

- 配置驱动的多彩票类型支持，当前已实现 `福彩双色球`、`体彩大乐透`
- 按 cron 定时生成 AI 推荐，并根据开奖日历推断目标期号与开奖日期
- 按 cron 定时同步当期开奖，支持手动补历史开奖
- 上传彩票原图、OCR 识别、手动校正、保存入库
- 自动判奖、重新判奖、历史记录查询
- 推荐与购买记录关联，一条推荐可挂多条购买记录
- 移动端优先的前端界面，同时支持 Web 使用
- 单镜像 Docker 发布

## 当前支持的彩票

| code | 名称 | 第三方 ID | 红/前区 | 蓝/后区 |
| --- | --- | --- | --- | --- |
| `ssq` | 福彩双色球 | `11` | `6 (1-33)` | `1 (1-16)` |
| `dlt` | 体彩大乐透 | `13` | `5 (1-35)` | `2 (1-12)` |

所有彩票规则、开奖日历、推荐模型、推荐数量、同步 cron 都从 [backend/config/config.yaml](backend/config/config.yaml) 读取。

## 核心能力

### 1. 推荐系统

- 每种彩票可单独配置 `recommendation.cron`
- 每种彩票可单独配置推荐模型、提示词、推荐数量、历史窗口
- 根据彩种开奖日历自动推断推荐对应的期号与开奖日期
- 同一用户、同一彩种、同一期号重复生成时自动覆盖
- 支持推荐列表、详情、购买记录关联、隐蔽号码全屏展示

### 2. 票据记录

- 上传彩票原图并保留图片附件
- OCR 自动提取彩种、期号、开奖日期、号码、倍数、金额
- 识别结果可人工修改，OCR 只做辅助填写
- 支持从推荐直接录入购买记录
- 自动检测重复票据，避免重复录入
- 同一上传记录只能成功入库一次

### 3. 开奖同步与判奖

- 当期开奖同步使用 `query` 接口
- 历史开奖补录使用 `history` 接口
- 同步开奖结果后自动结算相关票据和推荐
- 支持手动重新判奖

### 4. 数据隔离

- 所有推荐、上传记录、票据、统计均按当前登录用户隔离
- 不同用户之间不会互相看到历史、推荐与购买记录

## 技术栈

- 前端：React + Vite + TypeScript
- 后端：Go + Fiber + GORM
- 数据库：SQLite
- OCR：PaddleOCR HTTP 服务
- 推荐模型：OpenAI Compatible API
- 开奖数据：极速数据 API

## 项目结构

```text
.
├─ backend/                Go Fiber 后端
│  ├─ cmd/                 服务启动入口
│  ├─ config/              YAML 配置
│  ├─ docs/                Swagger 产物
│  ├─ internal/
│  │  ├─ api/              HTTP 接口
│  │  ├─ model/            数据模型
│  │  └─ service/lottery/  彩票核心业务
│  └─ pkg/                 配置、数据库、日志等基础模块
├─ frontend/               React 前端
├─ docs/                   设计文档
├─ scripts/                Docker 构建与推送脚本
├─ Dockerfile              单镜像构建文件
├─ docker-compose.yml      部署编排
└─ .env.example            Compose 环境变量示例
```

## 本地开发

本地开发不需要 Docker，前后端直接运行即可。

### 1. 准备配置

后端采用双层 YAML 配置：

1. [backend/config/config.yaml](backend/config/config.yaml)
2. `backend/config/config.local.yaml`

先复制一份本地覆盖配置：

```powershell
Copy-Item backend/config/config.local.example.yaml backend/config/config.local.yaml
```

至少需要补这些关键配置：

- `jwt.secret`
- `jisu.appKey`
- `ai.baseURL`
- `ai.apiKey`
- `vision.baseURL`
- `vision.apiKey`

### 2. 启动后端

```powershell
cd backend
go run ./cmd
```

默认地址：

- API: [http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger: [http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 3. 启动前端

```powershell
cd frontend
pnpm install
pnpm dev
```

默认地址：

- Frontend: [http://127.0.0.1:3000](http://127.0.0.1:3000)

## 配置说明

### 业务配置

业务配置只使用 YAML：

- [backend/config/config.yaml](backend/config/config.yaml)
- `backend/config/config.local.yaml`

每种彩票都可以独立配置：

- 启用状态
- 基础号码规则
- 官方开奖日历
- 推荐生成 cron
- 推荐模型与提示词
- 推荐数量
- 开奖同步 cron
- 历史同步默认期数

### Docker 发布配置

根目录 `.env` 只服务于 `docker compose`，不参与业务配置。

示例文件见 [\.env.example](.env.example)。

可用变量：

- `LOTTERY_APP_PORT`
- `LOTTERY_CONFIG_DIR`
- `LOTTERY_DATA_DIR`
- `LOTTERY_LOG_DIR`

## Docker 部署

项目当前采用“整个项目单镜像”的部署方式。

镜像中包含：

- Go 后端可执行文件
- Swagger 文档
- 已构建的前端静态资源

前端静态资源由后端直接托管，不再依赖额外的 Caddy 或 Nginx 容器。

### 1. 准备 Compose 环境变量

```powershell
Copy-Item .env.example .env
```

### 2. 启动

```powershell
docker compose up -d --build
```

默认访问地址：

- App: [http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger: [http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 3. 宿主机持久化目录

默认挂载如下：

- `./backend/config -> /app/config`
- `./backend/data -> /app/data`
- `./backend/log -> /app/log`

因此以下内容都可以在容器外覆盖和持久化：

- `config.yaml`
- `config.local.yaml`
- SQLite 数据库文件
- 上传的彩票图片
- 服务日志

## Docker 镜像脚本

仓库提供了 PowerShell 构建与推送脚本：

- [scripts/docker-build.ps1](scripts/docker-build.ps1)
- [scripts/docker-push.ps1](scripts/docker-push.ps1)
- [scripts/build-and-push.ps1](scripts/build-and-push.ps1)

默认镜像名：

```text
ydfk/lottery
```

一键构建并推送：

```powershell
.\scripts\build-and-push.ps1
```

只构建：

```powershell
.\scripts\docker-build.ps1
```

只推送：

```powershell
.\scripts\docker-push.ps1
```

可以通过环境变量覆盖镜像名与标签：

```powershell
$env:DOCKER_IMAGE_NAME = "ydfk/lottery"
$env:DOCKER_IMAGE_TAG = "latest"
.\scripts\build-and-push.ps1
```

脚本会同时打两个标签：

- 主标签，例如 `latest`
- 当前 Git 提交短 SHA，例如 `a1b2c3d`

## 常用接口

### 认证

- `POST /api/auth/login`
- `POST /api/auth/register`
- `GET /api/auth/profile`

### 推荐

- `GET /api/lotteries/recommendations`
- `GET /api/lotteries/:code/recommendations`
- `GET /api/lotteries/:code/recommendations/:recommendationId`
- `POST /api/lotteries/:code/recommendations/generate`

### 票据

- `POST /api/lotteries/tickets/upload-image`
- `POST /api/lotteries/tickets/recognize`
- `POST /api/lotteries/tickets`
- `GET /api/lotteries/tickets/history`
- `POST /api/lotteries/tickets/:ticketId/recheck`

### 开奖同步

- `POST /api/lotteries/:code/draws/sync`
- `POST /api/lotteries/:code/draws/sync-history`
- `POST /api/lotteries/draws/sync-history`

## 测试与构建

后端：

```powershell
cd backend
go test ./...
```

前端：

```powershell
cd frontend
pnpm build
pnpm test:run
```

## 清理说明

仓库中已经移除了早期不再使用的录票拆分页面和部分遗留组件，当前代码以实际运行路径为准，避免同时维护两套旧流程。

## GitHub

目标仓库：

```text
git@github.com:ydfk/lottery.git
```

如果本地还没有设置远程仓库：

```powershell
git remote add origin git@github.com:ydfk/lottery.git
git branch -M main
git push -u origin main
```

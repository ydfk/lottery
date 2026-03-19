# 彩迹 Lottery

移动优先的彩票管理与推荐系统。

当前已经实现：

- 多彩票类型配置驱动，已支持 `福彩双色球`、`体彩大乐透`
- 历史开奖同步、当期开奖同步、自动判奖
- AI 推荐号码生成
- 拍照上传彩票、OCR 识别、手动校正、票据入库
- 推荐记录、购彩历史、花费与中奖统计
- 单镜像 Docker 发布

设计文档见 [docs/system-design.md](docs/system-design.md)。

## 技术栈

- 前端：React + Vite
- 后端：Go + Fiber + GORM
- 数据库：SQLite
- OCR：PaddleOCR HTTP 服务
- 开奖数据：极速数据 API

## 当前支持的彩票

| code | 名称 | remoteLotteryId |
| --- | --- | --- |
| `ssq` | 福彩双色球 | `11` |
| `dlt` | 体彩大乐透 | `13` |

彩票类型、推荐模型、开奖同步、生成推荐时间等均由 [backend/config/config.yaml](backend/config/config.yaml) 驱动。

## 核心功能

### 1. 推荐

- 按彩票配置的 `recommendation.cron` 定时生成推荐
- 根据开奖日历自动推断目标期号和开奖日期
- 同一彩票同一期号重复生成时自动覆盖
- 前端支持待开奖 / 已开奖分组、筛选、分页与详情查看

### 2. 票据记录

- 上传原图
- OCR 识别期号、开奖日期、号码、倍数、金额
- 手动修正识别结果后入库
- 自动关联开奖结果并判奖
- 支持重复票据校验

### 3. 开奖同步

- 定时同步当期开奖：`/caipiao/query`
- 手动批量同步历史开奖：`/caipiao/history`
- 同步后自动重算推荐和已录入票据的中奖状态

### 4. 移动端前端

- 看板：个人信息、票据总数、总花费、总中奖
- 推荐：紧凑卡片、筛选、自动加载更多
- 记录：上传 + 识别 + 校正 + 保存一体化
- 历史：筛选、排序、分页、详情、重新判奖

## 项目结构

```text
.
├─ backend/         Go Fiber 后端
├─ frontend/        React 前端
├─ docs/            设计文档
├─ Dockerfile       单镜像构建
├─ docker-compose.yml
└─ .env.example     Docker Compose 环境变量示例
```

## 本地开发

本地开发不需要 Docker。

### 1. 准备配置

后端配置使用两层 YAML：

1. [backend/config/config.yaml](backend/config/config.yaml)
2. `backend/config/config.local.yaml`

复制本地覆盖示例：

```powershell
Copy-Item backend/config/config.local.example.yaml backend/config/config.local.yaml
```

至少补齐这些配置：

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

默认前端开发地址：

- [http://127.0.0.1:3000](http://127.0.0.1:3000)

## 配置说明

### 后端业务配置

业务配置全部放在 YAML 中：

- [backend/config/config.yaml](backend/config/config.yaml)
- `backend/config/config.local.yaml`

每种彩票可单独配置：

- 基础规则
- 开奖日历
- 推荐生成 `cron`
- 推荐数量
- 推荐模型和提示词
- 开奖同步 `cron`
- 历史同步数量

### Docker Compose 环境变量

根目录 `.env` 只用于容器发布，不参与业务配置。

示例文件见 [\.env.example](.env.example)。

可配置项：

- `LOTTERY_APP_PORT`
- `LOTTERY_CONFIG_DIR`
- `LOTTERY_DATA_DIR`
- `LOTTERY_LOG_DIR`

## Docker 发布

项目提供了单镜像发布方案：

- [Dockerfile](Dockerfile)
- [docker-compose.yml](docker-compose.yml)

容器内结构：

- Go 后端负责 API、Swagger、上传文件访问
- Caddy 负责前端静态资源和反向代理
- 前端与后端都打进同一个镜像

### 1. 准备 Compose 环境变量

```powershell
Copy-Item .env.example .env
```

### 2. 启动

```powershell
docker compose up -d --build
```

默认访问地址：

- 前端：[http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger：[http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 3. 宿主机持久化目录

默认挂载：

- `./backend/config -> /app/config`
- `./backend/data -> /app/data`
- `./backend/log -> /app/log`

因此这些内容都可以在容器外覆盖：

- `config.yaml`
- `config.local.yaml`
- SQLite 数据库文件
- 上传的彩票原图
- 服务日志

## 常用接口

### 开奖同步

- `POST /api/lotteries/:code/draws/sync`
- `POST /api/lotteries/:code/draws/sync-history`
- `POST /api/lotteries/draws/sync-history`

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

## GitHub 推送

仓库目标地址：

```text
git@github.com:ydfk/lottery.git
```

如果本地还没有远程仓库，可执行：

```powershell
git remote add origin git@github.com:ydfk/lottery.git
git branch -M main
git push -u origin main
```

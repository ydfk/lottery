# lottery

## 项目结构

- `backend`: Go Fiber 后端
- `frontend`: React 前端
- `docs`: 设计文档

## 当前阶段

当前已按以下优先级推进：

1. 先完成可扩展系统设计
2. 先实现双色球
3. 优先实现拍照识别入库与自动判奖

详细设计见：

- `docs/system-design.md`

## 开奖同步

当前分两类同步：

- `query`：用于定时同步或手动同步当期开奖
- `history`：用于手动补录历史开奖

当期开奖同步：

```bash
POST /api/lotteries/:code/draws/sync
```

示例：

```json
{
  "issue": ""
}
```

说明：

- 走 `https://api.jisuapi.com/caipiao/query`
- 不传 `issue` 时默认同步当前期
- 开奖入库后会自动处理当期推荐号码和手工录入票据的中奖状态、中奖金额

## 手动同步历史开奖

后端已支持手动同步历史开奖，默认同步最近 `100` 期，并按极速数据接口的上限 `20` 条自动分页抓取。

支持的彩种：

- `ssq`：福彩双色球，`caipiaoid=11`
- `dlt`：体彩大乐透，`caipiaoid=13`

单彩种同步：

```bash
POST /api/lotteries/:code/draws/sync-history
```

示例：

```json
{
  "count": 100,
  "start": 0,
  "issue": ""
}
```

多彩种同步：

```bash
POST /api/lotteries/draws/sync-history
```

示例：

```json
{
  "lotteryCodes": ["ssq", "dlt"],
  "count": 100,
  "start": 0,
  "issue": ""
}
```

## 彩种配置驱动

当前彩票种类、AI 模型、提示词、开奖同步配置都从 [backend/config/config.yaml](F:/workSpace/lottery/backend/config/config.yaml) 读取，不再写死在代码里。

敏感信息也走 YAML。后端现在会按以下顺序加载配置：

1. `config.yaml`
2. `config.local.yaml`

也就是说，公开仓库里保留 [config.yaml](F:/workSpace/lottery/backend/config/config.yaml)，本地把敏感配置写进 `backend/config/config.local.yaml` 即可。这个文件已经加入忽略，不会提交到仓库。

每种彩票都可以单独配置：

- `remoteLotteryId`
- 推荐模型、提示词、推荐数量、历史窗口
- 定时同步 `cron`
- 开奖同步是否启用
- 每次同步历史期数

通用配置：

- `vision` 是全局通用 OCR 配置，当前默认使用 PaddleOCR，不再按彩种单独配置

说明：

- 服务启动后不会立刻同步开奖
- 只会按各彩种的 `sync.cron` 定时同步当期开奖
- 手动接口仍然可以随时调用
- 当前 `cron` 使用 6 段格式，包含秒位，例如 `0 35 21,22 * * *`

## 拍照识别

当前票据识别链路：

- 上传原图后先落盘保存
- 票据表会记录图片路径
- 接口返回 `imageUrl`，可直接查看原图
- OCR 默认调用 PaddleOCR 服务接口，不走大模型

PaddleOCR 相关配置在 [backend/config/config.yaml](F:/workSpace/lottery/backend/config/config.yaml)：

```yaml
vision:
  provider: "paddleocr"
  model: "layout-parsing"
  baseURL: "https://your-paddleocr-service/layout-parsing"
  apiKey: "your-paddleocr-token"
  timeoutSeconds: 30
  useDocOrientationClassify: false
  useDocUnwarping: false
  useChartRecognition: false
```

后端会把图片转成 Base64 后直接请求 PaddleOCR HTTP 接口；图片文件自动用 `fileType=1`，PDF 自动用 `fileType=0`。

## Docker Compose 发布

根目录已经提供单镜像发布方案：

- [Dockerfile](F:/workSpace/lottery/Dockerfile)
- [docker-compose.yml](F:/workSpace/lottery/docker-compose.yml)
- [.env.example](F:/workSpace/lottery/.env.example)

发布方式：

- 一个镜像同时包含 `backend` 和 `frontend`
- 前端构建产物会打进镜像，并由后端直接提供静态页面
- 本地开发仍然按现在的前后端开发方式运行，不依赖 Docker

### 启动

先在根目录准备 Compose 环境变量文件：

```bash
cp .env.example .env
```

Windows PowerShell 可直接执行：

```powershell
Copy-Item .env.example .env
```

`.env` 只用于 `docker compose` 的挂载目录和端口，不是应用业务配置。应用本身仍然读取：

- [backend/config/config.yaml](F:/workSpace/lottery/backend/config/config.yaml)
- [backend/config/config.local.yaml](F:/workSpace/lottery/backend/config/config.local.yaml)

在项目根目录执行：

```bash
docker compose up -d --build
```

默认端口：

- 应用：`25610`

访问地址：

- 前端：[http://127.0.0.1:25610](http://127.0.0.1:25610)
- Swagger：[http://127.0.0.1:25610/swagger/index.html](http://127.0.0.1:25610/swagger/index.html)

### 容器外覆盖配置和数据库

`docker-compose.yml` 已经把这些目录挂载到容器外：

- `${LOTTERY_CONFIG_DIR:-./backend/config}` -> `/app/config`
- `${LOTTERY_DATA_DIR:-./backend/data}` -> `/app/data`
- `${LOTTERY_LOG_DIR:-./backend/log}` -> `/app/log`

也就是说：

- 数据库文件 `db.sqlite` 在容器外持久化
- 上传的彩票图片也在容器外持久化
- `config.yaml` / `config.local.yaml` 可以直接在宿主机修改后重启容器生效

### 自定义挂载目录

如果你不想直接用仓库里的 `backend/config`、`backend/data`、`backend/log`，可以直接修改根目录 `.env`：

```bash
LOTTERY_CONFIG_DIR=/srv/lottery/config
LOTTERY_DATA_DIR=/srv/lottery/data
LOTTERY_LOG_DIR=/srv/lottery/log
LOTTERY_APP_PORT=25610
```

### 配置建议

生产环境至少准备这些文件：

- [backend/config/config.yaml](F:/workSpace/lottery/backend/config/config.yaml)
- [backend/config/config.local.yaml](F:/workSpace/lottery/backend/config/config.local.yaml)

其中敏感项建议放在 `config.local.yaml`：

- `jisu.appKey`
- `ai.baseURL`
- `ai.apiKey`
- `vision.baseURL`
- `vision.apiKey`

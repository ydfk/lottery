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
- OCR 默认调用项目内的 PaddleOCR 脚本，不走大模型

PaddleOCR 相关配置在 [backend/config/config.yaml](F:/workSpace/lottery/backend/config/config.yaml)：

```yaml
vision:
  provider: "paddleocr"
  model: "paddleocr"
  command: "python"
  args:
    - "scripts/paddleocr_runner.py"
  lang: "ch"
  useAngleCls: true
  timeoutSeconds: 30
```

运行前需要在后端环境安装 PaddleOCR。当前脚本默认通过以下命令调用：

```bash
python scripts/paddleocr_runner.py <imagePath> ch true
```

# 彩票管理和推荐系统设计

## 1. 目标与范围

### 1.1 MVP 目标

- 先支持福彩双色球
- 支持记录已购买的彩票
- 支持拍照识别双色球彩票并入库
- 支持同步开奖数据并自动判奖
- 支持 AI 推荐号码并在前端展示
- 前端移动优先，同时兼容 Web

### 1.2 扩展目标

- 支持更多彩种
- 不同彩种绑定不同推荐模型、提示词和推荐数量
- 支持多种识别方式
- 支持历史数据分析、命中率统计、推荐效果追踪

## 2. 总体架构

系统拆分为 `frontend` 和 `backend` 两个应用。

- `frontend`
  - 基于 `react-starter`
  - 负责移动端优先的展示、录入、上传、推荐查看、中奖展示
- `backend`
  - 基于 `go-fiber-starter`
  - 负责开奖同步、票据识别、推荐生成、判奖、数据存储、定时任务

核心数据流：

1. 定时任务从极速数据彩票接口同步开奖历史
2. 用户拍照上传彩票
3. 后端调用视觉识别提供者提取期号和号码
4. 后端将号码入库，并根据已同步的开奖结果自动判奖
5. 用户进入前端查看推荐、购彩记录、中奖状态和金额
6. 推荐服务基于历史开奖数据调用 AI 模型生成号码并存档

## 3. 可扩展设计原则

### 3.1 彩种插件化

每个彩种都实现统一的能力接口：

- 基础配置
- 号码解析
- 中奖判定
- 推荐参数
- OCR 提示模板
- 开奖同步映射

MVP 只实现双色球插件，后续增加彩种时复用同一套接口。

### 3.2 外部能力提供者抽象

外部能力统一做成 Provider：

- `DrawProvider`
  - 负责从外部 API 拉取开奖数据
- `RecommendationProvider`
  - 负责生成推荐号码
- `VisionRecognizer`
  - 负责从图片识别彩票内容

这样后续替换不同 AI 模型、不同 OCR 服务时，不影响业务层。

### 3.3 数据层向通用结构靠拢

虽然先只做双色球，但表结构不直接写死成“双色球专用表”。

- `lottery_types`
- `draw_results`
- `draw_prizes`
- `tickets`
- `ticket_entries`
- `recommendations`
- `recommendation_entries`

彩种差异通过 `lottery_code`、配置和规则服务承接。

## 4. 领域模型

### 4.1 彩种定义 `lottery_types`

字段建议：

- `code`: 彩种编码，如 `ssq`
- `name`: 彩种名称
- `status`: 启用状态
- `remote_lottery_id`: 外部开奖 API 对应 ID
- `red_count`
- `blue_count`
- `red_min`
- `red_max`
- `blue_min`
- `blue_max`
- `recommendation_count`
- `recommendation_provider`
- `recommendation_model`
- `vision_provider`
- `vision_model`

### 4.2 开奖结果 `draw_results`

字段建议：

- `lottery_code`
- `issue`
- `draw_date`
- `red_numbers`
- `blue_numbers`
- `sale_amount`
- `prize_pool_amount`
- `source`
- `raw_payload`

唯一键：

- `lottery_code + issue`

### 4.3 奖级明细 `draw_prizes`

字段建议：

- `draw_result_id`
- `prize_name`
- `prize_rule`
- `winner_count`
- `single_bonus`

说明：

- 用于保存每期开奖的奖级奖金
- 浮动奖和固定奖统一处理

### 4.4 购票记录 `tickets`

字段建议：

- `lottery_code`
- `issue`
- `source`
- `image_path`
- `recognized_text`
- `status`
- `cost_amount`
- `prize_amount`
- `purchased_at`
- `checked_at`
- `notes`

状态建议：

- `pending`
- `checked`
- `won`
- `not_won`
- `recognition_failed`

### 4.5 票据号码明细 `ticket_entries`

字段建议：

- `ticket_id`
- `sequence`
- `red_numbers`
- `blue_numbers`
- `is_winning`
- `prize_name`
- `prize_amount`
- `match_summary`

说明：

- 一张彩票可能包含多注号码，所以拆成明细表

### 4.6 推荐记录 `recommendations`

字段建议：

- `lottery_code`
- `issue`
- `provider`
- `model`
- `strategy`
- `prompt_version`
- `summary`
- `basis`
- `raw_payload`

### 4.7 推荐号码明细 `recommendation_entries`

字段建议：

- `recommendation_id`
- `sequence`
- `red_numbers`
- `blue_numbers`
- `confidence`
- `reason`

## 5. 关键业务流程

### 5.1 开奖同步流程

1. 定时任务扫描已启用彩种
2. 按彩种配置调用开奖接口
3. 落库存储 `draw_results` 和 `draw_prizes`
4. 对未判奖且期号已开奖的票据执行补判

MVP 使用极速数据：

- 查询单期：`/caipiao/query`
- 查询历史：`/caipiao/history`
- 彩种列表：`/caipiao/class`

### 5.2 拍照识别流程

1. 前端上传图片
2. 后端保存原图
3. 调用视觉识别提供者提取 OCR 文本和结构化号码
4. 通过双色球规则解析器生成票据明细
5. 入库保存票据和票据明细
6. 如果该期已开奖，立即判奖
7. 返回识别结果给前端确认

MVP 允许通过两种方式识别：

- 配置视觉模型后自动识别
- 未配置视觉模型时，前端可提交 OCR 文本做降级验证

### 5.3 判奖流程

1. 根据 `lottery_code + issue` 找到开奖结果
2. 由彩种规则服务计算命中情况
3. 优先使用 `draw_prizes` 中的奖金信息
4. 回写 `ticket_entries` 和 `tickets`

### 5.4 AI 推荐流程

1. 获取该彩种近期历史开奖
2. 构造推荐上下文
3. 调用指定推荐模型
4. 解析结构化结果
5. 入库存储推荐批次和推荐号码
6. 前端按彩种展示最新推荐

## 6. 双色球规则实现策略

### 6.1 基础规则

- 红球 6 个，范围 1-33
- 蓝球 1 个，范围 1-16
- 单注 2 元

### 6.2 判奖逻辑

判奖只依赖：

- 红球命中数
- 蓝球是否命中

奖级映射：

- 一等奖：6 红 + 蓝
- 二等奖：6 红
- 三等奖：5 红 + 蓝
- 四等奖：5 红 或 4 红 + 蓝
- 五等奖：4 红 或 3 红 + 蓝
- 六等奖：2 红 + 蓝、1 红 + 蓝、0 红 + 蓝

奖金来源：

- 优先读取同步到库中的开奖奖级数据
- 若某些固定奖级未返回，则按固定规则兜底

## 7. 前端信息架构

### 7.1 页面结构

- 首页仪表盘
  - 最近一期开奖
  - 今日推荐
  - 最近购票
  - 中奖汇总
- 拍照录票
  - 拍照上传
  - OCR 结果确认
  - 自动判奖结果
- 推荐中心
  - 当前彩种推荐
  - 推荐批次历史
- 票据记录
  - 按状态筛选
  - 查看号码、中奖、金额

### 7.2 交互原则

- 移动优先
- 关键操作 3 步内完成
- 号码展示视觉分区明确，红蓝球强对比
- 识别失败时允许人工修正

## 8. 后端 API 设计

MVP 建议提供以下接口：

- `GET /api/lotteries`
  - 获取彩种列表
- `GET /api/lotteries/:code/dashboard`
  - 获取首页聚合数据
- `POST /api/lotteries/:code/recommendations/generate`
  - 生成推荐
- `GET /api/lotteries/:code/recommendations/latest`
  - 获取最新推荐
- `POST /api/lotteries/:code/draws/sync`
  - 手动同步开奖
- `POST /api/lotteries/:code/tickets/scan`
  - 上传图片并识别入库
- `GET /api/lotteries/:code/tickets`
  - 获取票据列表

## 9. 推荐服务设计

### 9.1 提供者接口

- `Generate(code, issue, count, history) -> recommendation batch`

### 9.2 MVP 提供者

- `mock`
  - 无外部密钥也能跑通
  - 基于历史频次和随机扰动生成号码
- `openai-compatible`
  - 使用兼容 OpenAI Chat/Responses 的接口
  - 配置不同模型、提示词和返回数量

### 9.3 扩展方式

未来新增彩种时，可按彩种配置：

- 推荐模型
- 推荐数量
- 推荐提示模板
- 历史窗口大小

## 10. 识别服务设计

### 10.1 提供者接口

- `Recognize(image, lottery) -> structured ticket`

### 10.2 MVP 识别策略

- 优先使用视觉模型识别图片
- 识别失败时允许用户补 OCR 文本
- 使用双色球解析器做二次纠错

### 10.3 结构化输出

统一输出：

- 期号
- 票据号码列表
- 原始 OCR 文本
- 识别置信度

## 11. 定时任务设计

MVP 使用进程内定时任务：

- 启动后每隔固定分钟执行一次同步扫描
- 每次同步最近 N 期历史
- 同步成功后触发待判奖票据补判

后续可以扩展为：

- 独立 worker
- 消息队列
- 更精细的按彩种调度

## 12. 实现顺序

### 第一阶段

- 初始化目录结构
- 建立双色球领域模型
- 实现开奖同步
- 实现判奖逻辑

### 第二阶段

- 实现拍照上传和 OCR 识别接入
- 完成识别入库
- 完成移动端录票页面

### 第三阶段

- 实现推荐服务
- 完成首页和推荐中心
- 增加统计卡片和历史记录

### 第四阶段

- 支持更多彩种
- 抽离任务调度
- 增加命中率分析和推荐效果评估

## 13. 当前实现决策

为保证 MVP 可落地，本次代码实现将采用以下决策：

- 只落地双色球插件
- 后端使用 SQLite 持久化
- 使用极速数据接口同步开奖
- 推荐服务和视觉识别都做成可插拔 Provider
- 前端先做移动优先的一体化界面
- 允许 OCR 文本降级输入，保证在未接入真实视觉模型前也能完成闭环


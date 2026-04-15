# Agent Guide

本文件用于指导代码智能体在本仓库中进行一致、安全、可验证的开发。

## 仓库概览

- 后端：Go + Fiber + GORM + Swagger
- 前端：React 19 + Vite + TypeScript + Tailwind + shadcn/ui
- 包管理：前端使用 pnpm
- 运行方式：支持本地运行与 Docker

## 目录约定

- `backend/`：Go API 服务
- `frontend/`：前端应用
- `docs/`：系统设计与项目文档
- `scripts/`：项目脚本

## 常用命令

### 后端

在 `backend/` 目录执行：

- `go mod tidy`
- `go test ./...`
- `go run ./cmd/main.go`

Windows 脚本：

- `scripts\test.bat`
- `scripts\run.bat`
- `scripts\build.bat`

### 前端

在 `frontend/` 目录执行：

- `pnpm install`
- `pnpm dev`
- `pnpm build`
- `pnpm lint`
- `pnpm test`
- `pnpm test:run`

## 开发规则

- 优先最小改动，避免无关重构。
- 保持与现有代码风格一致。
- 不要修改不相关文件。
- 修改行为逻辑时，必须补充或更新测试。
- 不要硬编码密钥、Token、密码。
- 错误处理要有明确上下文，禁止静默吞错。

## 后端规则

- 变更接口时，同时更新 Swagger 注释与路由注册。
- 数据库变更时，检查迁移与模型一致性。
- 返回结构尽量复用统一响应封装。
- 新增业务逻辑优先放在 `internal/service/`，路由层保持轻量。

## 前端规则

- 使用 `@/` 路径别名导入 `src` 下模块。
- 优先函数式组件，保持组件职责单一。
- 新增 UI 优先复用 `src/components/ui/` 与现有通用组件。
- 避免引入与业务无关的大型依赖。
- 页面交互涉及表单时，优先使用 `react-hook-form` + `zod` 模式。

## 提交流程建议

1. 明确需求与影响范围。
2. 先补测试或先写失败测试。
3. 实现最小可行改动。
4. 运行后端与前端相关测试。
5. 自检安全项与回归风险。

## 自检清单

- 是否只修改了需求相关文件。
- 是否包含必要测试并通过。
- 是否有破坏性改动（API、数据结构、配置）。
- 是否更新了必要文档（README、Swagger、说明文档）。
- 是否存在敏感信息泄露风险。

## 参考文件

- `frontend/AGENTS.md`
- `README.md`
- `backend/README.md`
- `backend/README_zh.md`

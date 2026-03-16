# React Starter Scaffold (React 19 + Vite 8 + TS + Tailwind v4)

> 这是一个脚手架项目，用于快速启动中后台/组件演示型 React 应用。

---

## 中文说明

### 项目定位
- 这是一个可复用的 React 脚手架，包含路由、状态管理、表单、UI 组件与基础布局。
- 适合作为新项目起点或组件演示/验证环境。

### 技术栈
- React 19 + Vite 8 + TypeScript
- Tailwind v4 + shadcn/ui + Radix UI
- React Router + Zustand
- React Hook Form + Zod
- Vitest + Testing Library

### 快速开始
```bash
pnpm install
pnpm dev
```

### 常用命令
```bash
pnpm dev           # 本地开发
pnpm build         # 生产构建 (tsc -b + vite build)
pnpm lint          # 代码检查
pnpm format        # 代码格式化
pnpm format:check  # 检查格式
pnpm test          # 单测(监听)
pnpm test:run      # 单测(一次性)
```

#### 运行单个测试
```bash
pnpm test -- src/path/to/file.test.tsx
pnpm test -- -t "test name"
pnpm test -- shimmer-button
```

### 目录结构
```text
src/
  components/
    ui/            # shadcn/ui 组件
    magicui/       # Magic UI 组件
    layout/        # 布局与侧边栏
  pages/
    components/    # 组件演示页面
  test/            # 测试初始化
```

### 代码规范（简要）
- 使用 `@/` 作为 `src` 路径别名。
- 组件优先函数式组件，样式使用 `className` + `cn`。
- Tailwind 为主，尽量避免新增全局 CSS。
- 组件命名 PascalCase，文件名 kebab-case。

### 注意事项
- 这是脚手架项目，请保持结构简洁，避免引入与业务强绑定的内容。
- Vite 版本为 `8.0.0-beta.11`，不要随意降级。

---

## English

### Purpose
- This is a reusable React scaffold with routing, state, forms, UI components, and a basic layout.
- Ideal as a starting point for new projects or a component demo playground.

### Tech Stack
- React 19 + Vite 8 + TypeScript
- Tailwind v4 + shadcn/ui + Radix UI
- React Router + Zustand
- React Hook Form + Zod
- Vitest + Testing Library

### Quick Start
```bash
pnpm install
pnpm dev
```

### Common Commands
```bash
pnpm dev           # Start dev server
pnpm build         # Production build (tsc -b + vite build)
pnpm lint          # Lint
pnpm format        # Format
pnpm format:check  # Format check
pnpm test          # Tests (watch)
pnpm test:run      # Tests (run once)
```

#### Run a Single Test
```bash
pnpm test -- src/path/to/file.test.tsx
pnpm test -- -t "test name"
pnpm test -- shimmer-button
```

### Structure
```text
src/
  components/
    ui/            # shadcn/ui components
    magicui/       # Magic UI components
    layout/        # layout + sidebar
  pages/
    components/    # component demos
  test/            # test setup
```

### Style Notes (Short)
- Use `@/` alias for `src`.
- Prefer function components; use `className` + `cn`.
- Tailwind first; avoid extra global CSS.
- PascalCase for components, kebab-case for filenames.

### Notes
- This is a scaffold. Keep it clean and reusable; avoid app-specific coupling.
- Vite version is `8.0.0-beta.11`; avoid downgrades unless necessary.

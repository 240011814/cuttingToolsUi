# CLAUDE.md

## 项目概述

AI 驱动的切割优化与英语学习管理平台，基于 SoybeanAdmin (Vue 3) 模板构建。

- **前端**: Vue 3 + TypeScript + Naive UI + UnoCSS + Pinia (端口 9527)
- **后端**: Go + Gin + GORM + MySQL (端口 8080)
- **AI 集成**: OpenAI API + Mem0 记忆服务

## 常用命令

### 前端 (在 `frontend/` 目录下)

```bash
pnpm dev          # 启动开发服务器
pnpm build        # 生产构建
pnpm lint         # OxLint + ESLint 检查并自动修复
pnpm typecheck    # TypeScript 类型检查
```

### 后端 (在 `backend/` 目录下)

```bash
go run .          # 启动服务器
go build ./...    # 编译检查
```

## 项目结构

```
backend/
  api/            # HTTP 处理器 (按领域一个文件)
  model/          # GORM 数据模型
  service/        # 业务逻辑 + 数据库初始化
    db/migrations/ # Goose SQL 迁移文件
  config/         # 配置结构体和加载器
  main.go         # 入口、路由注册

frontend/
  src/
    service/api/  # API 调用层 (按领域一个文件)
    store/        # Pinia 状态管理 (auth, app, theme, route, tab)
    views/        # 页面组件
      ai/         # AI 功能 (聊天、词汇、笔记、训练等)
      cut/        # 切割优化功能
      system/     # 系统管理 (用户、权限、AI 配置)
    hooks/        # Vue 组合式函数
    layouts/      # 布局组件
    locales/      # 国际化 (zh-CN, en-US)
    router/       # 路由 (Elegant Router 自动生成)
  packages/       # 内部共享包 (@sa/axios, @sa/hooks 等)
```

## 代码规范

### 后端

- 三层架构: `api` (处理器) → `service` (业务逻辑) → `model` (数据模型)
- 命名: `HandleXxx` 处理器, `NewXxxService()` 构造函数
- 全局数据库: `var DB *gorm.DB` 在 `service/db.go` 初始化
- 迁移: `//go:embed` 嵌入, Goose 启动时自动执行
- 中间件: `AuthMiddleware` (JWT) → `RequirePermission("resource:action")` (RBAC)
- 响应: `SendSuccess(c, data)` / `SendError(c, code, msg)`
- 文件命名: `snake_case.go`

### 前端

- API 层: `src/service/api/` 下按领域导出类型化函数
- 状态管理: Pinia store 按模块组织
- 路由: `route.json` 文件配合 Elegant Router 自动生成
- 文件命名: Vue 组件 `kebab-case.vue`, API 文件 `kebab-case.ts`
- 格式化: 单引号, 无尾逗号, 打印宽度 120, 箭头函数无括号

## 配置文件

- `backend/config.yaml` — 数据库、mem0 API 密钥 (环境变量优先)
- `frontend/.env` — 默认环境变量
- `frontend/.env.test` / `.env.prod` — 环境覆盖
- `frontend/.oxfmtrc.json` — OxFmt 格式化规则
- `frontend/.oxlintrc.json` — OxLint 规则

## 注意事项

- Vue 模板中的 IDE 诊断误报可忽略 (TypeScript 检查正常即可)
- 后端 `config.yaml` 不要提交到 git (包含敏感信息)
- 数据库迁移文件按时间戳排序命名: `YYYYMMDDHHMMSS_description.sql`
- 前端 `packages/` 下的内部包通过 pnpm workspace 链接

## 开发注意事项

- 新增接口需要配置对应的权限，后端和前端都需要权限控制
- 数据库有变化，需要编写迁移文件
- 开发中遇到的过的问题和错误以及一些经验，写入到CLAUDE.MD 开发经验中

## 开发经验

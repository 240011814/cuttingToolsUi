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

- `backend/config.yaml` — 数据库、mem0 API 密钥、baostock API 地址 (环境变量优先)
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

MySQL 不支持在一个查询中执行多条 SQL 语句。需要将建表语句拆分成独立的 -- +goose StatementBegin 块

- baostock 代理 (baostock/): 底层 socket 无超时, 服务端偶发卡顿会永久阻塞并占住全局锁导致连锁超时, 已通过 `socket.setdefaulttimeout(60)` 兜底; 登录会话已做复用 (shared.py 包装 login/logout), 查询异常时 force_disconnect 强制重连
- baostock 周线/月线只返回已完成周期 (周线日期=周最后交易日, 月线=月末交易日), 当前周期不出 K 线; 股票代码与指数代码有重叠 (如 sz.000003 退市股 vs sh.000003 上证B指), 必须结合市场标识区分, 指数走 query_history_index_k_data_plus
- baostock 财务数据从 2007Q1 开始; 每日调用限额默认 5 万次 (BAOSTOCK_API_DAILY_LIMIT), 同步必须增量+串行, 429 限额错误不要重试
- 股票同步水位: 日K按日全量走全局水位 sync_watermark(kline_daily, 只推进到完整成功的一天, 断点续跑); stock_sync_state (每股一行) 保留周/月K与财务水位+状态, 个股日K列由按日同步顺带推进; 原则: 数据表(stock_daily/stock_finance)是真相, 状态/水位表只是缓存+观测, 漂移时最多多拉一次数据(upsert 幂等兜底); 同步成功/失败都要更新状态(失败记录原因, 便于续跑与排查); 迁移里有存量数据回填
- baostock 新版 SDK (≥00.9.30) 提供按日全量接口: query_daily_history_k_AStock(date) 1次调用返回全市场个股日K(含 preclose/tradestatus/peTTM/pbMRQ/psTTM/pcfNcfTTM/isST), 另有 query_daily_history_k_ETF / query_daily_adjust_factor; 实测返回不含指数(sh.000xxx/sz.399xxx 仍走 query_history_index_k_data_plus)和北交所, 历史深度可到 1990 年代; 代理已加 /query_daily_history_k_astock 等端点, 部署需 pip install -U baostock 并跑 baostock/verify_daily_updates.py 验证
- 日K按日全量同步流程: 交易日历(1990起, 1次调用, 含未来交易日) → 逐日拉取全市场日K → 交易日返回0行=当日行情未发布(约17:30后), 本轮停住等下轮; 首次运行从 stock_daily 存量 MAX(trade_date) 初始化水位, force 才清零重放; 周/月K仍逐股增量(不做本地聚合), 指数随逐股路径补日K
- SyncStockList 走 query_stock_basic(空参, 1次调用, 全量证券含退市): is_active 由 status 派生(停牌不再像 query_all_stock 的 trade_status 那样误标 inactive, 退市股能正确下线), list_date 取 ipoDate, type 只留 1股票/2指数(排除可转债/ETF); stock_daily 已加 preclose/trade_status/pe_ttm/pb_mrq/ps_ttm/pcf_ncf_ttm 列(按日全量与逐股路径共用 stockDailyUpsertCols 填充), is_st 由按日同步的官方逐日 isST 每日更新; Screen 的 PE/PB 仍是 close/eps 现算, 后续可切官方 pe_ttm/pb_mrq
- 定时任务统一存于 job_definitions/job_runs (TaskRegistry 注册, 方法名是 key 不做反射); 备忘= task_name=reminder.notify + user_id + once/repeat 调度, 系统任务= cron + user_id IS NULL; 重复备忘按旧链式逻辑: 每条定义=一次执行, 执行完由调度器生成下一条 (chain_id 分组, 过期不补发循环补齐), 日历只查定义行不做 run 映射; 手动触发与重试不生成下一条; 管理接口不展示用户备忘

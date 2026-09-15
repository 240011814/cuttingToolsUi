# CLAUDE.md

# 提示词优化规则

**规则说明：**  
在执行任何操作之前，第一步始终是对用户的提示词进行优化，使其更加清晰和有效。

**优化原则：**

- 保持用户原始意图不变
- 仅优化表达清晰度、结构性和可执行性
- 不改变业务逻辑、技术选型或具体需求
- 优化程度：微调为主，避免大幅重写

**执行流程：**

1. **第一步：判断是否需要优化**

   - 如果用户明确要求"不优化"、"直接执行"等，跳过优化步骤
   - 如果提示词已经非常清晰具体，可以跳过优化
   - 其他情况进入优化流程

2. **第二步：优化提示词**

   - 分析用户原始提示词的核心意图
   - 识别模糊、不清晰或可能产生歧义的部分
   - 生成优化后的提示词（更清晰、更具体、更易执行）
   - 记录优化理由和主要改动点

3. **第三步：展示 A/B/C 选项（仅展示，不执行）**

   - **A：原始提示词** - 使用用户原始提示词执行
   - **B：优化后的提示词** - 使用优化后的提示词执行
   - **C：重新优化** - 根据反馈重新优化提示词，然后再次展示选项
   - **必须完整展示提示词内容**，不要只显示一行摘要或简化版本
   - 使用代码块或引用格式完整展示，确保用户能看到完整的提示词内容
   - **必须说明优化理由**：简要说明为什么这样优化，主要改进了哪些方面

4. **第四步：等待用户选择**

   - **必须等待用户明确选择 A、B 或 C**
   - 如果用户选择 C：
     - 询问用户希望如何调整（更具体、更简洁、关注不同方面等），或根据用户的反馈意见
     - 根据用户反馈重新优化提示词
     - 返回第三步，再次展示 A/B/C 选项
     - 可以多次迭代，直到用户选择 A 或 B
   - 在用户选择之前，**不得执行任何后续操作**
   - 不得进行代码分析、文件读取、工具调用等操作
   - 只能展示选项，等待用户交互

5. **第五步：执行操作**
   - 用户选择 A 或 B 后，使用选定的提示词执行后续操作

**重要约束：**

- ❌ 禁止在展示选项后立即继续执行
- ❌ 禁止在用户未选择时进行任何分析或工具调用
- ❌ 禁止只展示一行摘要或简化版本的提示词
- ❌ 禁止改变用户原始意图和核心需求
- ❌ 禁止过度优化导致提示词面目全非
- ✅ 必须完整展示提示词内容
- ✅ 必须说明优化理由
- ✅ 只能展示选项并明确说明"请选择 A、B 或 C"
- ✅ 用户选择后再继续执行
- ✅ 如果用户选择 C，根据反馈重新优化并再次展示选项

**用户反馈处理：**

- 如果用户选择 C（重新优化），AI 应该：

  1. 询问用户希望如何调整（更具体、更简洁、关注不同方面等），或根据用户提供的具体反馈
  2. 根据用户反馈重新优化提示词
  3. 返回第三步，再次展示 A/B/C 选项供用户选择
  4. 可以多次迭代，直到用户选择 A 或 B 为止

- ✅ 只能展示选项并明确说明"请选择 A 或 B"
- ✅ 用户选择后再继续执行

## 项目概述

多模块的 AI 应用平台 (Admin 后台)，基于 SoybeanAdmin (Vue 3) 模板构建，涵盖：

- **AI 与英语学习**: AI 对话/Agent、课程、词汇、笔记、错题本、训练、历史
- **切割优化**: 切割方案与记录管理
- **彩票 / 工具箱**: 彩票数据、A 股行情与选股 (`tool/stockdetail`、`tool/stockscreen`)
- **系统管理**: 用户、角色权限 (RBAC)、AI 模型/场景配置、定时任务与提醒
- **能力集成**: OpenAI API + Mem0 记忆服务、Telegram 通知、邮件通知

技术栈:

- **前端**: Vue 3 + TypeScript + Naive UI + UnoCSS + Pinia (端口 9527)
- **后端**: Go + Gin + GORM + MySQL (端口 8080)
- **后端架构**: `api` (领域处理器) → `service` (业务逻辑) → `model` (GORM 模型) 三层; 全局 DB 在 `service/db.go`, 迁移用 Goose (`//go:embed`), 中间件 `AuthMiddleware` (JWT) + `RequirePermission` (RBAC)
- **A 股数据源**: `baostock/` Python FastAPI 代理 (端口 3002), 封装 baostock SDK, Go 后端串行调用做增量同步

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
- baostock 服务端会不定期重置长连接 ([Errno 104] Connection reset / [Errno 32] Broken pipe, SDK 打印"接收数据异常"): 会话复用下死连接上首个请求必 502, 且 SDK 是混淆闭源 (源文件 %TSD-Header 加密) 内部 socket 状态动不了, 重置后首次重新登录还可能"服务器连接失败"; 代理侧治法 = main.py `_execute_with_retry` 查询异常即 force_disconnect 并自动重试整个请求 (MAX_QUERY_ATTEMPTS=3 + 退避 0.5s/1s; 不重试: ValueError 本地参数错误 + BaostockQueryError 服务端 error_code 非0 的确定性错误; 其余瞬时错误才重连重试), 瞬时断连在代理内消化不再抛 502 打断 Go 同步; 用量计数仍按 HTTP 请求计, 重试多耗的调用靠 45k<50k 余量吸收
- baostock 代理曾用 baostock_lock + _session_lock 两把锁: force_disconnect 在持 _session_lock 时做真实网络 logout, 而等 login 的线程在持 baostock_lock 时等 _session_lock, logout 卡住即锁序倒置导致全服务无响应(同步 def handler 共享 anyio 40 线程池, 池耗尽后 /health 也挂); 已合并为单一可重入锁 baostock_lock(login 复用同一把锁), 消除死锁; 遗留: SDK 调用本身卡死仍会占住该锁, 靠健康检查看门狗自愈
- DLP 加密陷阱: 本机装有文件透明加密(天锐/TSD), 凡 `python.exe`(含 .venv)写入仓库内文件会被加密成 `%TSD-Header-###%`, git/diff 视为二进制、Docker COPY 也会带密文; git.exe/powershell/编辑器写入则保持明文。批量改 .py 后务必用 `git diff --numstat`(出现 `- -` 即中招)或看文件头校验, 中招后 `git checkout -- <file>` 还原再用受信工具重写; 读源码时 baostock SDK 本身的 %TSD-Header 加密属正常
- baostock 周线/月线只返回已完成周期 (周线日期=周最后交易日, 月线=月末交易日), 当前周期不出 K 线; 股票代码与指数代码有重叠 (如 sz.000003 退市股 vs sh.000003 上证B指), 必须结合市场标识区分, 指数走 query_history_index_k_data_plus
- baostock 财务数据从 2007Q1 开始; 每日调用限额默认 5 万次 (BAOSTOCK_API_DAILY_LIMIT), 同步必须增量+串行, 429 限额错误不要重试
- 股票同步水位: 日K按日全量走全局水位 sync_watermark(kline_daily, 只推进到完整成功的一天, 断点续跑); stock_sync_state (每股一行) 保留周/月K与财务水位+状态, 个股日K列由按日同步顺带推进; 原则: 数据表(stock_daily/stock_finance)是真相, 状态/水位表只是缓存+观测, 漂移时最多多拉一次数据(upsert 幂等兜底); 同步成功/失败都要更新状态(失败记录原因, 便于续跑与排查); 迁移里有存量数据回填
- baostock 新版 SDK (≥00.9.30) 提供按日全量接口: query_daily_history_k_AStock(date) 1次调用返回全市场个股日K(含 preclose/tradestatus/peTTM/pbMRQ/psTTM/pcfNcfTTM/isST), 另有 query_daily_history_k_ETF / query_daily_adjust_factor; 实测返回不含指数(sh.000xxx/sz.399xxx 仍走 query_history_index_k_data_plus)和北交所, 历史深度可到 1990 年代; 代理已加 /query_daily_history_k_astock 等端点, 部署需 pip install -U baostock 并跑 baostock/verify_daily_updates.py 验证
- 新版 SDK 移除了 bs.query_history_index_k_data_plus, 但 query_history_k_data_plus 直接支持指数代码(代理 /query_history_index_k_data_plus 内部本就走它, 无需改); 服务端按"证券类型×频率"严格校验 fields: preclose/tradestatus/peTTM 等仅日线有效, 周/月线传了报 error_code 10004012 无效参数(代理 502), Go 侧已按 freq×isIndex 选字段; 直连 baostock 测试时 socket 无超时会被服务端卡死, 务必参考代理 shared.py 的 setdefaulttimeout 兜底
- 日K按日全量同步流程: 交易日历(1990起, 1次调用, 含未来交易日) → 逐日拉取全市场日K → 交易日返回0行=当日行情未发布(约17:30后), 本轮停住等下轮; 首次运行从 stock_daily 存量 MAX(trade_date) 初始化水位, force 才清零重放; 周/月K仍逐股增量(不做本地聚合), 指数随逐股路径补日K
- SyncStockList 走 query_stock_basic(空参, 1次调用, 全量证券含退市): is_active 由 status 派生(停牌不再像 query_all_stock 的 trade_status 那样误标 inactive, 退市股能正确下线), list_date 取 ipoDate, type 只留 1股票/2指数(排除可转债/ETF); stock_daily 已加 preclose/trade_status/pe_ttm/pb_mrq/ps_ttm/pcf_ncf_ttm 列(按日全量与逐股路径共用 stockDailyUpsertCols 填充), is_st 由按日同步的官方逐日 isST 每日更新; Screen 的 PE/PB 仍是 close/eps 现算, 后续可切官方 pe_ttm/pb_mrq
- 定时任务统一存于 job_definitions/job_runs (TaskRegistry 注册, 方法名是 key 不做反射); 备忘= task_name=reminder.notify + user_id + once/repeat 调度, 系统任务= cron + user_id IS NULL; 重复备忘按旧链式逻辑: 每条定义=一次执行, 执行完由调度器生成下一条 (chain_id 分组, 过期不补发循环补齐), 日历只查定义行不做 run 映射; 手动触发与重试不生成下一条; 管理接口不展示用户备忘
- 财务同步增量按 finance_sources 标记判断 (P盈利G成长O营运C现金流B偿债), 不是看行是否存在; 标记以数据列补齐(只增不减)——旧数据/被覆盖的标记(如利润列有值但标记缺P)按实际数据推断, 避免重复拉取, 标记齐全但列全空的来源不重拉(baostock 本身无值); 写标记必须与库中原有标记合并——只写本次来源会把旧行覆盖成 "G"/"PB" 乒乓, 每轮重复拉接口; 旧部署(9/12~9/14)曾把 O/C 数据跨股票/季度写串(600071/600072 互串), 已迁移清空 O/C 六列并剔除标记强制重拉; 所有百分比列统一 ×100 口径: roe/gross_margin/net_margin/debt_ratio 已从 baostock 原始小数改为百分比存储(迁移 ×100, 带 ABS<5 防重守卫), 与成长/营收同比列一致, 前端按百分比直接渲染; 周转率/流动比率/cfo_to_or 等倍数类列保持原始值不加 100; stock_finance 比率/同比列 decimal(8,4) 放宽为 (12,4) (小基数下同比可达上万, 1264 out of range), 比率列统一 (12,4) 治本
- GORM 零值陷阱: bool 字段带 `gorm:"default:true"` 时, Create/upsert 写入 false 会被零值替换成 true (INSERT 与 ON DUPLICATE KEY UPDATE 双双失效), 曾导致全部 7674 只证券 is_active=1、退市股/停牌股全进列表页; 修复=去掉 default 标签(临时表实测验证), 需要 DB 默认值与零值语义不一致的字段要么用指针要么不设 default 标签
- 证券代码全链路存完整格式 sz.000003/sh.600000/bj.430047 (9/14 定稿): 裸 6 位码曾使 sz.000003(退市股)与 sh.000003(上证B指)在 stock_info 唯一约束下互相覆盖、身份随同步翻转、stock_daily 混写双方K线; 对比过 code+type 复合键/指数分表/混合键等方案, 最终选完整码根治——唯一键不变无需复合, isIndexCode(code)=sh.000*/sz.399* 前缀判断(不再需要 market 参数), convertToBaostockCode 原样返回, resolveMarket 按前缀; 迁移(20260914163000): 000% 歧义K线/状态无法按行甄别直接删除(sh.000xxx 与 sz.000xxx 历史混写), 其余按 market(stock_info)/首数字(6=沪 0/3=深 4/8/9=北)回填前缀, 五表 code 加宽 varchar(16); type 列(1股票/2指数)保留用于 Screen 的 securityTypes 过滤; 前端展示剥前缀(列表代码列/详情页/东财实时行情链接); 部署后必须: ①同步股票列表 ②行情同步 force=1 跑一次(重建被删的 000% 日K历史, 约8800次调用), 周/月K与指数K线因状态已删自动全量重拉
- 小时线(60min)并入逐股K线同步(9/15): stock_daily.trade_date 由 date 放宽为 datetime(迁移 20260915000001; date→datetime 属重建表的类型转换, 大表需低峰/停机), 唯一键 uk_code_freq_date 不变——日/周/月线存交易日 00:00:00, 小时线存 bar 时间(10:30/11:30/14:00/15:00); **大坑**: DSN 为 parseTime=True&loc=Local, go-sql-driver 写 time.Time 前会先 .In(Local), 若日K仍用 time.Parse(得到 UTC 零点)解析, 新写日线会落成 08:00 而迁移后的老行是 00:00, 唯一键视作不同行→同一交易日重复写入且 upsert 失效; 治法=K线日期统一 time.ParseInLocation(..., time.Local)(parseTradeDay), 与迁移后 00:00 对齐; 小时线目标日 LatestCompletedTradeDay: 收盘(15:00, marketCloseHour)后取最新交易日, 盘中回退上一交易日(end_date 同步钳到该日, 避免把当日未完成 bar 写入), freqIsCurrent 要求已存到 15:00 收盘 bar(可自愈收盘后延迟发布最后一根); 小时线 fields 仅 date,time,open,high,low,close,volume,amount(无 preclose/tradestatus/估值/换手, 传了报 10004012 无效参数), time 字段格式 YYYYMMDDHHmmssSSS; stock_sync_state 增 kline_hourly_to(datetime, 迁移 20260915000002); 指数小时线暂不同步(baostock 指数分钟线支持待实测), 首次小时线全量约 5400 次调用, 勿与日K force 重放排同一天
- MySQL->ClickHouse 行情/财务复制(9/15): CH 只做下游只读副本(回测/分析), MySQL 仍是唯一真相源, 现有读写链路一行不动; 增量按 updated_at 时刻水位(sync_watermark 新增 last_time 列; stock_daily 为此新增 updated_at DEFAULT CURRENT_TIMESTAMP ON UPDATE, **GORM 模型刻意不加该字段**——加了会进入现有 upsert 的 INSERT/SET, 风险大且无必要, CH 侧用局部结构体读); CH 表用 ReplacingMergeTree(updated_at) 以 updated_at 为版本列, ORDER BY (code,frequency,trade_date) / (code,report_date), 重复/重放写入天然幂等(读侧要带 FINAL 或 argMax 才即时去重); CH 表结构在 service/db/ch_migrations/*.sql(与 MySQL 的 Goose 迁移相互独立, 启动时按 CH 内 schema_migrations 表应用, 表结构变更请新增文件); 首次/修复用定时任务 stock.sync_clickhouse 的参数 {"full": true} (TRUNCATE+全量), 日常走同任务默认增量(无水位时等价全量); 无 HTTP 接口, 水位名 clickhouse_stock_daily / clickhouse_stock_finance; 配置 config.clickhouse(enabled 默认 false, 支持 CH_ENABLED/CH_HOST/CH_PORT/CH_USER/CH_PASSWORD/CH_DATABASE 环境变量), 未启用或连不上只告警不阻断启动; 日期列用 DateTime('Asia/Shanghai') 与 MySQL 的 loc=Local 对齐(小时内 bar 时间不被时区平移); 坑: 水位为空时不要写零值时间(datetime 下限 1000-01-01 会报错, 已跳过); 限制: 行删除/TRUNCATE MySQL 不会自动反映到 CH(需 full 重建), 同秒并发更新可能靠 ReplacingMergeTree 收敛

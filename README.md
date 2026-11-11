# Utopia Daily

WIP

这是一个仍在开发中的项目，旨在构建模块化的新闻抓取、分析与分发平台，支持可拓展插件加载、定时任务调度、可视化管理后台以及简洁的报纸阅读前端。

灵感来源：[TrendRadar](https://github.com/sansan0/TrendRadar)

旧版 Python 原型源码已完整归档保留在 `old/` 目录。当前版本已全面采用 Go 与 Vue 3 进行重写重构，整体架构仍处于持续迭代阶段，部分功能尚未完全定型，分类模型模块当前仅保留了可扩展接口槽，代码主要面向个人研究与测试，暂不建议直接用于生产环境或发起大规模 PR。

本项目核心目标依然是构建结构清晰、模块解耦的可扩展系统，而非单一脚本工具。基于 Go 单一静态二进制与轻量化运行时设计，系统可在低配置 Linux 服务器或旧设备上稳定运行，同时通过 `uv` 隔离外部 Python 插件依赖与资源占用。

## 免责声明

本项目为通用数据处理与信息聚合框架，旨在提升个人资讯索引与阅读效率。严禁利用本框架开发任何违反当地法律法规、侵犯他人版权或非法获取计算机信息系统数据的插件或模块。

本仓库不对任何基于本框架开发的第三方插件的合法性负责。用户在使用过程中产生的任何法律风险与后果，概与本仓库贡献者无关。请仅用于合规的技术研究与个人学习。

## 系统架构

- **后端核心**：基于 Go 与 Gin 构建，采用 `controller`、`service`、`repository` 与 `model` 分层解耦设计，数据库遵循第三范式设计并支持 SQLite、MySQL 与 PostgreSQL，内置定时调度器、事件总线与滚动日志。
- **认证与权限**：采用 Altcha SHA-256 工作量证明校验与 Argon2id 密码哈希，结合短周期 JWT 访问令牌、数据库持久化轮转刷新会话与失败锁定保护，支持 SMTP 邮箱激活与密码重置，并提供细粒度 RBAC 权限控制。
- **国际化**：后端仅负责组装基础字典、内置模块语言包与外部插件 `locales/*.json` 缓存并下发字典，界面文案、错误提示与事件模板均由前端独立渲染。
- **内置模块**：
  - 抓取模块：`telegram_channel` 用于公开 Telegram 频道抓取，`bilibili_ranking_and_update` 用于全站热榜与关注 UP 主投稿抓取。
  - 推送模块：`smtp_pusher` 支持系统级定时摘要投递与用户自定义个人订阅推送。
- **外部插件机制**：通过基于 Stdio 的通用 JSON 管道协议调用跨语言插件，原生适配 Python 与 Node.js 等可执行程序，并通过容器内预置的 `uv` 为 Python 插件按目录自动创建独立 `.venv` 虚拟环境。
- **分类模型槽**：在 `internal/classification` 与数据库中已预留分类器接口、标签分组与文章标签关联表，等待后续模型方案重写接入。
- **双端界面**：使用两个独立的 Vue 3 + Vite 项目构建，通过 `go:embed` 直接内嵌进后端二进制文件，根路径 `/` 提供报纸阅读视图，`/dashboard` 提供控制台与初始化向导。

## 外部插件

外部抓取插件放置于 `plugins/scrapers/<plugin_id>/`，外部推送插件放置于 `plugins/push/<plugin_id>/`。目录内需包含 `manifest.json` 声明模块元数据、配置项与任务调度规则，主程序通过标准输入输出与插件子进程交换单次任务 JSON 数据。详细协议说明请参阅 [plugins/README.md](plugins/README.md)，目录下已附带示例插件 `example_hn_rss`。

## 开发阶段调试方法

### 直接调试

需要 Go 1.25+、Node.js 22+ 以及 Python 3，建议使用 `uv` 管理外部插件依赖。

```bash
# 构建控制面板与报纸前端
cd web/dashboard && npm ci && npm run build && cd ../..
cd web/newspaper && npm ci && npm run build && cd ../..

# 编译并启动服务
CGO_ENABLED=0 go build -o utopia_daily ./cmd/utopia_daily
./utopia_daily
```

启动后访问 `http://localhost:8000` 查看报纸首页，访问 `http://localhost:8000/dashboard` 进入初始化向导与管理控制台。

### Docker 部署调试

项目提供了多阶段构建的 `Dockerfile` 与 `docker-compose.example.yml`。容器运行时以 `/usr/bin/tini` 作为入口以回收跨语言插件子进程，内部已预装 `python3`、`nodejs` 与 `uv`。

```bash
docker build -t utopia-daily:latest .
docker run -d \
  --name utopia-daily \
  -p 8000:8000 \
  -v ${PWD}/data:/app/data \
  -v ${PWD}/plugins:/app/plugins \
  -v ${PWD}/logs:/app/logs \
  utopia-daily:latest
```

### 环境变量覆盖

除编辑 `data/config/config.yaml` 外，还可通过环境变量覆盖运行参数：

- 服务监听：`UDAILY_SERVER_HOST`、`UDAILY_SERVER_PORT`、`UDAILY_SERVER_MODE`
- 公开基址：`UDAILY_BASE_URL`
- 数据库配置：`UDAILY_DB_DRIVER`、`UDAILY_DB_DSN`
- 安全密钥：`UDAILY_JWT_SECRET`
- 存储与日志：`UDAILY_PLUGINS_DIR`、`UDAILY_CACHE_DIR`、`UDAILY_LOG_LEVEL`

## 开源协议

本项目基于 [LICENSE](LICENSE) 中的条款开源。

---
name: cobra-cli-design
description: Cobra CLI 命令行传入参数设计原则
---

# cobra_cli_design

用于设计、评审或重构基于 Cobra 的 Go 命令行程序。重点保证命令语义清晰、参数层级稳定、配置来源可预期、错误输出友好，并能与当前项目的 `cmd/root.go` 入口风格保持一致。

## 核心原则

- 命令表达“要做什么”，参数和 flag 表达“怎么做”。
- CLI 契约要稳定；命令名、flag 名、默认值和退出码都属于对用户的长期承诺。
- 默认行为应安全、可预测；危险操作必须显式确认或使用明确 flag。
- 优先使用 `RunE` 返回错误，避免在业务逻辑中大量 `log.Fatal` / `os.Exit`。
- 配置文件、环境变量、命令行 flag 的优先级必须明确，通常是：flag > env > config file > default。
- 帮助信息、示例、错误提示要能让用户无需读源码就完成常见操作。

## 命令层级

- 根命令只承载应用名称、全局说明、全局 flag 和默认启动行为。
- 子命令按业务能力划分，例如：`server`、`config`、`user`、`db`、`version`。
- 命令名使用小写短横线风格，避免下划线和驼峰：`init-config`、`migrate-db`。
- 对长期服务类程序，推荐保留清晰启动命令，例如 `lnmts_server serve`；如果根命令默认启动服务，也要在帮助中明确说明。
- 不要把多个不相关动作塞进一个命令并依赖复杂 flag 分支；应拆成子命令。
- 命令层级不宜过深，通常 2 到 3 层足够。

## Flag 设计

- 常用 flag 提供短别名，例如当前项目的 `--config` / `-c`。
- flag 名使用小写短横线：`--config`、`--http-addr`、`--log-level`。
- 布尔 flag 使用正向语义，避免双重否定：优先 `--verbose`，避免 `--no-disable-log`。
- 必填 flag 要显式标记并校验；可选 flag 必须有合理默认值和帮助说明。
- 全局 flag 使用 `PersistentFlags()`，仅当前命令使用的 flag 使用 `Flags()`。
- 不要为同一个含义设计多个 flag；兼容旧名称时要标记 deprecated。

## 参数与校验

- 位置参数用于核心对象，例如 `user_name`、`file`；可选配置优先用 flag。
- 使用 Cobra 的 `Args` 校验参数数量和格式，例如 `cobra.ExactArgs(1)`。
- 参数校验失败返回用户可理解的错误，不输出堆栈和内部实现细节。
- 路径参数要处理相对路径、绝对路径、文件不存在、权限不足等情况。
- 端口、地址、超时时间、日志级别等参数必须校验范围和合法值。
- 对密码、token、数据库 DSN 等敏感参数，不建议通过命令行明文传入。

## 配置加载

- 明确配置来源优先级：命令行 flag 覆盖环境变量，环境变量覆盖配置文件，配置文件覆盖默认值。
- 当前项目通过 `--config, -c` 指定 YAML 配置文件，默认值为 `./config.yml`。
- 配置读取失败要说明具体路径和原因，例如文件不存在、YAML 格式错误、权限不足。
- 配置加载后不要默认打印完整配置，避免泄露数据库密码、token、路径等敏感信息。
- 建议提供 `config init` 或 `config print --safe` 之类命令生成/查看脱敏配置。
- 默认配置应能本地最小化启动，例如 SQLite、本机监听地址、开发级日志。

## 输出与日志

- 普通结果输出到 stdout，错误和诊断信息输出到 stderr。
- 面向人的输出简洁清晰；面向脚本的输出可提供 `--json`。
- 启动 banner / ASCII art 只适合交互式启动，脚本模式或 `--quiet` 下应可关闭。
- 日志级别通过配置或 flag 控制，常见级别：`debug`、`info`、`warn`、`error`。
- 不在日志中打印密码、完整 token、数据库 DSN、完整请求体等敏感内容。
- 长时间运行服务要记录启动地址、配置来源、关闭信号和退出结果。

## 错误处理与退出码

- 成功退出码为 `0`，用户输入错误通常为 `2`，运行时错误通常为 `1`。
- 使用 `RunE` 将错误交给 Cobra 统一处理；必要时设置 `SilenceUsage` 避免运行时错误重复打印帮助。
- 参数错误可显示简短用法提示；业务错误只显示原因和建议操作。
- 不在库代码和深层业务逻辑中直接 `os.Exit`，由 CLI 边界统一决定退出。
- `log.Fatalf` 会直接退出，适合极少数无法恢复的入口错误；更推荐返回 error。
- 对服务启动失败、配置错误、端口占用等常见问题，要输出可操作的修复建议。

## 长运行服务命令

- 服务启动命令应支持优雅退出，监听 `SIGINT` / `SIGTERM`。
- 收到退出信号后关闭 HTTP 服务、数据库连接、后台 goroutine 和定时器。
- 启动成功后输出监听地址和配置文件路径，但不要输出敏感配置值。
- 支持 readiness / health 信息时，可在帮助中说明健康检查接口。
- 避免固定 `time.Sleep` 作为关闭保障；优先使用 context、WaitGroup 或服务端 Shutdown。
- 多端口或后台任务启动失败时要完整回滚已启动组件。

## 安全与可维护性

- CLI 参数、环境变量、配置文件都可能进入日志或 shell history，敏感信息要谨慎设计。
- 默认监听地址要保守；开发默认可用 `127.0.0.1`，对外暴露必须显式配置。
- destructive 命令如清库、重置用户、删除 token，需要确认参数或 `--force`。
- root command 初始化只做轻量工作，避免 import 时产生副作用。
- 业务逻辑放到独立包中，Cobra command 只负责解析参数、组装配置、调用服务。
- 命令注册、配置加载、业务执行要可测试，避免全局状态过多。

## Cobra 实现建议

- 根命令建议设置 `Use`、`Short`、`Long`、`Example`、`SilenceUsage` 和 `SilenceErrors`。
- 优先写 `RunE: func(cmd *cobra.Command, args []string) error { ... }`。
- 使用 `cmd.Context()` 传递取消信号和超时控制。
- flag 变量集中定义，避免散落在多个文件造成初始化顺序不清晰。
- 对不同子命令拆分文件，例如 `serve.go`、`config.go`、`user.go`。
- 修改 flag 或命令后同步更新 README、使用示例和发布说明。

## 当前项目关注点

- 当前根命令 `lnmts_server` 直接启动服务，并提供 `--config, -c` 指定配置文件。
- `config.NewConfigPath(configFilePath)` 是配置加载入口；后续新增 flag 时要明确是否覆盖配置文件字段。
- 当前启动时会打印完整 `config`，后续如包含数据库密码等敏感字段，应改为脱敏输出。
- 服务已有 `SIGINT` / `SIGTERM` 监听，后续应继续完善数据库和 HTTP 服务的优雅关闭。
- 如果新增 `version`、`config init`、`serve` 等子命令，应保持根命令帮助简洁，避免破坏现有启动体验。
- 当前项目使用 YAML 配置，CLI 设计应避免与配置文件字段重复且无优先级说明。

## 评审清单

- [ ] 命令名是否表达动作，flag 是否表达选项？
- [ ] 根命令、子命令、默认行为是否清晰且向后兼容？
- [ ] flag 名、短别名、默认值、必填规则是否一致？
- [ ] 配置文件、环境变量、flag 的优先级是否明确？
- [ ] 参数错误、配置错误、运行错误是否有合适退出码？
- [ ] 输出是否区分 stdout / stderr，是否支持脚本友好模式？
- [ ] 是否避免在命令行、日志、帮助示例中泄露敏感信息？
- [ ] README、示例、帮助信息是否与实际命令同步？

## 输出模板

当用户要求设计或评审 Cobra CLI 时，优先按以下格式输出：

```text
Cobra CLI 设计建议：

- 命令结构：<root / subcommands>
- Flags：<全局 flag 与命令级 flag>
- 配置优先级：<flag > env > config > default>
- 校验规则：<参数数量、格式、范围>
- 输出与退出码：<stdout/stderr、错误码>
- 文档变更：<README / --help / 示例>
```

## 示例

服务启动命令：

```text
lnmts_server serve --config ./config.yml

全局 flag:
- -c, --config string: 配置文件路径，默认 ./config.yml
- --log-level string: 日志级别，可选 debug/info/warn/error

行为：
- 读取配置文件并应用命令行覆盖项
- 校验 HTTP 地址、数据库配置和日志配置
- 启动服务并监听 SIGINT / SIGTERM 优雅退出
- 成功退出返回 0，配置错误返回 2，运行时错误返回 1
```

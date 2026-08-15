---
name: yaml-config-design
description: YAML配置设计原则
---

# yaml_config_design

用于设计、评审或重构 YAML 配置文件、默认配置、配置加载逻辑和示例文档。重点保证配置层级清晰、默认值安全、字段可校验、敏感信息不泄露，并能与当前 Go `yaml.v3` / Cobra CLI 项目风格保持一致。

## 核心原则

- 配置表达“可部署差异”，不要把业务常量、临时开关或运行时状态都塞进 YAML。
- 配置结构要稳定、可读、可校验；字段名、层级和默认值都属于用户契约。
- 默认配置必须安全可运行，避免默认暴露公网、弱密码、生产数据库或危险开关。
- 配置文件、环境变量、命令行 flag 的优先级必须明确，通常是：flag > env > config file > default。
- 示例配置必须与代码结构体同步；未实现字段不要出现在正式示例里，除非明确标注“预留/未生效”。
- 敏感配置最小化落盘，必要时使用环境变量、密钥管理或外部 secret 文件。

## 层级与命名

- 顶层按领域分组，例如当前项目的 `log`、`server`、`database`。
- 字段名使用小写 `snake_case`，与 Go 结构体 `yaml` tag 保持一致：`broadcast_port`、`http_addr`。
- 同一层级只放同一职责的字段，避免把数据库字段放进 `server` 或把日志字段放进 `database`。
- 嵌套层级不宜过深；通常 2 到 3 层足够。
- 驱动型配置使用 `driver` + 子配置块，例如 `database.driver: sqlite` 搭配 `database.sqlite.file`。
- 已发布字段避免随意重命名；如必须重命名，要提供兼容期和迁移说明。

## 默认值

- 默认值应集中定义在代码中，例如当前项目的 `config.NewConfig()`。
- YAML 文件只覆盖需要变化的字段；缺失字段应回落到安全默认值。
- 本地开发默认建议使用 SQLite、本机监听地址和 `info` 日志级别。
- 对外监听地址、MySQL 连接、调试日志、危险操作等必须显式配置。
- 默认路径使用相对路径时，要说明相对的是工作目录还是配置文件所在目录。
- 不要在默认配置中放真实密码、真实 token、真实生产地址或个人隐私数据。

## 类型与校验

- 每个配置字段都要有明确类型、单位、合法范围和默认值。
- 端口使用整数并校验范围 `1..65535`；地址字段校验 `host:port` 或明确格式。
- 日志级别使用枚举：`debug`、`info`、`warn`、`error`。
- 数据库驱动使用枚举：当前项目为 `sqlite` / `mysql`。
- 文件路径字段要检查空值、非法字符、父目录是否存在或是否可创建。
- YAML 解析成功不代表配置有效；解析后必须做业务校验。

## 敏感信息

- 密码、token、密钥、证书路径、数据库 DSN 都按敏感配置处理。
- 示例文件中使用占位值，例如 `<password>`、`${LNMTS_DB_PASSWORD}`，不要写真实凭据。
- 日志和启动输出不要打印完整配置；输出前要脱敏 `password`、`token`、`secret` 等字段。
- 优先通过环境变量注入敏感值，YAML 中只保留变量引用或说明。
- 文件权限要保守，生产配置不应被普通用户或日志采集系统随意读取。
- 如果配置支持导出/打印，提供 `--safe` 或默认脱敏模式。

## 环境变量与 CLI 覆盖

- 明确环境变量命名规则，例如 `LNMTS_SERVER_HTTP_ADDR`、`LNMTS_DATABASE_DRIVER`。
- 环境变量适合覆盖部署差异和敏感值；复杂结构仍由 YAML 表达。
- Cobra flag 适合覆盖启动时最常用配置，例如当前项目的 `--config, -c`。
- 新增 CLI flag 覆盖 YAML 字段时，要写清楚优先级和对应配置路径。
- 不要让同一配置项存在多个互相冲突的名称或来源。
- 加载完成后建议形成一个最终 `Config` 对象，再传给业务模块。

## 版本与兼容

- 对长期维护项目，可在顶层加入 `version` 或 `config_version`。
- 新增字段应保持向后兼容，并给出默认值。
- 删除或改名字段需要标记 deprecated，保留兼容解析期。
- 对字段语义变化要写迁移说明，例如端口单位、路径相对基准、日志级别含义。
- 解析未知字段时建议在开发/CI 中报错或告警，避免用户以为配置已生效。
- 示例配置、README、默认配置生成器必须与版本同步。

## 注释与文档

- YAML 示例要包含必要注释：用途、单位、可选值、默认值、是否必填。
- 注释要解释业务含义，不重复字段名本身。
- 对敏感字段注释必须提示推荐来源和安全注意事项。
- 示例配置应尽量可直接复制运行，但生产必改项要明确标出。
- README 中应包含最小配置、完整配置、常见部署配置和故障排查。
- 如果项目提供 `docs/default.yml`，它必须和 `config.Config` 结构一致。

## Go `yaml.v3` 实现建议

- Go struct 使用显式 `yaml` tag，字段名与 YAML 保持一致。
- 配置加载流程建议为：创建默认配置 → 读取 YAML → 解析覆盖 → 环境变量覆盖 → CLI flag 覆盖 → 校验。
- 使用 `yaml.Decoder` 并启用严格解析能力时，可更早发现未知字段；否则要额外检测或在文档中说明。
- 解析函数只负责加载与校验，不直接启动服务或产生全局副作用。
- 为配置提供 `Validate() error`，集中校验 driver、端口、地址、路径和必填字段。
- 为敏感配置提供脱敏输出方法，例如 `SafeString()` 或 `Redacted()`。

## 当前项目关注点

- 当前配置入口是 `config.NewConfigPath(file_path)`，先调用 `NewConfig()` 再用 YAML 覆盖默认值。
- 当前配置顶层包含 `log`、`server`、`database`，新增字段应同步 `config.Config` 结构体和示例文件。
- 当前 `docs/default.yml` 中存在 `log.max_size`、`log.max_backups`、`log.max_age`，但 `config.Log` 目前只定义 `level`、`dir`；如要生效需补结构体字段，否则应从示例移除或标注未实现。
- 当前 `server.http_addr` 在代码默认值中存在，但示例 `docs/default.yml` 尚未展示完整，应补全示例或说明默认值。
- 当前数据库支持 `sqlite` 与 `mysql`，示例应展示 `driver` 与对应子配置块。
- 当前 CLI 使用 `--config, -c` 指定配置文件，配置设计应和 CLI 优先级说明保持一致。

## 评审清单

- [ ] 顶层分组是否清晰，字段名是否使用 `snake_case`？
- [ ] YAML 示例是否与 Go struct 和默认值完全同步？
- [ ] 每个字段是否有类型、默认值、单位、合法范围说明？
- [ ] 是否明确 flag、环境变量、配置文件、默认值的优先级？
- [ ] 敏感字段是否避免明文示例、启动打印和日志泄露？
- [ ] 配置加载后是否有集中校验和友好错误提示？
- [ ] 未知字段、废弃字段、版本迁移是否有处理策略？
- [ ] README、默认配置、CLI 帮助是否同步更新？

## 输出模板

当用户要求设计或评审 YAML 配置时，优先按以下格式输出：

```text
YAML 配置设计建议：

- 配置结构：<顶层分组与关键字段>
- 默认值策略：<代码默认值 / 示例覆盖 / 必填项>
- 覆盖优先级：<flag > env > yaml > default>
- 校验规则：<枚举、端口、路径、必填、范围>
- 安全策略：<敏感字段、脱敏、环境变量>
- 同步变更：<Go struct / docs/default.yml / README / CLI help>
```

## 示例

当前项目推荐的最小配置示例：

```yaml
# 日志配置
log:
  dir: "./log"          # 日志目录
  level: "info"        # debug, info, warn, error

# 服务配置
server:
  broadcast_port: 5000  # UDP 内网广播端口，范围 1..65535
  http_addr: "127.0.0.1:28080" # HTTP 监听地址

# 数据库配置
database:
  driver: "sqlite"     # sqlite 或 mysql
  sqlite:
    file: "./data.db"  # SQLite 数据文件路径
  mysql:
    host: "127.0.0.1"
    port: 3306
    user: "lnmts"
    password: "${LNMTS_DB_PASSWORD}"
    database: "lnmts"
```

关键建议：示例字段必须与 `config.Config` 同步；启动时只打印配置路径和脱敏摘要，不打印完整数据库密码。

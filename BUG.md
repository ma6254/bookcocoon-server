# BookCocoon Server — Bug 清单

> 本清单基于对项目源码（Go 后端，排除 `vendor/` 第三方依赖）的逐文件静态审查。
> 审查范围：`main.go`、`cmd/`、`config/`、`database/`、`server/`、`utils/`、`validator/`、`docs/`、构建脚本与配置。
>
> 严重度分级：🔴 严重（安全/阻断） · 🟠 高（核心逻辑） · 🟡 中（功能缺陷） · 🟢 低（代码质量/边缘情况）
>
> **备注**：
> - 前期开发阶段不考虑迁移兼容问题（数据库 schema 变更、数据迁移、向后兼容等暂缓处理）。
> - 前期开发阶段先写死账户密码和 token 方便调试，后期务必记得修正。

## 总览表

| # | 严重度 | 位置 | 问题描述 | 影响 |
|---|--------|------|----------|------|
| 1 | 🔴 严重（前期写死，后期务必修正） | `server/server.go:311` | 管理员 token `tk+3f3bc6d6...` **硬编码在源码里**（前期调试用），且 `Alive` 恒为 1 永不过期 | 拿到源码/二进制的人可直接冒充管理员，是硬编码后门，上线前必须移除 |
| 2 | 🔴 严重（前期写死，后期务必修正） | `server/install.go:15` | 管理员默认密码 `1234567890` 硬编码（前期调试用）且是弱密码 | 初始管理员可被轻易暴力破解，上线前必须移除 |
| 3 | 🔴 严重 | `server/server.go:27,180-189` | 密码哈希用「固定盐 + 单次 SHA256」，盐 `passwd_salt` 也硬编码 | 无 bcrypt/argon2，可被彩虹表/离线暴力破解 |
| 4 | 🔴 严重 | `database/database.go:50-56` | MySQL DSN 用 `fmt.Sprintf` 直接拼接密码，未转义 | 密码含 `@ : / # ?` 等字符会破坏 DSN，导致连接失败甚至注入 |
| 5 | 🟠 高 | `database/database.go:24,143-157` + `server/session.go:27-29` | token 的 `Alive` 字段从不为 0，无登出/吊销接口；30 分钟会话超时只删内存 session，DB 里 token 仍 `alive=1` | Token 事实上永久有效，README 里「Token超时」也标注未完成 |
| 6 | 🟠 高 | `server/user_api.go:99-194` | 登录成功从不更新 `Auth.LoginAt` | 返回的 `login_at` 永远是空字符串 |
| 7 | 🟠 高 | `database/user.go:91-136,139-182` | `CreateUser`/`CreateUserByID` 先建 Auth 再建 User，无事务；「先 count 再插入」非原子 | 中途失败留下孤儿记录；并发下可能绕过用户名唯一性检查 |
| 8 | 🟠 高 | `server/user_api.go:145-194` | 并发登录时「查无会话/无 token → 各自生成」无锁 | 同一用户可能产生多个 `alive` token |
| 9 | 🟡 中 | `server/server.go:69,78` + `server/install.go:36-38,67` | `Run()` 用局部变量 `is_installed` 判断，`Install()` 却检查实例字段 `s.is_installed`（两者不同步） | 字段与局部变量不一致，安装逻辑易误导、难维护 |
| 10 | 🟡 中 | `server/install.go:19-30` | `IsInstalled()` 仅凭 SQLite 文件是否存在判断 | 文件存在但为空/半安装时跳过建表建用户，后续查询报错 |
| 11 | 🟡 中 | `database/user.go:193-237,281-301` | `Updates(struct)` 只更新非零字段，无法把昵称/邮箱清空；`UpdateUserInfoByID` 未做用户名/邮箱唯一性校验 | 无法清空字段；可写入重复用户名/邮箱 |
| 12 | 🟡 中（前期暂不处理） | `database/database.go:19-25` | `Token` 结构给 `Token` 和 `ID` 同时标 `primaryKey`，形成复合主键；注释却写「ID 自增」但无 `autoIncrement` | 语义混乱，注释与实际行为不符；涉及 schema 变更，前期暂不处理 |
| 13 | 🟡 中 | `validator/validator.go:9-11` | `user_name_max_len=256`，但正则 `{2,15}` 实际最多 16 字符 | 长度上限常量与实际校验矛盾 |
| 14 | 🟡 中 | `server/user_api.go:79-84` | 登录时调用 `ValidateUserPassword` 校验密码格式 | 不符合该字符集的合法历史密码永远无法登录（校验应只在注册/改密时做） |
| 15 | 🟡 中 | `database/user.go:20-36` | `AuthInfoByToken` 只按 token 查询，不校验 `Alive` | 已失效 token 仍可查到用户信息 |
| 16 | 🟡 中 | `server/install.go:25-27` | MySQL 安装检查分支为空，恒 `return true` | MySQL 下永不执行建表/建管理员 |
| 17 | 🟡 中 | `config/config.go:41-46` + 两个 yml | 配置结构只定义了 `log.level/dir`、`server.http_addr`、`database.*`，示例中的 `max_size/max_backups/max_age/broadcast_port` 会被静默忽略；且 `docs/default.yml` 误用 `broadcast_port`（应为 `http_addr`），与 `release/config.yml` 不一致 | 这些字段不生效；用 `docs/default.yml` 作为配置时 `http_addr` 回退到默认 `127.0.0.1:7890` 而非预期的 28080 |
| 18 | 🟢 低 | `build_run.ps1:6` | 每次运行 `del ".\release\data.db"` | 每次启动都删除数据库，数据丢失/强制重装 |
| 19 | 🟢 低 | `validator/validator.go:80-95`、`server/tokens/token.go:44-49` | `regexp.MustCompile` 失败本身会 panic，后面的 `if == nil` 是死代码 | 冗余代码，无实际危害 |
| 20 | 🟢 低 | `server/server.go:41,46-56` | `snowflake.NewNode(1)` 创建的 `snowflake_node` 从未用于生成任何 ID（用户 ID 走数据库自增） | 死代码 |
| 21 | 🟢 低 | `server/server.go:346-360` | `WriteJsonSuccessResponse` JSON 编码失败时只写 500 状态码、不写 body | 客户端拿不到错误信息 |
| 22 | 🟢 低 | `server/user_api.go:68` | `json.NewDecoder(r.Body).Decode` 无 body 大小限制 | 超大请求体可消耗内存 |
| 23 | 🟢 低 | `database/user.go:74-81,270-278` | `GetAllUsers`/`GetAllAuthInfo` 无分页全量返回 | 数据量大时性能/内存问题 |
| 24 | 🟢 低 | `server/session.go:27-29` + `server/server.go:240-264` | 心跳是 `time.NewTimer` 一次性定时器，无续期 | 活跃会话 30 分钟后也会被强制删除 |
| 25 | 🟢 低 | `server/server.go:285-303` | `FindSessionByToken` 遍历整个 `sync.Map` 按 token 匹配 | O(n) 查找，会话多时性能差 |
| 26 | 🟢 低 | `server/server.go:203-220` | `GenerateToken` 混入 `UnixNano()` 时间戳，且与 UUID v7 冗余 | 轻微可预测性/冗余，非直接漏洞 |

## 优先级建议

1. **#1、#2、#3、#4**：均为可直接利用的安全问题。#1/#2 前期为方便调试有意写死，但上线前务必修正（移除硬编码 token/密码、改用 bcrypt、转义 DSN）。
2. **#5、#6**：token 生命周期和登录时间属于核心业务逻辑缺陷。
3. **#17**：配置字段与示例文件不一致，建议统一 `docs/default.yml` 与 `release/config.yml`，并将 `broadcast_port` 改为 `http_addr`。

## 附注

- 编译级验证（`go build ./...`）因沙箱只读策略被拦截（Go 需写 `AppData\Local\Temp` 构建缓存），故本清单基于静态分析。

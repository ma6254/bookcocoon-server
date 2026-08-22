# BookCocoon Server — 完整 Bug 清单

> 基于对所有非 vendor Go 源码文件的逐文件静态审查。
> 严重度分级：🔴 严重（安全/阻断） · 🟠 高（核心逻辑） · 🟡 中（功能缺陷） · 🟢 低（代码质量/边缘情况）

---

## 🔴 严重（安全 / 可被利用）

### BUG-01: 管理员 token 硬编码
- **位置**: `server/server.go:318`
- **代码**: `admin_token = "tk+3f3bc6d699c048fa9c8cf6d46ffc80b39d75db71b9183846bd30d72fc618ca71"`
- **问题**: 管理员 token 硬编码在源码中，`Alive` 恒为 1 永不过期。任何人拿到源码/二进制即可冒充管理员。
- **修复**: 移除硬编码，安装时生成随机 token 并写入配置文件或数据库。

### BUG-02: 管理员默认密码硬编码
- **位置**: `server/install.go:15`
- **代码**: `admin_user_default_passwd = "1234567890"`
- **问题**: 弱密码硬编码，可被暴力破解。
- **修复**: 安装时要求用户输入密码，或生成随机密码并提示首次登录修改。

### BUG-03: 密码哈希算法过弱
- **位置**: `server/server.go:187-196`
- **代码**: `Salt` 函数使用固定盐 + 单次 SHA256
- **问题**: 
  1. 盐 `s.Config.DataBase.Salt` 硬编码在配置中（且配置结构 `DataBase.Salt` 始终为空字符串，因为 `config.yml` 和默认配置都没有设置它）
  2. 单次 SHA256 可被彩虹表/离线暴力破解
  3. 盐实际为空，等同于裸 SHA256
- **修复**: 改用 bcrypt 或 argon2；为每个用户生成随机 salt 并存储。

### BUG-04: MySQL DSN 密码未转义
- **位置**: `database/database.go:50-56`
- **代码**: `fmt.Sprintf("%s:%s@tcp(...)..." , cfg.MySQL.Password, ...)`
- **问题**: 密码含 `@` `:` `/` `#` `?` 等字符会破坏 DSN 解析，导致连接失败或 SQL 注入。
- **修复**: 使用 `url.QueryEscape()` 转义密码等敏感字段。

### BUG-05: Token 无法吊销/登出
- **位置**: `database/database.go:231-238` + `server/session.go:27-29`
- **问题**: 
  1. `GetUserAllAliveToken` 查询 `alive=1` 的 token，但 `Alive` 字段永远不被置为 0
  2. 30 分钟会话超时只删除内存中的 `sync.Map` 条目，DB 中 token 仍 `alive=1`
  3. 无登出/吊销接口
  4. `AuthInfoByToken` (`database/user.go:20`) 和 `UserInfoByToken` (`database/user.go:240`) 不校验 `Alive` 字段
- **修复**: 实现登出接口将 `Alive` 置 0；心跳使用 `time.NewTimer` 而非 `time.AfterFunc` 并支持续期。

### BUG-06: 路径遍历漏洞
- **位置**: `server/upload_api.go:231-282`（`http_upload_read_api_handler`）
- **问题**: `file_path = r.PathValue("file_id")` 直接用于 `path.Join("uploads", file_path)`，未校验 `file_path` 是否包含 `..` 等路径遍历序列。攻击者可构造 `file_id=../etc/passwd` 读取任意文件。
- **修复**: 对 `file_path` 做 `filepath.Clean` 并校验不包含 `..`；或仅允许数字 file_id 并查库获取真实路径。

---

## 🟠 高（核心逻辑缺陷）

### BUG-07: 登录时间 `LoginAt` 从不更新
- **位置**: `server/user_api.go:56-196`
- **问题**: 登录成功后从未更新 `Auth.LoginAt` 字段。返回的 `login_at` 永远为空字符串。
- **修复**: 登录成功后调用 `s.DB.UpdateLoginAt(user.ID, time.Now())`。

### BUG-08: 用户创建非原子操作（竞态条件）
- **位置**: `database/user.go:91-136`（`CreateUser`）和 `database/user.go:139-182`（`CreateUserByID`）
- **问题**: 先 `Count` 检查再 `Create`，非原子操作。并发下可能绕过用户名唯一性约束，产生孤儿记录（Auth 已创建但 User 创建失败）。
- **修复**: 使用数据库事务或唯一索引约束。

### BUG-09: 并发登录产生多个有效 token
- **位置**: `server/user_api.go:145-194`
- **问题**: 两个并发登录请求可能同时发现 "无内存会话 + 无 DB token"，各自生成新 token，导致同一用户有多个有效 token。
- **修复**: 使用 `sync.Mutex` 或数据库行锁保护登录逻辑。

### BUG-09: 心跳定时器是一次性的
- **位置**: `server/session.go:27-29`
- **代码**: `heartbeat_timer = time.NewTimer(token_expire_time)`
- **问题**: `time.NewTimer` 是一次性的，30 分钟后自动触发删除，无法续期。活跃用户也会在 30 分钟后被强制踢出。
- **修复**: 使用 `time.AfterFunc` + 递归重置，或改用 `time.Ticker`。

### BUG-10: MySQL 安装检查恒为 true
- **位置**: `server/install.go:25-27`
- **代码**: MySQL 分支为空，直接 `return true`
- **问题**: MySQL 模式下永远跳过建表/建管理员流程。
- **修复**: 添加 MySQL 安装检查（如尝试查询 `SELECT 1` 或检查表是否存在）。

### BUG-11: `IsInstalled()` 判断不完整
- **位置**: `server/install.go:19-30`
- **问题**: 仅检查 SQLite 文件是否存在，文件存在但为空/半安装时跳过建表，后续查询报错。
- **修复**: 检查关键表（如 `auth`）是否存在。

### BUG-12: `Install()` 与 `Run()` 安装状态不同步
- **位置**: `server/server.go:76`（`is_installed` 局部变量） vs `server/install.go:36-38,67`（`s.is_installed` 实例字段）
- **问题**: `Run()` 用局部变量 `is_installed` 判断，`Install()` 检查并设置 `s.is_installed` 实例字段，两者不一致。
- **修复**: 统一使用 `s.is_installed` 实例字段。

### BUG-13: `GetAllUploads` / `GetAllUsers` 返回已删除数据
- **位置**: `database/upload.go:145` 和 `database/user.go:270`
- **问题**: 无软删除过滤，`GetAllUploads` 返回被软删除的上传记录；`GetAllUsers` 返回所有用户（包括已删除的）。
- **修复**: 添加 `Where("deleted = ?", false)` 过滤。

---

## 🟡 中（功能缺陷）

### BUG-14: `Updates(struct)` 无法清空字段
- **位置**: `database/user.go:193-237`
- **问题**: GORM `Updates(struct)` 只更新非零值字段，无法把昵称/邮箱清空为 `""`。
- **修复**: 使用 `map[string]interface{}` 或 `Update("field", "")`。

### BUG-15: 用户名最大长度常量与正则矛盾
- **位置**: `validator/validator.go:9-11`
- **代码**: `user_name_max_len = 256`，但正则 `{2,15}` 实际最多 16 字符
- **问题**: 长度上限常量与实际校验矛盾，维护者可能误以为支持到 256。
- **修复**: 将 `user_name_max_len` 改为 16 或更新正则。

### BUG-16: 登录时校验密码格式
- **位置**: `server/user_api.go:79-84`
- **问题**: 登录时调用 `ValidateUserPassword` 校验密码格式。不符合该字符集的合法历史密码（如含 `@`、空格等）永远无法登录。此校验应仅在注册/改密时做。
- **修复**: 登录时移除密码格式校验，只验证密码非空。

### BUG-17: 配置字段与示例文件不一致
- **位置**: `config/config.go:42-46` + `docs/default.yml` + `release/config.yml`
- **问题**: 
  1. 配置结构只定义了 `log.level/dir`、`server.http_addr`、`database.*`，示例中的 `max_size/max_backups/max_age/broadcast_port` 会被静默忽略
  2. `docs/default.yml` 误用 `broadcast_port`（应为 `http_addr`），与 `release/config.yml` 不一致
- **修复**: 统一配置结构和示例文件，移除 `broadcast_port`。

### BUG-18: `CheckUserAccount` 用户名校验过于严格
- **位置**: `database/user.go:332`
- **问题**: 用 `ValidateUserName` 校验登录账号，但用户名正则 `^[a-zA-Z][a-zA-Z0-9_-]{2,15}$` 不允许中文/其他 Unicode 字符。如果用户用邮箱登录，会走到邮箱校验分支；如果账号是纯数字，会走到 ID 分支。但如果账号是其他合法格式（如含 `.` 的用户名），三个校验全部失败，返回 `nil, nil`（用户不存在），而非有意义的错误。
- **修复**: 放宽校验或按顺序尝试 ID → 邮箱 → 用户名。

### BUG-19: `CreateBook` 不设置创建者
- **位置**: `database/book.go:23-30`
- **问题**: `Book` 结构无 `CreatedByID` 字段，无法追溯书籍来源。

---

## 🟢 低（代码质量 / 边缘情况）

### BUG-20: `build_run.ps1` 每次删除数据库
- **位置**: `build_run.ps1:6`
- **问题**: 每次运行都 `del ".\release\data.db"`，导致数据丢失/强制重装。
- **修复**: 删除该行或改为条件删除。

### BUG-21: `regexp.MustCompile` 失败后的 nil 检查是死代码
- **位置**: `validator/validator.go:121-141` 和 `server/tokens/token.go:44-49`
- **问题**: `regexp.MustCompile` 编译失败会 panic，后面的 `if == nil` 永远不执行。
- **修复**: 删除 nil 检查。

### BUG-22: `snowflake_node` 创建后从未使用
- **位置**: `server/server.go:53`
- **问题**: `snowflake.NewNode(1)` 生成的节点仅用于 `upload_api.go:106` 的 `file_id` 生成，但 `user_id` 走数据库自增。Snowflake 用于文件 ID 是合理的，但注释说"用户 ID 走数据库自增"误导。
- **修复**: 确认 `file_id` 是否确实使用了 snowflake（是的，`upload_api.go:106`），更新注释。

### BUG-23: `WriteJsonSuccessResponse` JSON 编码失败时不写 body
- **位置**: `server/server.go:361-363`
- **问题**: JSON 编码失败时只写 500 状态码，不写 body。客户端拿不到错误信息。
- **修复**: 在 `http.Error` 中写入错误信息（代码已做，但 `w.Header().Set("Content-Type")` 已设置，行为不一致）。

### BUG-24: 请求体无大小限制
- **位置**: `server/user_api.go:68`
- **问题**: `json.NewDecoder(r.Body).Decode` 无 body 大小限制，超大请求体可消耗内存。
- **修复**: 使用 `http.MaxBytesReader` 限制请求体大小。

### BUG-25: `GetAllUsers` / `GetAllAuthInfo` 无分页
- **位置**: `database/user.go:270-278` 和 `database/user.go:74-81`
- **问题**: 全量返回所有用户/认证信息，数据量大时性能/内存问题。
- **修复**: 添加分页参数。

### BUG-26: `FindSessionByToken` O(n) 查找
- **位置**: `server/server.go:292-309`
- **问题**: 遍历整个 `sync.Map` 按 token 匹配，会话多时性能差。
- **修复**: 维护 `map[string]*Session`（token → session）或使用 `sync.Map` 的 `Load` 按 key 查找。

### BUG-27: `GenerateToken` 混入时间戳
- **位置**: `server/server.go:210-227`
- **问题**: Token 混入 `time.Now().UnixNano()`，增加可预测性；且 UUID v7 本身已包含时间戳，冗余。
- **修复**: 移除 `UnixNano()` 部分。

### BUG-28: `AuthInfoByToken` 不校验 `Alive`
- **位置**: `database/user.go:20-36`
- **问题**: 只按 token 查询，不校验 `Alive` 字段，已失效 token 仍可查到用户信息。
- **修复**: 添加 `WHERE alive = 1` 条件。

### BUG-29: `UserInfoByToken` 不校验 `Alive`
- **位置**: `database/user.go:240-256`
- **问题**: 同上，不校验 `Alive`。
- **修复**: 添加 `WHERE alive = 1` 条件。

### BUG-30: `UpdateUploadSize` 在文件保存前调用
- **位置**: `server/upload_api.go:197`
- **问题**: `UpdateUploadSize` 用 `r.ContentLength` 设置大小，但此时文件尚未保存（`saveUploadFile` 在之后调用）。如果 `ContentLength` 不准确（如 chunked 传输），文件大小会错误。
- **修复**: 在文件保存后根据实际写入的字节数更新大小。

### BUG-31: `GetBookByID` 双重查询
- **位置**: `database/book.go:40-53`
- **问题**: 先 `Count` 再 `First`，两次查询可合并为一次。
- **修复**: 直接 `First`，检查 `gorm.ErrRecordNotFound`。

### BUG-32: `CheckBookID` 双重查询
- **位置**: `database/book.go:89-99`
- **问题**: 同上，`Count` + 隐式 `First`（由调用方使用 `ErrRecordNotFound` 判断）。
- **修复**: 直接 `First`。

### BUG-33: `CreateUpload` 双重查询
- **位置**: `database/upload.go:22-56`
- **问题**: 先 `FindUploadByHash` 检查是否存在，再 `Create`。可合并为 `Create` 后检查唯一约束错误。
- **修复**: 直接 `Create`，捕获唯一约束错误。

### BUG-34: `CreateUserByID` 未检查 User 表重复
- **位置**: `database/user.go:139-182`
- **问题**: 只检查 `Auth` 表是否有重复 `id`，未检查 `User` 表。如果 `User` 表已有该 `id`，`Create` 会失败。
- **修复**: 同时检查 `User` 表。

### BUG-35: `CreateReadingRecord` 返回 `nil, nil` 而非错误
- **位置**: `database/reading_record.go:28-30`
- **问题**: 记录已存在时返回 `nil, nil`，调用方无法区分"记录已存在"和"数据库错误"。
- **修复**: 返回明确的错误如 `gorm.ErrDuplicatedKey` 或自定义错误。

---

## 优先级建议

### 立即修复（上线前）
1. **BUG-01, BUG-02, BUG-03**: 安全漏洞 — 硬编码凭证 + 弱哈希
2. **BUG-04**: MySQL DSN 注入风险
3. **BUG-05**: Token 永不过期
4. **BUG-06**: 路径遍历漏洞

### 尽快修复
5. **BUG-07**: 登录时间不更新
6. **BUG-08, BUG-09**: 竞态条件
7. **BUG-10, BUG-11**: MySQL 安装检查缺失
8. **BUG-12**: 安装状态不一致

### 中期优化
9. **BUG-14 ~ BUG-18**: 功能缺陷
10. **BUG-24 ~ BUG-26**: 性能和安全性优化

### 长期改进
11. **BUG-20 ~ BUG-35**: 代码质量、冗余查询、边缘情况

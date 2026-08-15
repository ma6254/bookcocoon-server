---
name: http-api-design
description: HTTP API 设计原则
---

# http_api_design

用于设计、评审或重构 HTTP API。重点保证接口语义清晰、资源建模稳定、错误响应一致、鉴权边界明确，并能与 OpenAPI / Swagger 文档保持同步。

## 核心原则

- 先明确业务资源，再设计路径；路径表达“对象”，HTTP 方法表达“动作”。
- API 必须稳定、可演进、可观测；避免把一次性页面交互直接暴露为长期接口契约。
- 请求、响应、错误、鉴权、分页、时间格式等规则必须全局一致。
- 默认返回 JSON；字段名使用 `snake_case`，与 Go 结构体 `json` tag 保持一致。
- 不在 URL 中暴露敏感信息；令牌、会话、签名等放在 Header 中传递。
- 接口实现要与文档同步更新，尤其是路由、方法、状态码、请求体和响应体。

## 路径与方法

- 路径使用小写名词和层级结构：`/api/users`、`/api/users/{user_id}`。
- 集合用复数名词，单个资源用路径参数，不用动词拼接：优先 `/api/users/{user_id}`，避免 `/api/getUser`。
- 使用 HTTP 方法表达意图：
  - `GET`：读取资源，不产生副作用。
  - `POST`：创建资源或执行非幂等操作。
  - `PUT`：整体替换资源，幂等。
  - `PATCH`：局部更新资源。
  - `DELETE`：删除资源，通常设计为幂等。
- 非 CRUD 行为优先建模为资源或子资源，例如登录可用 `POST /api/user/login`，心跳可用 `POST /api/heartbeat`。
- 当前 Go `net/http` 项目可使用方法匹配路由写法：`"POST /api/user/login"`。

## 请求设计

- JSON 请求体必须定义明确结构，不直接使用无约束的 `map[string]any` 作为业务输入。
- 必填字段在服务端显式校验；缺失、格式错误、非法范围统一返回 `400 Bad Request`。
- 查询参数用于过滤、排序、分页和轻量开关；复杂条件优先使用请求体。
- 时间字段统一使用 ISO 8601 / RFC3339 字符串；避免混用本地化格式。
- 布尔、枚举、分页参数要有默认值和合法值说明。
- 大字段、二进制内容和文件上传不要塞入普通 JSON；单独设计上传接口。

## 响应设计

- 成功响应返回稳定 JSON 对象；即使只有一个值，也优先使用对象包裹，例如 `{ "token": "..." }`。
- 列表响应包含数据和分页信息，建议结构：

  ```json
  {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
  ```

- 删除或无返回体操作可返回 `204 No Content`；如果需要兼容简单客户端，也可返回 `200 OK` 加简短 JSON。
- 服务端错误不要泄露数据库 SQL、堆栈、密钥、内部路径等敏感信息。
- 响应 Header 明确设置 `Content-Type: application/json`，除非确实没有响应体。

## 状态码约定

- `200 OK`：读取、更新、动作执行成功。
- `201 Created`：资源创建成功，可配合 `Location` Header。
- `204 No Content`：成功但无响应体。
- `400 Bad Request`：请求体、参数或字段校验失败。
- `401 Unauthorized`：未登录、令牌缺失、令牌无效或过期。
- `403 Forbidden`：已认证但无权限。
- `404 Not Found`：资源不存在或不应暴露其存在。
- `409 Conflict`：资源状态冲突，例如重复创建、版本冲突。
- `422 Unprocessable Entity`：语法正确但业务规则不满足；如项目未区分，可统一用 `400`。
- `429 Too Many Requests`：触发限流。
- `500 Internal Server Error`：未预期服务端错误。

## 错误响应

- 推荐统一错误结构，便于前端、测试和日志关联：

  ```json
  {
    "error": {
      "code": "invalid_request",
      "message": "Bad Request",
      "details": []
    }
  }
  ```

- `message` 面向调用方，简洁可读；内部诊断放日志，不放响应。
- `code` 使用稳定机器可读字符串，不直接依赖英文错误句子。
- 对认证失败统一模糊处理，避免泄露“用户名存在但密码错误”等细节。
- Go 实现中避免多次 `WriteHeader`；一旦写出状态码或响应体，立即 `return`。

## 鉴权与会话

- 令牌默认使用 `Authorization` Header；如采用 Bearer 方案，格式为 `Authorization: Bearer <token>`。
- 不把 token 放在 query string，避免被日志、浏览器历史、代理缓存泄露。
- 登录接口只返回必要凭证和基础会话信息，不返回密码散列、盐值、内部用户 ID 等敏感字段。
- 登出、心跳、资料读取等需要登录态的接口必须统一走鉴权中间件或包装函数。
- 会话过期、令牌无效、令牌格式错误统一返回 `401`。
- 对密码、token、session id 的日志输出必须脱敏或禁止。

## 分页、过滤与排序

- 分页参数建议统一为 `page` / `page_size`，或游标分页 `cursor` / `limit`。
- `page_size` 必须设置最大值，防止一次请求拖垮服务。
- 排序参数建议使用 `sort`，例如 `sort=created_at` 或 `sort=-created_at`。
- 过滤参数应与资源字段语义一致，例如 `department=rd`、`status=active`。
- 返回结果要保证稳定排序，避免分页时漏数据或重复数据。

## 版本与兼容

- 对外长期 API 建议使用版本前缀：`/api/v1/...`；内部或早期项目可先使用 `/api/...`，但要保留迁移空间。
- 向后兼容优先：新增字段通常安全，删除字段、改字段含义、改状态码属于破坏性变更。
- 废弃字段或接口需在文档中标记，并给出替代接口和迁移窗口。
- 不让客户端依赖响应字段顺序；字段语义和类型才是契约。

## OpenAPI / Swagger 文档

- 新增或修改接口时，同步更新 OpenAPI 文档，包括：路径、方法、请求体、响应体、状态码、鉴权要求。
- 文档示例必须可真实调用，避免使用过期字段或伪造枚举值。
- Schema 命名与 Go 类型命名保持可追踪，例如 `LoginForm`、`LoginResponse`、`UserInfo`。
- 文档中明确标注哪些接口需要 `Authorization` Header。
- 如果项目暴露 `/swagger/` 或 `/docs/swagger.yaml`，改完接口后检查文档能正常访问和解析。

## Go `net/http` 实现建议

- Handler 只负责 HTTP 边界：解析请求、校验参数、调用业务逻辑、写响应。
- 重复逻辑抽成 helper：`writeJSON`、`writeError`、`decodeJSON`、`requireAuth`。
- 写 JSON 前设置 Header：`w.Header().Set("Content-Type", "application/json")`。
- `json.NewDecoder(r.Body).Decode(&req)` 后要处理错误，并限制请求体大小以防滥用。
- 对所有分支保持“一次写响应 + 立即 return”，避免重复写头。
- 日志包含接口名、用户/会话标识、错误摘要；不要记录密码、完整 token、请求体敏感字段。

## 接口评审清单

- [ ] 路径是否是资源名词，方法是否表达动作语义？
- [ ] 请求字段是否有明确类型、必填规则、默认值和合法范围？
- [ ] 成功响应是否稳定、字段命名是否一致？
- [ ] 错误结构和状态码是否与全局约定一致？
- [ ] 鉴权、权限、会话过期行为是否明确？
- [ ] 是否考虑分页、排序、限流、幂等性和并发冲突？
- [ ] 是否避免泄露敏感信息到 URL、响应和日志？
- [ ] OpenAPI / Swagger 文档是否同步更新？

## 输出模板

当用户要求设计或评审 HTTP API 时，优先按以下格式输出：

```text
接口设计建议：

- 路径与方法：<METHOD> <PATH>
- 请求参数：<body/query/header 的关键字段>
- 成功响应：<状态码 + 响应结构>
- 错误响应：<主要状态码 + 错误 code>
- 鉴权要求：<是否需要 Authorization / 权限范围>
- 文档变更：<需要同步的 OpenAPI / Swagger 项>
```

## 示例

登录接口：

```text
POST /api/user/login

Request JSON:
{
  "user_name": "alice",
  "password": "******"
}

Response 200:
{
  "token": "tk+..."
}

错误：
- 400：请求体错误或字段缺失
- 401：用户名或密码无效
- 500：服务端内部错误
```

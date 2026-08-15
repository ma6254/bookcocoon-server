---
name: swagger-user-guide
description: Swagger 用户指南
---

# swagger_user_guide

用于指导用户访问、阅读、调试和维护 Swagger / OpenAPI 文档。重点保证接口文档可访问、可测试、与服务端路由一致，并能准确说明当前项目的鉴权方式、文档路径和常见问题。

## 核心原则

- Swagger 文档是接口契约，不只是展示页面；路径、方法、参数、状态码、响应体必须与代码一致。
- 用户指南应先告诉用户“在哪里打开”，再说明“怎么登录、怎么填 token、怎么调用接口”。
- 示例请求必须能真实调用；不要使用已经失效的 schema、路径、字段名或鉴权格式。
- 文档问题要区分：UI 无法打开、YAML/JSON 无法加载、接口调用失败、鉴权失败、schema 与实现不一致。
- 当前项目使用 Swagger 2.0 文档文件，并通过 `http-swagger` 提供网页 UI。
- 修改 API、路由、请求体、响应体或鉴权逻辑后，必须同步更新 Swagger 文档。

## 当前项目访问方式

- 启动服务后，通过浏览器访问 Swagger UI：

  ```text
  http://<server_http_addr>/swagger/
  ```

- 当前默认 HTTP 地址来自配置默认值：

  ```text
  http://127.0.0.1:28080/swagger/
  ```

- 原始 Swagger YAML 文档路径：

  ```text
  http://<server_http_addr>/docs/swagger.yaml
  ```

- 原始 Swagger JSON 文档路径：

  ```text
  http://<server_http_addr>/docs/swagger.json
  ```

- 项目路由中 `/docs/` 使用嵌入式静态文件，`/swagger/` 使用 `httpSwagger.Handler(httpSwagger.URL("/docs/swagger.yaml"))`。

## 使用流程

1. 启动服务，确认控制台没有配置、端口、数据库初始化错误。
2. 打开 `/swagger/`，确认页面能加载接口列表。
3. 如页面空白或报错，直接打开 `/docs/swagger.yaml`，确认 YAML 能正常访问。
4. 调用 `POST /api/user/login`，填写 `user_name` 和 `password`。
5. 从登录响应中复制 `token`。
6. 点击 Swagger UI 的 `Authorize`，填入 token。
7. 调用需要鉴权的接口，例如 `GET /api/user/profile`、`GET /api/user/logout`、`POST /api/heartbeat`。

## 鉴权说明

- 当前服务端从请求头 `Authorization` 读取 token。
- 当前 token 格式由服务端生成，形如：

  ```text
  tk+<64位十六进制字符串>
  ```

- 当前实现校验的是原始 token 字符串，因此 Swagger Authorize 中应填写：

  ```text
  tk+...
  ```

- 不要填写 `Bearer tk+...`，除非服务端已实现 Bearer 前缀解析。
- 文档中的 security definition 可以命名为 `Bearer`，但实际 Header 值必须与服务端 `CheckTokenValid` 逻辑一致。
- 如果希望使用标准 Bearer 方案，需要同步修改服务端解析逻辑和 Swagger 示例：

  ```text
  Authorization: Bearer <token>
  ```

## 当前接口速览

- `POST /api/user/login`：用户登录，不需要 token。
- `GET /api/user/logout`：用户登出，需要 `Authorization` Header。
- `GET /api/user/profile`：获取用户资料，需要 `Authorization` Header。
- `POST /api/heartbeat`：心跳检测，需要 `Authorization` Header。

## 请求与响应填写建议

- 登录请求体使用 JSON：

  ```json
  {
    "user_name": "admin",
    "password": "password"
  }
  ```

- 所有 JSON 请求都应设置 `Content-Type: application/json`。
- Swagger 示例中的字段名必须和 Go struct 的 `json` tag 一致，例如 `user_name`、`password`。
- 成功响应示例应展示真实结构；当前登录实现返回：

  ```json
  {
    "token": "tk+..."
  }
  ```

- 不要用通用 `{ "status": "ok" }` 替代真实响应，除非接口实现确实返回该结构。

## 文档维护规范

- 新增路由时，同步新增 Swagger `paths` 项，包含 method、summary、description、tags、parameters、responses。
- 修改请求体时，同步更新 `definitions` 或 schema。
- 修改响应结构时，同步更新 `responses` 示例和 schema。
- 修改鉴权要求时，同步更新 `securityDefinitions` 和每个接口的 `security`。
- 删除或废弃接口时，在文档中标注 deprecated，或同步删除已不可用接口。
- 文档标题、版本、描述要与实际项目阶段一致。

## Swagger 2.0 注意事项

- 当前 `docs/swagger.yaml` 使用：

  ```yaml
  swagger: "2.0"
  ```

- Swagger 2.0 的响应 schema 通常写在 `responses.<code>.schema` 下，而不是 OpenAPI 3 的 `content.application/json.schema`。
- 如果继续使用 Swagger 2.0，应避免混用 OpenAPI 3 写法。
- 如果需要使用 `content`、`requestBody` 等 OpenAPI 3 字段，应整体升级为：

  ```yaml
  openapi: 3.0.3
  ```

- 评审文档时要检查规范版本和字段写法是否匹配。

## 常见问题排查

- Swagger UI 打不开：检查服务是否启动、监听地址是否正确、端口是否被占用。
- UI 能打开但接口列表为空：直接访问 `/docs/swagger.yaml`，检查 YAML 是否能加载和解析。
- 点击 Try it out 后 404：检查 Swagger path/method 是否与 `server/route.go` 一致。
- 登录接口 400：检查 JSON 格式、`user_name` / `password` 是否存在。
- 登录接口 401：用户名或密码无效，或初始化用户不存在。
- 鉴权接口 401：检查是否填写 `Authorization`，是否误加 `Bearer ` 前缀，token 是否过期或会话不存在。
- 响应结构与文档不一致：以服务端实现为准，立即更新 Swagger schema 和示例。
- YAML 修改后页面没变化：确认嵌入资源是否重新构建，服务是否重启，浏览器是否缓存旧文档。

## 安全提示

- 不要在 Swagger 示例中写真实用户名、密码、token、数据库地址或内部 IP。
- 不要把生产环境 Swagger UI 暴露给公网，除非有额外访问控制。
- 登录接口示例密码使用占位值，例如 `******` 或 `<password>`。
- token 示例只保留前缀和省略号，例如 `tk+...`。
- 对外发布文档前，检查是否泄露内部路径、测试账号、错误堆栈或服务拓扑。

## 用户说明模板

当用户询问如何使用 Swagger 调试接口时，优先按以下格式输出：

```text
Swagger 使用步骤：

1. 打开 Swagger UI：<base_url>/swagger/
2. 如页面异常，检查原始文档：<base_url>/docs/swagger.yaml
3. 先调用登录接口：POST /api/user/login
4. 复制返回的 token
5. 点击 Authorize，填入 Authorization: <token>
6. 再调用需要鉴权的接口

注意：当前项目 Authorization 直接填写 tk+...，不要加 Bearer 前缀，除非服务端已支持。
```

## 文档评审清单

- [ ] Swagger UI 路径和原始 YAML/JSON 路径是否可访问？
- [ ] `paths` 中的路径和方法是否与 `server/route.go` 一致？
- [ ] 请求字段是否与 Go struct 的 `json` tag 一致？
- [ ] 响应 schema 和示例是否与真实 handler 返回一致？
- [ ] 需要鉴权的接口是否声明 `security`？
- [ ] Authorize 填写说明是否匹配服务端 token 解析逻辑？
- [ ] Swagger/OpenAPI 版本是否与字段写法一致？
- [ ] 示例中是否避免真实密码、token 和内部敏感信息？

## 示例

当前项目本地调试示例：

```text
1. 启动服务：
   lnmts_server --config ./config.yml

2. 打开 Swagger UI：
   http://127.0.0.1:28080/swagger/

3. 调用登录接口：
   POST /api/user/login
   Body: { "user_name": "admin", "password": "******" }

4. 复制响应 token：
   { "token": "tk+..." }

5. 点击 Authorize，填入：
   tk+...

6. 调用：
   GET /api/user/profile
   POST /api/heartbeat
```

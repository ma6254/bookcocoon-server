---
name: local-git-summary
description: 本地 Git 只读检查与中文提交信息生成流程
---

# local_git

用于根据当前仓库未提交变更生成**尽可能简短的中文 Git 提交信息**。默认只读检查，不执行提交。

## 核心原则

- 优先使用 `git_mcp` 查看仓库状态和 diff 摘要。
- 如果 `git_mcp` 仓库授权不匹配、连接不可用或无法访问目标仓库，先向用户说明原因；经用户允许后，可改用本地 `git` 只读命令。
- 只生成提交信息，**禁止执行** `git add`、`git commit`、`git push` 等会改变仓库状态的命令。
- 所有本地 Git 检查必须避免 pager 卡住：使用 `git --no-pager ...`，不要直接运行可能进入 `more` / `less` 的完整 diff。
- 输出要简短，通常只给 1 条推荐提交信息，并明确说明未提交。

## 推荐只读检查流程

1. 查看工作区状态：

   ```bash
   git --no-pager status --short
   ```

2. 查看变更规模摘要：

   ```bash
   git --no-pager diff --stat
   git --no-pager diff --cached --stat
   ```

3. 查看文件级变更类型：

   ```bash
   git --no-pager diff --name-status
   git --no-pager diff --cached --name-status
   ```

4. 如需理解新增文件，只读取目标文件的有限内容；避免一次性输出大型完整 diff。

5. 如命令环境是 PowerShell 且 `&&` 报错，可改用分号或包一层 `cmd /d /c`，但仍必须使用 `git --no-pager`。

## 提交信息格式

- 默认使用 Conventional Commit 风格：`type: 中文描述`
- 尽量一行完成，不写长正文。
- 常用类型：
  - `feat`: 新功能
  - `fix`: 修复问题
  - `refactor`: 重构/拆分代码
  - `docs`: 文档变更
  - `chore`: 构建、配置、忽略文件等杂项
  - `style`: 格式调整，无逻辑变化
  - `test`: 测试相关

## 输出模板

```text
推荐提交信息：

<type>: <尽可能简短的中文描述>

未执行 git commit。
```

## 示例

当变更包括拆分 `server/api.go`，新增 `server/route.go` / `server/user_api.go`，并调整 Swagger 心跳接口方法时，可输出：

```text
refactor: 拆分路由和用户接口
```

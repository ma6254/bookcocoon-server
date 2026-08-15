---
name: powershell-terminal-exec
description: PowerShell 终端执行注意事项
---

# powershell_terminal_exec

用于在 Windows PowerShell / PowerShell 7 终端中安全、稳定地执行命令。重点避免命令连接符、路径引号、编码、别名、重定向、退出码和交互式命令造成的误判或卡住。

## 核心原则

- 先确认当前终端是 Windows PowerShell 5.x、PowerShell 7+、CMD、Git Bash 还是 WSL；不同 Shell 语法不能混用。
- PowerShell 中优先使用 PowerShell 原生命令与语法；需要 CMD 行为时显式包一层 `cmd /d /c "..."`。
- 命令必须尽量非交互式，避免等待输入、分页器、确认提示或后台服务长期占用当前会话。
- 路径、参数和字符串必须正确加引号，特别是包含空格、中文、括号、`&`、`;` 的路径。
- 对会修改文件、安装依赖、删除目录、提交代码等操作，要先确认目标路径和影响范围。
- 命令执行失败时先看退出码和错误流，不要只根据最后一行输出判断成功。

## 命令连接与执行顺序

- PowerShell 7 支持 `&&` 和 `||`，但 Windows PowerShell 5.1 不支持；为兼容性优先使用分号或显式判断。
- 顺序执行命令可用分号：

  ```powershell
  git --no-pager status --short; git --no-pager diff --stat
  ```

- 仅前一个命令成功才执行下一个命令，兼容写法：

  ```powershell
  go test ./...
  if ($LASTEXITCODE -eq 0) { go build ./... }
  ```

- 需要 CMD 的 `&&` 行为时：

  ```powershell
  cmd /d /c "go test ./... && go build ./..."
  ```

- 不要在不确定 Shell 版本时直接使用 Bash 写法：

  ```bash
  command1 && command2
  ```

## 路径与引号

- Windows 路径建议用双引号包裹：

  ```powershell
  Get-Content "C:\Users\Name With Space\config.yml"
  ```

- 相对路径建议从当前工作目录出发，避免隐式依赖 IDE 打开的目录。
- 包含通配符 `*`、`[`、`]` 的路径如果要按字面量处理，使用 `-LiteralPath`：

  ```powershell
  Remove-Item -LiteralPath ".\data[backup].db"
  ```

- 调用包含空格路径的可执行文件，用调用运算符 `&`：

  ```powershell
  & "C:\Program Files\Git\bin\git.exe" --version
  ```

- 不要把 `&` 当作 Bash 后台运行符；在 PowerShell 中它是调用运算符。

## 引号与变量展开

- 单引号不展开变量，双引号会展开变量：

  ```powershell
  '$env:PATH'
  "$env:PATH"
  ```

- JSON 字符串建议使用单引号包裹外层，减少双引号转义：

  ```powershell
  curl.exe -X POST "http://127.0.0.1:28080/api/user/login" -H "Content-Type: application/json" -d '{"user_name":"admin","password":"******"}'
  ```

- 多行字符串使用 here-string：

  ```powershell
  $body = @'
  {
    "user_name": "admin",
    "password": "******"
  }
  '@
  ```

- 参数中包含 `$` 时要注意变量展开；如需字面值可用单引号或反引号转义。

## 常见别名陷阱

- Windows PowerShell 中 `curl` 可能是 `Invoke-WebRequest` 的别名，不一定是 curl 可执行文件。
- 需要真实 curl 行为时使用：

  ```powershell
  curl.exe --version
  ```

- `ls`、`cat`、`rm`、`cp`、`mv` 在 PowerShell 中通常是别名，参数语义与 Linux 命令不同。
- 跨平台命令示例中，优先写完整 PowerShell 命令：

  ```powershell
  Get-ChildItem
  Get-Content
  Remove-Item
  Copy-Item
  Move-Item
  ```

- 如果必须使用 Unix 工具，先确认 Git Bash、MSYS、WSL 或对应 exe 是否可用。

## 输出、编码与重定向

- PowerShell 5.1 默认编码可能不是 UTF-8，处理中文文件时要显式指定编码。
- 写入 UTF-8 文件：

  ```powershell
  Set-Content -Path ".\README.md" -Value $content -Encoding UTF8
  ```

- 追加文件：

  ```powershell
  Add-Content -Path ".\log.txt" -Value "message" -Encoding UTF8
  ```

- 捕获 stdout 和 stderr：

  ```powershell
  go test ./... *> test-output.txt
  ```

- 兼容传统重定向：

  ```powershell
  go test ./... > test-output.txt 2>&1
  ```

- 不要把包含中文的文件内容通过不明确编码的命令反复重写。

## 退出码与错误判断

- PowerShell cmdlet 的成功状态通常看 `$?`。
- 外部程序（如 `go`、`git`、`npm`）的退出码看 `$LASTEXITCODE`。
- 示例：

  ```powershell
  go test ./...
  if ($LASTEXITCODE -ne 0) { throw "go test failed: $LASTEXITCODE" }
  ```

- `Write-Error` 不一定等同于进程失败；脚本中需要显式 `exit 1` 或 `throw`。
- 自动化脚本建议设置：

  ```powershell
  $ErrorActionPreference = "Stop"
  ```

- 但对可能返回非终止错误的命令，要单独处理，避免误中断后续清理逻辑。

## Git 命令注意事项

- 避免 pager 卡住，统一使用：

  ```powershell
  git --no-pager status --short
  git --no-pager diff --stat
  git --no-pager diff --name-status
  ```

- 不要直接运行可能进入分页器的完整 diff，除非明确需要人工查看。
- 分支名、文件名包含特殊字符时加引号。
- 只读检查优先，不要自动执行 `git add`、`git commit`、`git push` 等改变仓库状态的命令。
- PowerShell 中换行续写使用反引号 `` ` ``，但容易因尾随空格失效；长命令优先拆成变量或数组。

## Node / Go / Python 命令注意事项

- 安装依赖前确认当前目录、锁文件和包管理器，不要混用 `npm` / `pnpm` / `yarn`。
- Go 项目常用只读或构建命令：

  ```powershell
  go test ./...
  go build ./...
  go run . --help
  ```

- Python 在 Windows 上可能是 `python` 或 `py`，执行前可检查：

  ```powershell
  python --version
  py --version
  ```

- 长时间运行服务前说明端口和停止方式，必要时让用户确认。
- 后台启动服务需记录 PID 或停止命令，避免残留进程占用端口。

## 删除、覆盖与危险操作

- 删除文件前先列出目标：

  ```powershell
  Get-ChildItem -LiteralPath ".\target"
  ```

- 删除目录使用明确路径，谨慎使用 `-Recurse -Force`：

  ```powershell
  Remove-Item -LiteralPath ".\dist" -Recurse -Force
  ```

- 不要在变量为空或未校验时拼接删除命令：

  ```powershell
  Remove-Item "$dir\*" -Recurse -Force
  ```

- 覆盖文件前确认目标是生成物还是用户手写文件。
- 清理缓存、数据库、日志、构建产物等操作要明确影响范围。

## 网络请求与 API 调试

- 使用真实 curl 时写 `curl.exe`，避免被 `Invoke-WebRequest` 别名影响。
- JSON 请求示例：

  ```powershell
  curl.exe -X POST "http://127.0.0.1:28080/api/user/login" `
    -H "Content-Type: application/json" `
    -d '{"user_name":"admin","password":"******"}'
  ```

- PowerShell 原生命令示例：

  ```powershell
  Invoke-RestMethod -Method Post `
    -Uri "http://127.0.0.1:28080/api/user/login" `
    -ContentType "application/json" `
    -Body '{"user_name":"admin","password":"******"}'
  ```

- Header 示例：

  ```powershell
  Invoke-RestMethod -Uri "http://127.0.0.1:28080/api/user/profile" -Headers @{ Authorization = "tk+..." }
  ```

- 不要在示例中写真实 token、密码或生产地址。

## 输出模板

当用户要求给出 PowerShell 命令时，优先按以下格式输出：

```text
PowerShell 命令：

<命令>

注意：
- 当前目录：<需要在哪个目录执行>
- 影响范围：<只读 / 会写文件 / 会启动服务 / 会删除文件>
- 兼容性：<Windows PowerShell 5.1 / PowerShell 7 / 需要 cmd 包裹>
```

## 执行前检查清单

- [ ] 是否确认当前 Shell 是 PowerShell，而不是 CMD / Bash / WSL？
- [ ] 是否避免了 PowerShell 5.1 不支持的 `&&` / `||`？
- [ ] 路径是否加引号，特殊路径是否使用 `-LiteralPath`？
- [ ] 是否避免 `curl`、`ls`、`cat`、`rm` 等别名语义误判？
- [ ] 是否明确编码，避免中文文件乱码？
- [ ] 是否处理外部程序退出码 `$LASTEXITCODE`？
- [ ] 是否避免交互式命令、pager、确认提示导致卡住？
- [ ] 删除、覆盖、安装、提交、推送等危险操作是否已确认影响范围？

## 示例

兼容 Windows PowerShell 5.1 的测试与构建：

```powershell
go test ./...
if ($LASTEXITCODE -eq 0) {
  go build ./...
}
```

查看 Git 状态和摘要，避免 pager：

```powershell
git --no-pager status --short; git --no-pager diff --stat
```

需要 CMD 连接符时：

```powershell
cmd /d /c "go test ./... && go build ./..."
```
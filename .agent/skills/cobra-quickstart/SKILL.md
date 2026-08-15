---
name: cobra-quickstart
description: 编写和搭建基于 spf13/cobra 的 Go CLI 应用时的权威参考。当用户需要创建命令行工具、新增子命令、解析 flags 与位置参数、生成 shell 自动补全，或提到 cobra、Go CLI、命令行入口、cmd 包时使用——即使没有明确说出 cobra 这个名字。覆盖从零搭建最小 CLI、常见任务代码模板、本项目（bookcocoon-server）的既定约定。
---

# Cobra 快速开始指南

基于 spf13/cobra（本项目使用 v1.10.x，Go 1.23+）的 Go CLI 开发参考。目标是让任何 CLI 任务都能照着模板直接落地，且新代码与项目现有风格保持一致。

## 心智模型

Cobra 的一切围绕 `cobra.Command` 展开：

- **Command** = 一个命令（根命令或子命令），一棵 Command 树就是整个 CLI
- **Flags** 挂在 Command 上，分两种：`Flags()` 局部（只属于本命令）、`PersistentFlags()` 持久（本命令及所有子孙命令可见）
- **Args** 是命令后跟的位置参数，由 `Args` 字段声明的校验器约束
- 程序入口 `main()` 只做一件事：调用 `Execute()`

先浏览本项目现有代码作为基准：`cmd/root.go`（根命令与 flags 定义）、`main.go`（入口）。新增代码应延续其中的模式。

## 依赖与脚手架

```bash
# 已有 go.mod 的项目中引入 cobra（本项目已引入 v1.10.2，无需重复执行）
go get github.com/spf13/cobra

# 可选：cobra-cli 脚手架工具，用于生成样板代码
go install github.com/spf13/cobra-cli@latest
cobra-cli init   # 生成 main.go + cmd/root.go
cobra-cli add serve  # 生成 cmd/serve.go 子命令
```

注意：新版 cobra-cli 生成的样板已不再默认集成 viper。若项目需要配置文件解析，自行引入 viper 或直接手写 YAML/JSON 解析——不要假设样板里会有。

## 最小可运行结构

`main.go`：

```go
package main

import "github.com/ma6254/bookcocoon-server/cmd"

func main() {
	cmd.Execute()
}
```

`cmd/root.go`（骨架，与本项目现有文件一致）：

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "一句话描述应用",
	Long:  "详细描述，出现在 help 输出顶部",
	Run: func(cmd *cobra.Command, args []string) {
		// 不带子命令时的默认行为
		fmt.Println("hello")
	},
}

// Execute 由 main.main() 调用，只需调用一次
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// 局部 flag：仅根命令可用
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./config.yml", "config file path")
}
```

要点解释：

- `Use` 同时决定命令名和用法行（usage line），cobra 会基于它自动生成 `help` 输出，不要手写帮助文本
- `Short` 必须短（一行摘要，出现在父命令的命令列表中）；`Long` 只在 `myapp help` 或 `myapp --help` 时完整展示
- flags 定义放在 `init()` 是社区惯例，cobra-cli 生成的代码也如此——这样定义与命令声明分离，文件再长也好定位

## 常见任务模板

### 新增子命令

`cmd/serve.go`：

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务",
	Args:  cobra.NoArgs, // 位置参数校验，见下节
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer(cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "listen port")
}
```

子命令注册发生在各自文件的 `init()` 里，通过 `rootCmd.AddCommand` 挂载。这样新增命令只加文件、不改 root.go，符合 cobra-cli 的项目布局。

### 位置参数校验（Args）

用 cobra 内置校验器表达参数约束，而不是在 Run 里手写 `len(args)` 检查——声明式校验会同时修正 help 输出中的用法行，用户不猜也知道参数怎么传：

| 校验器 | 含义 |
|---|---|
| `cobra.NoArgs` | 不允许任何参数 |
| `cobra.ArbitraryArgs` | 任意参数（默认行为） |
| `cobra.ExactArgs(n)` | 恰好 n 个 |
| `cobra.MinimumNArgs(n)` / `cobra.MaximumNArgs(n)` | 最少 / 最多 n 个 |
| `cobra.RangeArgs(min, max)` | 闭区间 |
| `cobra.OnlyValidArgs` | 只接受 `ValidArgs` 字段列出的值 |
| `cobra.MatchAll(fns...)` / `cobra.MatchAll(ifn, fns...)` | 组合多个校验器 |

### 错误处理：优先用 RunE

`RunE` 返回 error 时，cobra 统一打印错误并让 `Execute()` 返回非零值，由 `main` 决定退出码。相比在 `Run` 里 `log.Fatal` 或 `os.Exit`，RunE 的好处是：错误可测试（见进阶参考）、行为一致、也便于上层统一处理。只有确实需要立即终止进程时才用 `log.Fatalf`（本项目 `root.go` 中 `os.Chdir` 失败即属此类）。

沉默使用提示的开关：若错误信息已足够清晰、不想每次打印整段 usage，设 `rootCmd.SilenceUsage = true`；连错误本身也想自行打印时再加 `SilenceErrors = true`（配合 `Execute()` 中自行处理 err）。

### 持久 flag 与局部 flag 的取舍

- 局部 flag（`Flags()`）：只属于定义它的命令。默认选择——大多数 flag 只服务一个命令，缩小作用域可避免子命令间命名冲突
- 持久 flag（`PersistentFlags()`）：定义处及其全部子孙命令可见。典型场景是根命令上的 `--verbose`、`--config`、`--log-level` 这类全局开关

### 版本信息与构建信息注入

本项目通过 `build` 包（`build/build.go`，构建时用 ldflags 注入）提供 `BuildTime`、`BuildVersion`。在 cobra 中可进一步接入标准的 `--version`：

```go
// 在 cmd/root.go 的 init() 中
rootCmd.Version = build.BuildVersion
rootCmd.SetVersionTemplate("{{.Version}}\n")
```

设置了 `Version` 后 cobra 自动提供 `--version` flag。注意：`Version` 非空会与自定义的 `version` 子命令冲突，二选一。

### help 文本微调

- 隐藏命令：`serveCmd.Hidden = true` —— 命令仍可执行，但不进 help 列表（常用于内部调试命令）
- 分组展示：v1.6+ 支持命令分组，用 `AddGroup` + `GroupID` 把 `serve`、`migrate` 归入 "Server Commands" 之类的小节，命令多了之后 help 会清爽很多
- 用法行精简：`DisableFlagsInUseLine = true` 让 usage line 不罗列所有 flags

## 写新命令时的检查清单

1. `Use`、`Short` 已填；长说明放 `Long`
2. 位置参数约束用了 `Args` 校验器而非手写判断
3. 能返回错误的地方用 `RunE` 而非 `os.Exit`
4. flag 作用域选对（局部 vs 持久），flag 变量用 `XxxVarP` 绑定到包级变量
5. 新子命令在各自文件的 `init()` 里 `rootCmd.AddCommand` 注册
6. 写完跑一遍 `go run . <新命令> --help` 确认 help 输出合理

## 进阶主题

需要 shell 自动补全、运行时动态 flags、命令测试、PreRun 执行链、cobra-cli 详解时，读 `references/advanced.md`。

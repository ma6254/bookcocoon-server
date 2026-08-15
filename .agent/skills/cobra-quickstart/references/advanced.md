# Cobra 进阶参考

主指南未覆盖的进阶主题。按需查阅，不必通读。

## 目录

1. 命令测试
2. 执行链：PersistentPreRun / PreRun / PostRun
3. Shell 自动补全
4. 运行时动态 flags 与动态命令
5. 隐藏命令与命令分组
6. cobra-cli 详解
7. 常见陷阱

## 1. 命令测试

用 `SetArgs` + `Execute` 在测试里驱动整条命令，验证输出与退出行为：

```go
func TestServeCmd(t *testing.T) {
	cmd := serveCmd
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)   // 捕获输出
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--port", "9090"})  // 模拟命令行参数

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "listening on :9090") {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}
```

注意两点：`Execute()` 后要 `cmd.SetArgs(nil)` 重置参数，否则同包内其他测试会被上一次的参数污染；测试中 `RunE` 返回的 error 会被 `Execute()` 直接返回（当 `SilenceErrors` 为 true 时尤其直接），这正是用 `RunE` 而非 `os.Exit` 的价值——错误可断言。

## 2. 执行链：PersistentPreRun / PreRun / PostRun

cobra 在 `Run` 之前按「父先子后、持久先局部」的顺序执行钩子：

```
PersistentPreRunE（父）→ PersistentPreRunE（子）→ PreRunE → RunE → PostRunE → PersistentPostRunE
```

典型用途：

- 根命令的 `PersistentPreRunE`：初始化日志、校验全局配置、检查运行环境——所有子命令共享的准备工作放这里，而不是在每个子命令里复制
- `PostRunE`：记录指标、清理临时资源

钩子同样有 `*E` 变体（返回 error 会中断执行链），与 `RunE` 的理由一致，优先用 `*E` 版本。

## 3. Shell 自动补全

cobra 原生支持 bash / zsh / fish / PowerShell 补全。两个步骤：

1. 命令结构上启用：`rootCmd.CompletionOptions.DisableDefaultCmd = false`（默认即启用，勿主动关掉），并为位置参数补全提供候选：实现 `ValidArgsFunction`（动态候选，接收 `cmd`、`args`、`toComplete`）或填静态 `ValidArgs` 列表
2. 用户侧安装一次：

```bash
# bash
source <(myapp completion bash)
# PowerShell
myapp completion powershell | Out-String | Invoke-Expression
```

候选函数示例——按已完成参数动态提示：

```go
serveCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"dev", "prod"}, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
```

## 4. 运行时动态 flags 与动态命令

静态 `init()` 定义不了、依赖运行时信息（配置文件、环境探测）的场景：

- **动态 flag**：在 `PreRun` 里调 `cmd.Flags().AddFlag(...)`（持久 hook 更早、更稳）
- **动态子命令**：`rootCmd.AddCommand(buildDynamicCmd(...))` 后需 `rootCmd.InitDefaultHelpCmd()` 刷新 help 命令，否则新命令不显示

```go
// 例：按配置文件中的 profile 动态挂载命令
for _, profile := range cfg.Profiles {
	rootCmd.AddCommand(newProfileCmd(profile))
}
rootCmd.InitDefaultHelpCmd()
```

## 5. 隐藏命令与命令分组

- `cmd.Hidden = true`：可执行但不进 help 列表，适合内部调试命令（如 `debug-dump-config`）
- 分组（v1.6+）：help 里按小节组织命令，命令多了之后可读性显著提升

```go
rootCmd.AddGroup(&cobra.Group{ID: "server", Title: "Server Commands"})
serveCmd.GroupID = "server"
```

## 6. cobra-cli 详解

脚手架工具的命令行约定：

```bash
cobra-cli init [name]          # 生成 main.go + cmd/root.go（模块名缺省取 go.mod）
cobra-cli add serve -p rootCmd # 生成 cmd/serve.go 并挂到 rootCmd
```

- `add` 默认挂到 rootCmd；`-p parentCmd` 挂到其他父命令下
- 生成的子命令文件自带 `var serveCmd = &cobra.Command{...}` + `init()` 注册，与本项目布局一致
- 新版（v1.8+ 的 cobra-cli）不再生成 viper 样板；需要配置解析时自行添加 viper 依赖

## 7. 常见陷阱

- **flag 名冲突**：父子命令定义同名局部 flag 会在运行时 panic。需要共享开关时，用根命令的 `PersistentFlags()` 定义一处，子命令直接引用
- **`os.Exit` 死在 Run 里**：跳过所有 defer、无法测试。让错误沿 `RunE` → `Execute()` → `main` 的路径返回，退出码处理集中在入口
- **重复调用 `Execute()`**：一个进程只应调用一次。测试里复用命令对象时注意重置 `SetArgs(nil)`
- **`Version` 与自定义 version 子命令冲突**：二者只能留一个，`rootCmd.Version` 非空时 `--version` 自动生成
- **`Args` 校验器与 `Run` 里二次校验重复**：声明式校验覆盖不了的业务约束（如两个参数互斥）再进 `Run` 检查，简单数量/枚举约束交给 `Args` + `ValidArgs`
- **help 输出变乱**：多为 `Use` 字段写得随意所致（含空格、大小写不统一）。`Use` 的第一个词就是命令名，保持小写、无空格

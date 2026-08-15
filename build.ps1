<#
.SYNOPSIS
    BookCocoon Server 构建脚本
.DESCRIPTION
    生成 swagger 文档并编译 release/app.exe，同时把构建时间和版本信息注入 build 包。
.PARAMETER OutputDir
    输出目录，默认 release
.PARAMETER BinaryName
    输出二进制文件名，默认 app.exe
.PARAMETER Version
    手动指定版本号，默认通过 git describe 自动获取
.PARAMETER SkipSwag
    跳过 swagger 文档生成（docs 已生成时可用）
.EXAMPLE
    .\build.ps1
.EXAMPLE
    .\build.ps1 -Version v1.0.0 -SkipSwag
#>
[CmdletBinding()]
param(
    [string]$OutputDir = "release",
    [string]$BinaryName = "app.exe",
    [string]$Version = "",
    [switch]$SkipSwag
)

$ErrorActionPreference = "Stop"

# 项目根目录为脚本所在目录，先切换过去，避免从其他目录调用时路径错乱
$RootDir = $PSScriptRoot
$Package = "github.com/ma6254/bookcocoon-server"

Push-Location -LiteralPath $RootDir
try {
    # 检查 go 是否可用
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "未找到 go，请先安装 Go 并加入 PATH"
    }

    # 1. 生成 swagger 文档
    if (-not $SkipSwag) {
        if (-not (Get-Command swag -ErrorAction SilentlyContinue)) {
            throw "未找到 swag，请先执行: go install github.com/swaggo/swag/cmd/swag@latest"
        }
        Write-Host "==> swag init --parseDependency"
        swag init --parseDependency
        if ($LASTEXITCODE -ne 0) {
            throw "swag init 失败，退出码: $LASTEXITCODE"
        }
    }

    # 2. 收集构建信息
    # 构建时间使用 UTC，避免本地时区差异
    $BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

    if ($Version) {
        # 手动指定版本
        $BuildVersion = $Version
    } else {
        # 优先用最近的 tag 描述版本，无 tag 时回退为短提交哈希，工作区有未提交改动时附加 -dirty
        try {
            $BuildVersion = git describe --tags --always --dirty 2>$null
        } catch {
            $BuildVersion = $null
        }
        if (-not [string]::IsNullOrWhiteSpace($BuildVersion)) {
            $BuildVersion = $BuildVersion.Trim()
        } else {
            $BuildVersion = "unknown"
        }
    }

    # 3. 编译
    # -s -w 去除符号表与调试信息以减小体积
    # 注意：不用 -trimpath，mattn/go-sqlite3 的 cgo 合并源码在 -trimpath 下会找不到相对 include 的 .c 文件
    $OutputPath = Join-Path $OutputDir $BinaryName
    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

    $LdFlags = "-s -w " +
        "-X ${Package}/build.BuildTime=$BuildTime " +
        "-X ${Package}/build.BuildVersion=$BuildVersion"

    Write-Host "==> go build"
    Write-Host "    output : $OutputPath"
    Write-Host "    version: $BuildVersion"
    Write-Host "    time   : $BuildTime"

    go build -v -o $OutputPath -ldflags $LdFlags
    if ($LASTEXITCODE -ne 0) {
        throw "go build 失败，退出码: $LASTEXITCODE"
    }

    # 4. 首次构建时从 docs/default.yml 复制默认配置，不覆盖已有配置
    $ConfigPath = Join-Path $OutputDir "config.yml"
    if (-not (Test-Path -LiteralPath $ConfigPath)) {
        $DefaultConfig = Join-Path $RootDir "docs/default.yml"
        if (Test-Path -LiteralPath $DefaultConfig) {
            Copy-Item -LiteralPath $DefaultConfig -Destination $ConfigPath
            Write-Host "==> 已复制默认配置: $ConfigPath"
        }
    }

    # 5. 构建结果摘要
    $Size = (Get-Item -LiteralPath $OutputPath).Length
    Write-Host ("==> 构建完成: {0} ({1:N2} MB)" -f $OutputPath, ($Size / 1MB))
} finally {
    Pop-Location
}

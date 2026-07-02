#Requires -Version 5.1
<#
.SYNOPSIS
    WinDownloader233 — 游戏开发环境一键安装脚本
.DESCRIPTION
    从 GitHub 远程 tools.json 读取工具列表（本地 + 远程时间戳对比取最新），
    通过 winget 一键安装。支持分类选择、中国镜像加速。
.PARAMETER Category
    安装指定分类: programming, art, planning, audio, qa, pm, ai
    不指定则显示交互式菜单。
.PARAMETER All
    安装全部工具。
.PARAMETER List
    仅列出可用工具，不安装。
.PARAMETER Proxy
    设置 GitHub 代理，如 https://ghfast.top/
.EXAMPLE
    .\install.ps1                      # 交互式菜单
    .\install.ps1 -Category ai         # 仅安装 AI 环境
    .\install.ps1 -All                 # 安装全部
    .\install.ps1 -List                # 列出工具
#>

[CmdletBinding()]
param(
    [string]$Category,
    [switch]$All,
    [switch]$List,
    [string]$Proxy = ""
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# ── Colors ──────────────────────────────────────────────
function Write-Color($Text, $Color) {
    $prev = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $Color
    Write-Host $Text -NoNewline
    $Host.UI.RawUI.ForegroundColor = $prev
}
function Write-Header {
    Write-Host ""
    Write-Color "  ╔══════════════════════════════════════════╗`n" Cyan
    Write-Color "  ║" Cyan; Write-Color "  WinDownloader233 — Game Dev Toolkit  " White; Write-Color "║`n" Cyan
    Write-Color "  ╚══════════════════════════════════════════╝`n" Cyan
    Write-Host ""
}

# ── Registry URLs ───────────────────────────────────────
$RemoteURLs = @(
    "https://raw.githubusercontent.com/neko233-com/WinDownloader233/main/registry/data/tools.json"
)

# Apply proxy if set
if ($Proxy) {
    $RemoteURLs = $RemoteURLs | ForEach-Object { "$Proxy/$_" }
}

# ── Fetch registry ──────────────────────────────────────
function Get-Registry {
    # Try embedded local first
    $localPath = Join-Path (Join-Path (Join-Path $PSScriptRoot "registry") "data") "tools.json"
    $localReg = $null
    if (Test-Path $localPath) {
        try {
            $localReg = Get-Content $localPath -Raw -Encoding UTF8 | ConvertFrom-Json
        } catch {}
    }

    # Try remote
    $remoteReg = $null
    foreach ($url in $RemoteURLs) {
        try {
            $resp = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 15
            $remoteReg = $resp.Content | ConvertFrom-Json
            break
        } catch {
            Write-Verbose "Failed to fetch $url : $_"
        }
    }

    # Compare timestamps, use newest
    if ($localReg -and $remoteReg) {
        if ($remoteReg.timestamp -gt $localReg.timestamp) {
            Write-Color "  ✓ " Green; Write-Host "Remote registry updated, using latest"
            return $remoteReg
        }
        return $localReg
    }
    if ($remoteReg) {
        Write-Color "  ✓ " Green; Write-Host "Using remote registry"
        return $remoteReg
    }
    if ($localReg) {
        Write-Color "  ! " Yellow; Write-Host "Using local registry (remote unavailable)"
        return $localReg
    }
    Write-Color "  ✗ " Red; Write-Host "No registry found!"
    exit 1
}

# ── Category display names ──────────────────────────────
$CategoryNames = @{
    "programming" = "⌨️  程序开发 Programming"
    "art"         = "🎨 美术设计 Art & Design"
    "planning"    = "📋 策划文档 Planning"
    "audio"       = "🎵 音频制作 Audio"
    "qa"          = "🧪 测试 QA"
    "pm"          = "📊 项目管理 Project Mgmt"
    "ai"          = "🤖 AI 环境 AI Environment"
}

$CategoryOrder = @("programming","art","planning","audio","qa","pm","ai")

# ── Install via winget ──────────────────────────────────
function Install-Tool($Tool) {
    $name = $Tool.name.zh
    $wingetId = $Tool.wingetId

    if (-not $wingetId) {
        Write-Color "  ⏭ " Yellow; Write-Host "$name — skip (no winget ID, install manually: $($Tool.homepage))"
        return
    }

    # Check if already installed
    $check = winget list --id $wingetId --disable-interactivity 2>$null
    if ($check -and $check -notmatch "No installed package found") {
        Write-Color "  ✓ " Green; Write-Host "$name — already installed"
        return
    }

    Write-Color "  ↓ " Cyan; Write-Host "$name ($wingetId)..."
    try {
        winget install --id $wingetId `
            --accept-source-agreements `
            --accept-package-agreements `
            --silent `
            --disable-interactivity 2>&1 | Out-Null

        if ($LASTEXITCODE -eq 0) {
            Write-Color "  ✓ " Green; Write-Host "$name — installed"
        } else {
            Write-Color "  ✗ " Red; Write-Host "$name — winget exit code $LASTEXITCODE"
        }
    } catch {
        Write-Color "  ✗ " Red; Write-Host "$name — error: $_"
    }
}

# ── Interactive menu ────────────────────────────────────
function Show-Menu($Registry) {
    Write-Host "  Select category to install:"
    Write-Host ""

    $i = 1
    foreach ($cat in $CategoryOrder) {
        $count = ($Registry.tools | Where-Object { $_.category -eq $cat }).Count
        $label = $CategoryNames[$cat]
        Write-Color "  [$i] " Cyan; Write-Host "$label ($count tools)"
        $i++
    }
    Write-Color "  [A] " Cyan; Write-Host "Install ALL"
    Write-Color "  [L] " Cyan; Write-Host "List tools only"
    Write-Color "  [Q] " Cyan; Write-Host "Quit"
    Write-Host ""

    $choice = Read-Host "  Enter choice"
    switch -Wildcard ($choice.ToUpper()) {
        "Q" { exit 0 }
        "A" { return "all" }
        "L" { return "list" }
        default {
            $idx = [int]$choice - 1
            if ($idx -ge 0 -and $idx -lt $CategoryOrder.Count) {
                return $CategoryOrder[$idx]
            }
            Write-Color "  ✗ " Red; Write-Host "Invalid choice"
            exit 1
        }
    }
}

# ── List tools ──────────────────────────────────────────
function Show-Tools($Registry, $FilterCat) {
    $tools = $Registry.tools
    if ($FilterCat -and $FilterCat -ne "all") {
        $tools = $tools | Where-Object { $_.category -eq $FilterCat }
    }

    foreach ($cat in $CategoryOrder) {
        $catTools = $tools | Where-Object { $_.category -eq $cat }
        if ($catTools.Count -eq 0) { continue }

        Write-Host ""
        Write-Color "  $($CategoryNames[$cat])" Cyan
        Write-Host "  $('-' * 40)"

        foreach ($t in $catTools) {
            $freeTag = if ($t.isFree) { "[FREE]" } else { "[PRO]" }
            $freeColor = if ($t.isFree) { "Green" } else { "Yellow" }
            Write-Host "    " -NoNewline
            Write-Color "$freeTag " $freeColor
            Write-Host "$($t.name.zh) — $($t.description.zh) ($($t.size))"
        }
    }
}

# ── Main ────────────────────────────────────────────────
Write-Header

# Check winget
try {
    $null = winget --version 2>$null
} catch {
    Write-Color "  ✗ " Red; Write-Host "winget not found! Please install from https://aka.ms/getwinget"
    exit 1
}

Write-Color "  ⏳ " Cyan; Write-Host "Loading registry..."
$registry = Get-Registry
Write-Color "  ✓ " Green; Write-Host "Registry: $($registry.timestamp) ($($registry.tools.Count) tools)"
Write-Host ""

# Determine action
$target = $null
if ($All) {
    $target = "all"
} elseif ($List) {
    $target = "list"
} elseif ($Category) {
    $target = $Category
} else {
    $target = Show-Menu $registry
}

if ($target -eq "list") {
    Show-Tools $registry $null
    exit 0
}

# Get tools to install
$toolsToInstall = @()
if ($target -eq "all") {
    $toolsToInstall = $registry.tools
    Write-Host ""
    Write-Color "  ▶ " Cyan; Write-Host "Installing ALL $($toolsToInstall.Count) tools..."
} else {
    $toolsToInstall = $registry.tools | Where-Object { $_.category -eq $target }
    Write-Host ""
    Write-Color "  ▶ " Cyan; Write-Host "Installing $($CategoryNames[$target]) ($($toolsToInstall.Count) tools)..."
}

Write-Host ""

# Install
$installed = 0
$skipped = 0
$failed = 0

foreach ($tool in $toolsToInstall) {
    try {
        $wingetId = $tool.wingetId
        if (-not $wingetId) {
            $skipped++
            Write-Color "  ⏭ " Yellow; Write-Host "$($tool.name.zh) — no winget ID"
            continue
        }

        # Check installed
        $check = winget list --id $wingetId --disable-interactivity 2>$null
        if ($check -and $check -notmatch "No installed package found") {
            $skipped++
            Write-Color "  ✓ " Green; Write-Host "$($tool.name.zh) — already installed"
            continue
        }

        Write-Color "  ↓ " Cyan; Write-Host "$($tool.name.zh) ($wingetId)..."
        $result = winget install --id $wingetId `
            --accept-source-agreements `
            --accept-package-agreements `
            --silent `
            --disable-interactivity 2>&1

        if ($LASTEXITCODE -eq 0) {
            $installed++
            Write-Color "  ✓ " Green; Write-Host "$($tool.name.zh) — done"
        } else {
            $failed++
            Write-Color "  ✗ " Red; Write-Host "$($tool.name.zh) — failed (exit $LASTEXITCODE)"
        }
    } catch {
        $failed++
        Write-Color "  ✗ " Red; Write-Host "$($tool.name.zh) — error: $_"
    }
}

# Summary
Write-Host ""
Write-Color "  ══════════════════════════════════════════`n" Cyan
Write-Color "  ✓ " Green; Write-Host "Installed: $installed  "
Write-Color "⏭ " Yellow; Write-Host "Skipped: $skipped  "
Write-Color "✗ " Red; Write-Host "Failed: $failed"
Write-Color "  ══════════════════════════════════════════`n" Cyan
Write-Host ""

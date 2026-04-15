$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$backendRoot = Join-Path $projectRoot "backend"
$airConfig = Join-Path $backendRoot ".air.toml"

function Find-AirExecutable {
  $command = Get-Command air -ErrorAction SilentlyContinue
  if ($command) {
    return $command.Source
  }

  $goBin = (go env GOBIN).Trim()
  if ($goBin) {
    $candidate = Join-Path $goBin "air.exe"
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  $goPath = (go env GOPATH).Trim()
  if ($goPath) {
    $candidate = Join-Path (Join-Path $goPath "bin") "air.exe"
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  return $null
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "彩迹 后端开发模式" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "项目目录: $backendRoot" -ForegroundColor DarkGray
Write-Host ""

$airExecutable = Find-AirExecutable

if (-not $airExecutable) {
  Write-Host "未检测到 air，开始安装..." -ForegroundColor Yellow
  go install github.com/air-verse/air@latest

  if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "air 安装失败，请确认 Go 环境后重试。" -ForegroundColor Red
    exit 1
  }

  $airExecutable = Find-AirExecutable
}

if (-not $airExecutable) {
  Write-Host "未找到 air 可执行文件，请检查 GOPATH/GOBIN 配置。" -ForegroundColor Red
  exit 1
}

Set-Location $backendRoot

Write-Host "使用 air 热更新启动后端..." -ForegroundColor Green
Write-Host "配置文件: $airConfig" -ForegroundColor DarkGray
Write-Host ""

& $airExecutable -c $airConfig

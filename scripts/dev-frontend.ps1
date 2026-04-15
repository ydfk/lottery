$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$frontendRoot = Join-Path $projectRoot "frontend"

function Find-PnpmExecutable {
  $command = Get-Command pnpm -ErrorAction SilentlyContinue
  if ($command) {
    return $command.Source
  }

  $candidates = @(
    "C:\Program Files\Volta\pnpm.cmd",
    "C:\Program Files\Volta\pnpm.exe",
    (Join-Path $env:LOCALAPPDATA "pnpm\pnpm.cmd"),
    (Join-Path $env:LOCALAPPDATA "pnpm\pnpm.exe")
  )

  foreach ($candidate in $candidates) {
    if ($candidate -and (Test-Path $candidate)) {
      return $candidate
    }
  }

  return $null
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "彩迹 前端开发模式" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "项目目录: $frontendRoot" -ForegroundColor DarkGray
Write-Host ""

$pnpmExecutable = Find-PnpmExecutable

if (-not $pnpmExecutable) {
  Write-Host "未检测到 pnpm，请先安装 pnpm 或 Volta。" -ForegroundColor Red
  exit 1
}

Set-Location $frontendRoot

Write-Host "使用 pnpm dev 启动前端..." -ForegroundColor Green
Write-Host ""

if ($pnpmExecutable.ToLower().EndsWith(".cmd")) {
  cmd /c """$pnpmExecutable"" dev"
  exit $LASTEXITCODE
}

& $pnpmExecutable dev

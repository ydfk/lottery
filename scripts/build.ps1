$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$frontendRoot = Join-Path $projectRoot "frontend"
$backendRoot = Join-Path $projectRoot "backend"
$buildVersion = if ($env:BUILD_VERSION) { $env:BUILD_VERSION } elseif ($env:APP_VERSION) { $env:APP_VERSION } else { "dev-local" }

if (-not (Get-Command "pnpm" -ErrorAction SilentlyContinue)) {
  throw "缺少命令：pnpm"
}

$go = Get-Command "go" -ErrorAction SilentlyContinue
$mise = Get-Command "mise" -ErrorAction SilentlyContinue
if (-not $go -and -not $mise) {
  throw "缺少命令：go（也未检测到 mise）"
}

function Invoke-Go {
  param([string[]]$Arguments)
  if ($go) {
    & $go.Source @Arguments
  } else {
    & $mise.Source exec -- go @Arguments
  }
  if ($LASTEXITCODE -ne 0) { throw "Go 命令执行失败：$($Arguments -join ' ')" }
}

Write-Host "安装并构建前端" -ForegroundColor Cyan
Push-Location $frontendRoot
try {
  pnpm install --frozen-lockfile
  if ($LASTEXITCODE -ne 0) { throw "前端依赖安装失败" }
  $env:VITE_APP_VERSION = $buildVersion
  pnpm build
  if ($LASTEXITCODE -ne 0) { throw "前端构建失败" }
} finally {
  Pop-Location
}

Write-Host "构建后端" -ForegroundColor Cyan
$backendBin = Join-Path $backendRoot "bin"
New-Item -ItemType Directory -Path $backendBin -Force | Out-Null
Push-Location $backendRoot
try {
  Invoke-Go -Arguments @("mod", "download")
  $env:CGO_ENABLED = "0"
  Invoke-Go -Arguments @("build", "-o", "bin/lottery.exe", "./cmd")
} finally {
  Pop-Location
}

Write-Host "构建完成：frontend/dist，backend/bin/lottery.exe" -ForegroundColor Green

param(
  [ValidateSet("all", "backend", "frontend")]
  [string]$Service = "all"
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$backendRoot = Join-Path $projectRoot "backend"
$frontendRoot = Join-Path $projectRoot "frontend"
$runningProcesses = New-Object System.Collections.ArrayList

function Assert-Command {
  param([string]$Name)
  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "缺少命令：$Name"
  }
}

function Start-Backend {
  $air = Get-Command "air" -ErrorAction SilentlyContinue
  $go = Get-Command "go" -ErrorAction SilentlyContinue
  $mise = Get-Command "mise" -ErrorAction SilentlyContinue
  if ($air) {
    Write-Host "启动后端（Air 热更新）" -ForegroundColor Cyan
    $process = Start-Process -FilePath $air.Source -ArgumentList "-c", ".air.toml" -WorkingDirectory $backendRoot -NoNewWindow -PassThru
  } elseif ($go) {
    Write-Host "未检测到 Air，使用 go run 启动后端" -ForegroundColor Yellow
    $process = Start-Process -FilePath $go.Source -ArgumentList "run", "./cmd" -WorkingDirectory $backendRoot -NoNewWindow -PassThru
  } elseif ($mise) {
    Write-Host "未检测到 Air，使用 mise 管理的 Go 启动后端" -ForegroundColor Yellow
    $process = Start-Process -FilePath $mise.Source -ArgumentList "exec", "--", "go", "run", "./cmd" -WorkingDirectory $backendRoot -NoNewWindow -PassThru
  } else {
    throw "缺少命令：go（也未检测到 mise）"
  }
  [void]$runningProcesses.Add($process)
}

function Start-Frontend {
  Assert-Command "pnpm"
  if (-not (Test-Path (Join-Path $frontendRoot "node_modules"))) {
    Write-Host "安装前端依赖" -ForegroundColor Cyan
    Push-Location $frontendRoot
    try {
      pnpm install --frozen-lockfile
      if ($LASTEXITCODE -ne 0) { throw "前端依赖安装失败" }
    } finally {
      Pop-Location
    }
  }

  Write-Host "启动前端（Vite）" -ForegroundColor Cyan
  $process = Start-Process -FilePath $env:ComSpec -ArgumentList "/d", "/s", "/c", "pnpm dev" -WorkingDirectory $frontendRoot -NoNewWindow -PassThru
  [void]$runningProcesses.Add($process)
}

try {
  switch ($Service) {
    "all" { Start-Backend; Start-Frontend }
    "backend" { Start-Backend }
    "frontend" { Start-Frontend }
  }

  Write-Host "开发服务已启动，按 Ctrl+C 停止" -ForegroundColor Green
  $exitCode = 0
  while ($true) {
    foreach ($process in $runningProcesses) {
      if ($process.HasExited) {
        $exitCode = $process.ExitCode
        break
      }
    }
    if ($runningProcesses | Where-Object HasExited) { break }
    Start-Sleep -Seconds 1
  }
} finally {
  foreach ($process in $runningProcesses) {
    if (-not $process.HasExited) {
      taskkill.exe /PID $process.Id /T /F *> $null
    }
  }
}

exit $exitCode

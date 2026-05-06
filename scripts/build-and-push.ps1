$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
. "$scriptDir\build-version.ps1"

Set-Location $projectRoot

$buildVersion = Get-LotteryBuildVersion
$env:BUILD_VERSION = $buildVersion
$env:APP_VERSION = $buildVersion
$env:DOCKER_IMAGE_TAG = $buildVersion

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "彩迹 一键构建并推送 Docker 镜像" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "构建版本: $buildVersion" -ForegroundColor DarkGray
Write-Host ""

Write-Host "========================================" -ForegroundColor Yellow
Write-Host "步骤 1/2: 构建 Docker 镜像" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""

& "$scriptDir\docker-build.ps1"

if ($LASTEXITCODE -ne 0) {
  Write-Host ""
  Write-Host "构建失败，停止推送" -ForegroundColor Red
  exit 1
}

Write-Host ""
Write-Host "✓ 镜像构建成功" -ForegroundColor Green
Write-Host ""

Start-Sleep -Seconds 2

Write-Host "========================================" -ForegroundColor Yellow
Write-Host "步骤 2/2: 推送到镜像仓库" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""

& "$scriptDir\docker-push.ps1"

if ($LASTEXITCODE -ne 0) {
  Write-Host ""
  Write-Host "推送失败" -ForegroundColor Red
  exit 1
}

$imageName = if ($env:DOCKER_IMAGE_NAME) { $env:DOCKER_IMAGE_NAME } else { "ydfk/lottery" }
$imageTag = $buildVersion

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "✓ 全部完成" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "镜像已推送到:" -ForegroundColor Cyan
Write-Host "  $imageName`:$imageTag" -ForegroundColor White
if ($imageTag -ne "latest") {
  Write-Host "  $imageName`:latest" -ForegroundColor White
}

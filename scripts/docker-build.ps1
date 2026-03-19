$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

$imageName = if ($env:DOCKER_IMAGE_NAME) { $env:DOCKER_IMAGE_NAME } else { "ydfk/lottery" }
$imageTag = if ($env:DOCKER_IMAGE_TAG) { $env:DOCKER_IMAGE_TAG } else { "latest" }
$commitTag = git -C $projectRoot rev-parse --short HEAD 2>$null

if (-not $commitTag) {
  $commitTag = Get-Date -Format "yyyyMMddHHmmss"
}

Set-Location $projectRoot

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "彩迹 Docker 镜像构建" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "项目目录: $projectRoot" -ForegroundColor DarkGray
Write-Host "镜像名称: $imageName" -ForegroundColor DarkGray
Write-Host "主标签:   $imageTag" -ForegroundColor DarkGray
Write-Host "提交标签: $commitTag" -ForegroundColor DarkGray
Write-Host ""

$tags = @("$imageName`:$imageTag")
if ($commitTag -ne $imageTag) {
  $tags += "$imageName`:$commitTag"
}

$arguments = @("build", "-f", "Dockerfile")
foreach ($tag in $tags) {
  $arguments += @("-t", $tag)
}
$arguments += "."

Write-Host "开始构建镜像..." -ForegroundColor Yellow
docker @arguments

if ($LASTEXITCODE -ne 0) {
  Write-Host ""
  Write-Host "镜像构建失败" -ForegroundColor Red
  exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "✓ 镜像构建完成" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
foreach ($tag in $tags) {
  Write-Host "  $tag" -ForegroundColor White
}

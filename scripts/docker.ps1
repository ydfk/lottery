param(
  [ValidateSet("build", "up", "down", "logs", "push", "help")]
  [string]$Action = "help"
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$imageName = if ($env:DOCKER_IMAGE_NAME) { $env:DOCKER_IMAGE_NAME } else { "ydfk/lottery" }
$imageTag = if ($env:DOCKER_IMAGE_TAG) { $env:DOCKER_IMAGE_TAG } elseif ($env:BUILD_VERSION) { $env:BUILD_VERSION } else { "latest" }

if ($Action -eq "help") {
  Write-Host "用法：.\scripts\docker.ps1 <build|up|down|logs|push>"
  Write-Host "环境变量：DOCKER_IMAGE_NAME、DOCKER_IMAGE_TAG、BUILD_VERSION"
  return
}

if (-not (Get-Command "docker" -ErrorAction SilentlyContinue)) {
  throw "缺少命令：docker"
}

$imageTags = @("$imageName`:$imageTag")
if ($imageTag -ne "latest") {
  $imageTags += "$imageName`:latest"
}

Push-Location $projectRoot
try {
  switch ($Action) {
    "build" {
      $arguments = @("build", "-f", "Dockerfile", "--build-arg", "APP_VERSION=$imageTag")
      foreach ($image in $imageTags) { $arguments += @("-t", $image) }
      $arguments += "."
      docker @arguments
    }
    "up" { docker compose up -d --build }
    "down" { docker compose down }
    "logs" { docker compose logs -f --tail=100 }
    "push" {
      foreach ($image in $imageTags) {
        docker push $image
        if ($LASTEXITCODE -ne 0) { break }
      }
    }
  }

  if ($LASTEXITCODE -ne 0) {
    throw "Docker 操作失败：$Action"
  }
} finally {
  Pop-Location
}

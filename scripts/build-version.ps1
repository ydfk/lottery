$ErrorActionPreference = "Stop"

function Get-LotteryBuildVersion {
  if ($env:BUILD_VERSION) {
    return $env:BUILD_VERSION
  }
  if ($env:APP_VERSION) {
    return $env:APP_VERSION
  }
  if ($env:DOCKER_IMAGE_TAG -and $env:DOCKER_IMAGE_TAG -ne "latest") {
    return $env:DOCKER_IMAGE_TAG
  }
  return Get-Date -Format "yyyyMMddHHmm"
}

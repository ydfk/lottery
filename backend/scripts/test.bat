@echo off
setlocal
set ROOT=%~dp0..

pushd "%ROOT%"
go test ./...
set EXIT=%errorlevel%
popd

exit /b %EXIT%

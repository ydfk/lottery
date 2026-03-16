@echo off
setlocal
set ROOT=%~dp0..

pushd "%ROOT%"
go run ./cmd
set EXIT=%errorlevel%
popd

exit /b %EXIT%

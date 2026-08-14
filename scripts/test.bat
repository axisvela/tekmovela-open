@echo off
REM TEKMOVELA Open repository test script.
REM Runs contract fixture validation and Go SDK/example tests.
setlocal
cd /d "%~dp0.."

pushd contracts
call npm install --no-audit --no-fund >nul
call npm run test:contracts || exit /b 1
popd

go vet ./sdk/... ./examples/...
go test ./sdk/... ./examples/...
exit /b %ERRORLEVEL%

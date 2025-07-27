@echo off
REM PDF合并工具构建脚本 (Windows)

echo 开始构建PDF合并工具...

REM 检查Go是否已安装
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: Go语言未安装。请先安装Go语言环境。
    echo 访问 https://golang.org/dl/ 下载安装Go
    pause
    exit /b 1
)

REM 显示Go版本
echo Go版本信息:
go version

REM 下载依赖
echo 下载依赖包...
go mod tidy

REM 运行测试
echo 运行测试...
go test ./...

REM 构建应用程序
echo 构建应用程序...
go build -ldflags="-s -w" -o pdf-merger.exe ./cmd/pdfmerger

if %errorlevel% equ 0 (
    echo 构建完成！
    echo 可执行文件: pdf-merger.exe
    echo.
    echo 使用方法:
    echo   pdf-merger.exe
    echo.
    echo 注意: 首次运行时，Fyne可能需要下载额外的系统依赖。
) else (
    echo 构建失败！
    pause
    exit /b 1
)

pause
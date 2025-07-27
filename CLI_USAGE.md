# PDF合并工具 - 命令行版本使用指南

## 📖 概述

PDF合并工具命令行版本是一个跨平台的PDF文件合并工具，支持Windows、Linux、macOS等所有主流操作系统。与GUI版本相比，CLI版本具有以下优势：

- **跨平台兼容性更好** - 无需图形界面依赖
- **文件体积更小** - 仅2MB左右
- **适合自动化** - 可集成到脚本和工作流中
- **服务器友好** - 可在无图形界面的服务器上运行

## 🚀 快速开始

### 基本用法

```bash
# 合并两个PDF文件
./pdf-merger-cli-[platform] -input file1.pdf,file2.pdf -output merged.pdf

# 合并多个PDF文件
./pdf-merger-cli-[platform] -input doc1.pdf,doc2.pdf,doc3.pdf,doc4.pdf -output combined.pdf
```

### 平台特定示例

#### Windows
```cmd
# Windows 64位
pdf-merger-cli-windows-64bit.exe -input report1.pdf,report2.pdf -output final_report.pdf

# Windows 32位
pdf-merger-cli-windows-32bit.exe -input doc1.pdf,doc2.pdf -output merged.pdf
```

#### Linux
```bash
# Linux 64位
./pdf-merger-cli-linux-64bit -input chapter1.pdf,chapter2.pdf,chapter3.pdf -output book.pdf

# Linux ARM64 (如树莓派)
./pdf-merger-cli-linux-arm64 -input part1.pdf,part2.pdf -output complete.pdf
```

#### macOS
```bash
# macOS Intel
./pdf-merger-cli-macos-intel -input invoice1.pdf,invoice2.pdf -output invoices.pdf

# macOS Apple Silicon (M1/M2)
./pdf-merger-cli-macos-apple-silicon -input contract1.pdf,contract2.pdf -output contracts.pdf
```

## 📋 命令行选项

| 选项 | 必需 | 描述 | 示例 |
|------|------|------|------|
| `-input` | ✅ | 输入PDF文件路径，用逗号分隔 | `-input file1.pdf,file2.pdf` |
| `-output` | ❌ | 输出PDF文件路径 (默认: merged.pdf) | `-output result.pdf` |
| `-version` | ❌ | 显示版本信息 | `-version` |
| `-help` | ❌ | 显示帮助信息 | `-help` |

## 💡 使用技巧

### 1. 使用通配符 (Linux/macOS)
```bash
# 合并当前目录下所有PDF文件
./pdf-merger-cli-linux-64bit -input *.pdf -output all_documents.pdf

# 合并特定模式的文件
./pdf-merger-cli-macos-intel -input chapter_*.pdf -output complete_book.pdf
```

### 2. 使用绝对路径
```bash
# 使用完整路径
./pdf-merger-cli-windows-64bit.exe -input "C:\Documents\file1.pdf,C:\Documents\file2.pdf" -output "C:\Output\merged.pdf"
```

### 3. 处理包含空格的文件名
```bash
# Linux/macOS - 使用引号
./pdf-merger-cli-linux-64bit -input "My Document 1.pdf,My Document 2.pdf" -output "Final Document.pdf"

# Windows - 使用引号
pdf-merger-cli-windows-64bit.exe -input "Report 2023.pdf,Summary 2023.pdf" -output "Annual Report 2023.pdf"
```

## 🔧 集成到脚本

### Bash脚本示例 (Linux/macOS)
```bash
#!/bin/bash

# PDF合并脚本
INPUT_DIR="/path/to/input"
OUTPUT_DIR="/path/to/output"
CLI_TOOL="./pdf-merger-cli-linux-64bit"

# 检查输入目录
if [ ! -d "$INPUT_DIR" ]; then
    echo "错误: 输入目录不存在: $INPUT_DIR"
    exit 1
fi

# 获取所有PDF文件
PDF_FILES=$(find "$INPUT_DIR" -name "*.pdf" -type f | tr '\n' ',' | sed 's/,$//')

if [ -z "$PDF_FILES" ]; then
    echo "错误: 在 $INPUT_DIR 中未找到PDF文件"
    exit 1
fi

# 执行合并
echo "开始合并PDF文件..."
$CLI_TOOL -input "$PDF_FILES" -output "$OUTPUT_DIR/merged_$(date +%Y%m%d_%H%M%S).pdf"

if [ $? -eq 0 ]; then
    echo "✅ PDF合并完成"
else
    echo "❌ PDF合并失败"
    exit 1
fi
```

### PowerShell脚本示例 (Windows)
```powershell
# PDF合并脚本
$InputDir = "C:\Documents\PDFs"
$OutputDir = "C:\Documents\Output"
$CliTool = "pdf-merger-cli-windows-64bit.exe"

# 检查输入目录
if (!(Test-Path $InputDir)) {
    Write-Error "输入目录不存在: $InputDir"
    exit 1
}

# 获取所有PDF文件
$PdfFiles = Get-ChildItem -Path $InputDir -Filter "*.pdf" | ForEach-Object { $_.FullName }

if ($PdfFiles.Count -eq 0) {
    Write-Error "在 $InputDir 中未找到PDF文件"
    exit 1
}

# 创建输入参数
$InputParam = $PdfFiles -join ","
$OutputFile = "$OutputDir\merged_$(Get-Date -Format 'yyyyMMdd_HHmmss').pdf"

# 执行合并
Write-Host "开始合并PDF文件..."
& $CliTool -input $InputParam -output $OutputFile

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ PDF合并完成: $OutputFile"
} else {
    Write-Error "❌ PDF合并失败"
    exit 1
}
```

## 🐛 故障排除

### 常见问题

#### 1. 权限被拒绝 (Linux/macOS)
```bash
# 解决方案：添加执行权限
chmod +x pdf-merger-cli-*
```

#### 2. 文件未找到
```bash
# 检查文件路径是否正确
ls -la pdf-merger-cli-*

# 使用绝对路径
/full/path/to/pdf-merger-cli-linux-64bit -input file1.pdf,file2.pdf -output merged.pdf
```

#### 3. Windows安全警告
- 右键点击exe文件 → 属性 → 解除阻止
- 或者添加到Windows Defender排除列表

#### 4. 输入文件不存在
```bash
# 检查文件是否存在
ls -la file1.pdf file2.pdf

# 使用绝对路径
./pdf-merger-cli-linux-64bit -input /full/path/to/file1.pdf,/full/path/to/file2.pdf -output merged.pdf
```

### 调试技巧

#### 1. 检查版本信息
```bash
./pdf-merger-cli-linux-64bit -version
```

#### 2. 查看帮助信息
```bash
./pdf-merger-cli-linux-64bit -help
```

#### 3. 测试单个文件
```bash
# 先测试两个文件的合并
./pdf-merger-cli-linux-64bit -input file1.pdf,file2.pdf -output test.pdf
```

## 📊 性能说明

### 文件大小限制
- **单个文件**: 建议不超过100MB
- **总文件数**: 建议不超过50个文件
- **输出文件**: 根据输入文件总大小而定

### 内存使用
- **基础内存**: 约10MB
- **处理时内存**: 约为输入文件总大小的1.5倍
- **建议系统内存**: 至少512MB可用内存

### 处理速度
- **小文件** (< 1MB): 几乎瞬时完成
- **中等文件** (1-10MB): 1-5秒
- **大文件** (10-100MB): 5-30秒

## 🔗 相关链接

- [GUI版本使用指南](docs/USER_GUIDE.md)
- [技术开发指南](docs/TECHNICAL_GUIDE.md)
- [下载页面](DOWNLOAD.md)
- [项目主页](README.md)

---

**版本**: v1.0.0  
**最后更新**: 2025年7月27日  
**支持平台**: Windows, Linux, macOS

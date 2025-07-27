# PDF处理模块

这个模块提供了PDF合并工具所需的PDF处理功能，包括PDF读取、验证、解密和合并。

## 功能特性

### PDF服务 (PDFService)
- PDF文件格式验证
- PDF文件信息提取（页数、加密状态、文件大小、标题）
- PDF文件合并
- 加密状态检测

### PDF验证器 (PDFValidator)
- PDF文件格式验证
- PDF版本兼容性检查
- 文件完整性验证
- 加密状态检测
- 基本PDF信息提取

### PDF解密器 (PDFDecryptor)
- PDF文件解密
- 支持多密码尝试
- 自动解密流程

## 使用示例

### 基本PDF验证

```go
// 创建PDF服务
pdfService := pdf.NewPDFService()

// 验证PDF文件
err := pdfService.ValidatePDF("/path/to/file.pdf")
if err != nil {
    if pdfErr, ok := err.(*pdf.PDFError); ok {
        fmt.Printf("PDF错误类型: %v\n", pdfErr.Type)
        fmt.Printf("错误消息: %s\n", pdfErr.Message)
    }
}

// 获取PDF信息
info, err := pdfService.GetPDFInfo("/path/to/file.pdf")
if err != nil {
    log.Printf("获取PDF信息失败: %v", err)
} else {
    fmt.Printf("页数: %d\n", info.PageCount)
    fmt.Printf("文件大小: %d 字节\n", info.FileSize)
    fmt.Printf("标题: %s\n", info.Title)
    fmt.Printf("是否加密: %v\n", info.IsEncrypted)
}
```

### PDF解密

```go
// 创建PDF解密器
decryptor := pdf.NewPDFDecryptor("/tmp/pdf-merger")

// 检查是否加密
isEncrypted, err := decryptor.IsPDFEncrypted("/path/to/file.pdf")
if err != nil {
    log.Printf("检查加密状态失败: %v", err)
}

if isEncrypted {
    // 尝试使用指定密码解密
    decryptedPath, err := decryptor.DecryptPDF("/path/to/file.pdf", "password")
    if err != nil {
        log.Printf("解密失败: %v", err)
    } else {
        fmt.Printf("解密成功，解密后的文件: %s\n", decryptedPath)
    }

    // 尝试使用多个密码解密
    passwords := []string{"", "password", "123456", "admin"}
    decryptedPath, password, err := decryptor.TryDecryptPDF("/path/to/file.pdf", passwords)
    if err != nil {
        log.Printf("解密失败: %v", err)
    } else {
        fmt.Printf("解密成功，使用密码: %s\n", password)
        fmt.Printf("解密后的文件: %s\n", decryptedPath)
    }
}
```

### PDF合并

```go
// 创建PDF服务
pdfService := pdf.NewPDFService()

// 合并PDF文件
mainFile := "/path/to/main.pdf"
additionalFiles := []string{"/path/to/file1.pdf", "/path/to/file2.pdf"}
outputPath := "/path/to/output.pdf"

// 创建进度写入器（可选）
var progressBuffer bytes.Buffer

// 执行合并
err := pdfService.MergePDFs(mainFile, additionalFiles, outputPath, &progressBuffer)
if err != nil {
    log.Printf("合并失败: %v", err)
} else {
    fmt.Println("合并成功")
    fmt.Printf("进度输出: %s\n", progressBuffer.String())
}
```

## 错误处理

模块定义了以下错误类型：

- `ErrorInvalidFile`: 文件格式无效或已损坏
- `ErrorEncrypted`: 文件已加密
- `ErrorCorrupted`: 文件已损坏
- `ErrorPermission`: 没有访问文件的权限
- `ErrorMemory`: 内存不足
- `ErrorIO`: 文件读写错误

每个错误都包含用户友好的中文错误消息。

## 性能考虑

- 使用流式处理，避免将整个PDF文件加载到内存
- 使用unidoc库的高效PDF解析和处理功能
- 解密后的文件保存在临时目录，避免修改原始文件
- 合并过程中逐页处理，避免内存溢出
- 支持进度跟踪，便于用户了解处理进度
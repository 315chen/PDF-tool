# 文件管理模块

这个模块提供了PDF合并工具所需的文件管理和验证功能。

## 功能特性

### 文件管理 (FileManager)
- 文件存在性和可访问性验证
- 临时文件创建和自动清理
- 目录创建和管理
- 文件信息获取

### PDF验证 (PDFValidator)
- PDF文件格式验证
- PDF版本兼容性检查
- 文件完整性验证
- 加密状态检测
- 基本PDF信息提取

### 集成验证 (FileValidator)
- 一站式文件验证服务
- 综合文件和PDF信息获取
- 批量文件验证支持

## 使用示例

### 基本文件验证

```go
// 创建文件管理器
tempDir := "/tmp/pdf-merger"
fileManager := file.NewFileManager(tempDir)

// 验证文件
err := fileManager.ValidateFile("/path/to/file.pdf")
if err != nil {
    log.Printf("文件验证失败: %v", err)
}

// 获取文件信息
fileInfo, err := fileManager.GetFileInfo("/path/to/file.pdf")
if err != nil {
    log.Printf("获取文件信息失败: %v", err)
} else {
    fmt.Printf("文件大小: %d 字节\n", fileInfo.Size)
}
```

### PDF格式验证

```go
// 创建PDF验证器
validator := pdf.NewPDFValidator()

// 验证PDF格式
err := validator.ValidatePDFFile("/path/to/file.pdf")
if err != nil {
    if pdfErr, ok := err.(*pdf.PDFError); ok {
        fmt.Printf("PDF错误类型: %v\n", pdfErr.Type)
        fmt.Printf("错误消息: %s\n", pdfErr.Message)
    }
}

// 获取PDF信息
pdfInfo, err := validator.GetBasicPDFInfo("/path/to/file.pdf")
if err != nil {
    log.Printf("获取PDF信息失败: %v", err)
} else {
    fmt.Printf("是否加密: %v\n", pdfInfo.IsEncrypted)
    fmt.Printf("文件大小: %d 字节\n", pdfInfo.FileSize)
}
```

### 集成验证

```go
// 创建集成验证器
validator := file.NewFileValidator("/tmp/pdf-merger")

// 验证并获取完整信息
result, err := validator.ValidateAndGetInfo("/path/to/file.pdf")
if err != nil {
    log.Printf("验证失败: %v", err)
} else if result.IsValid {
    fmt.Println("文件验证通过")
    fmt.Printf("文件大小: %d 字节\n", result.FileInfo.Size)
    fmt.Printf("是否加密: %v\n", result.PDFInfo.IsEncrypted)
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

## 测试

运行单元测试：

```bash
go test ./pkg/file -v
go test ./pkg/pdf -v
```

运行示例程序：

```bash
go run examples/file_validation_example.go
```

## 性能特性

- 轻量级文件验证，不加载整个文件到内存
- 自动临时文件清理，防止磁盘空间泄漏
- 线程安全的临时文件管理
- 高效的PDF格式检测算法
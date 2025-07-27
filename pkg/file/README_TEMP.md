# 临时文件管理

这个模块提供了PDF合并工具所需的临时文件管理功能，确保临时资源在使用后被正确清理。

## 功能特性

### 临时文件管理器 (TempFileManager)
- 创建和管理临时文件
- 自动清理过期的临时文件
- 会话隔离，确保不同会话的临时文件互不干扰
- 支持带前缀和后缀的临时文件创建
- 支持带内容的临时文件创建
- 支持文件复制到临时文件

### 资源管理器 (ResourceManager)
- 管理需要清理的资源（文件、目录、自定义资源）
- 按优先级清理资源
- 支持单独清理特定资源
- 批量清理所有资源

### 自动资源清理器 (AutoCleaner)
- 监听系统信号，在程序异常退出时自动清理资源
- 全局默认实例，方便在整个应用程序中使用
- 支持添加自定义清理操作

## 使用示例

### 基本临时文件操作

```go
// 创建文件管理器
tempDir := filepath.Join(os.TempDir(), "pdf-merger")
fileManager := file.NewFileManager(tempDir)

// 创建临时文件
tempFile, err := fileManager.CreateTempFile()
if err != nil {
    log.Fatalf("创建临时文件失败: %v", err)
}

// 创建带前缀的临时文件
prefixFile, fileObj, err := fileManager.CreateTempFileWithPrefix("prefix_", ".pdf")
if err != nil {
    log.Fatalf("创建带前缀的临时文件失败: %v", err)
}
defer fileObj.Close()

// 创建带内容的临时文件
content := []byte("文件内容")
contentFile, err := fileManager.CreateTempFileWithContent("content_", ".txt", content)
if err != nil {
    log.Fatalf("创建带内容的临时文件失败: %v", err)
}

// 复制文件到临时文件
copyFile, err := fileManager.CopyToTempFile("/path/to/source.pdf", "copy_")
if err != nil {
    log.Fatalf("复制到临时文件失败: %v", err)
}

// 清理所有临时文件
if err := fileManager.CleanupTempFiles(); err != nil {
    log.Printf("清理临时文件失败: %v", err)
}
```

### 资源管理

```go
// 创建资源管理器
resourceManager := file.NewResourceManager()

// 添加文件资源
resourceManager.AddFile("/path/to/file.txt", 1)

// 添加目录资源
resourceManager.AddDirectory("/path/to/dir", 2)

// 添加自定义资源
resourceManager.AddCustom(func() error {
    fmt.Println("执行自定义清理操作")
    return nil
}, 3)

// 清理特定资源
if err := resourceManager.CleanupResource("/path/to/file.txt"); err != nil {
    log.Printf("清理资源失败: %v", err)
}

// 清理所有资源
if errors := resourceManager.Cleanup(); len(errors) > 0 {
    for _, err := range errors {
        log.Printf("清理资源时发生错误: %v", err)
    }
}
```

### 自动资源清理

```go
// 设置自动资源清理
file.SetupDefaultAutoCleaner()

// 添加文件到自动清理
file.AddFileToAutoClean("/path/to/file.txt", 1)

// 添加目录到自动清理
file.AddDirectoryToAutoClean("/path/to/dir", 2)

// 添加自定义清理操作
file.AddCustomToAutoClean(func() error {
    fmt.Println("执行自定义清理操作")
    return nil
}, 3)

// 手动触发清理
if errors := file.CleanupAll(); len(errors) > 0 {
    for _, err := range errors {
        log.Printf("清理资源时发生错误: %v", err)
    }
}
```

## 最佳实践

1. **使用自动资源清理器**：在应用程序入口点设置自动资源清理器，确保即使程序异常退出也能清理资源。

2. **设置合理的优先级**：清理资源时，设置合理的优先级，确保资源按正确的顺序清理。例如，先清理文件，再清理目录。

3. **使用会话隔离**：每次应用程序运行时使用不同的会话目录，避免不同会话之间的临时文件冲突。

4. **定期清理过期文件**：设置合理的过期时间，定期清理过期的临时文件，避免磁盘空间浪费。

5. **错误处理**：妥善处理清理过程中的错误，避免因清理失败而导致资源泄漏。

## 性能考虑

- 临时文件管理器使用延迟清理机制，避免频繁的文件系统操作。
- 资源管理器按优先级清理资源，确保资源按正确的顺序清理，避免依赖问题。
- 自动资源清理器使用信号处理机制，在程序异常退出时自动清理资源，避免资源泄漏。
- 临时文件使用系统临时目录，确保在不同操作系统上都能正常工作。
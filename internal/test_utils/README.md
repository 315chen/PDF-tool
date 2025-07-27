# 测试工具包 (Test Utils)

这个包提供了一套完整的测试工具，包括模拟对象、测试数据工厂和测试助手，用于支持PDF合并工具的测试。

## 主要组件

### 1. 模拟对象 (Mock Objects)

#### MockPDFService
模拟PDF服务，支持设置预期的合并结果、验证结果和文件信息。

```go
// 创建模拟PDF服务
service := test_utils.NewMockPDFService()

// 设置合并结果
service.SetMergeResult("/output/test.pdf", nil) // 成功
service.SetMergeResult("/output/error.pdf", fmt.Errorf("合并失败"))

// 设置处理延迟
service.SetMergeDelay(100 * time.Millisecond)

// 使用服务
ctx := context.Background()
err := service.Merge(ctx, "/main.pdf", []string{"/file1.pdf"}, "/output/test.pdf", nil)

// 验证调用次数
count := service.GetCallCount("Merge")
```

#### MockFileManager
模拟文件管理器，支持虚拟文件系统操作。

```go
// 创建模拟文件管理器
manager := test_utils.NewMockFileManager()

// 添加虚拟文件
manager.AddFile("/test/file.pdf", []byte("PDF content"))

// 验证文件
entry, err := manager.ValidateFile("/test/file.pdf")

// 创建临时文件
tempFile, err := manager.CreateTempFile("prefix")
```

#### MockProgressCallback
模拟进度回调，记录所有进度更新。

```go
// 创建进度回调
callback := test_utils.NewMockProgressCallback()

// 模拟进度更新
callback.OnProgress(50.0, "processing", "合并中...")

// 获取更新记录
updates := callback.GetUpdates()
statuses := callback.GetStatuses()
```

### 2. 测试数据工厂 (Test Data Factory)

#### TestDataFactory
用于创建各种测试数据的工厂类。

```go
// 创建工厂
factory := test_utils.NewTestDataFactory()

// 创建有效文件条目
entry := factory.CreateValidFileEntry("/test/file.pdf")

// 创建加密文件条目
encryptedEntry := factory.CreateEncryptedFileEntry("/test/encrypted.pdf")

// 创建无效文件条目
invalidEntry := factory.CreateInvalidFileEntry("/test/invalid.pdf", "文件损坏")

// 创建大文件条目
largeEntry := factory.CreateLargeFileEntry("/test/large.pdf", 100) // 100MB

// 创建混合文件列表
entries := factory.CreateMixedFileEntryList(3, 1, 1) // 3个有效，1个无效，1个加密

// 创建合并任务
job := factory.CreatePendingMergeJob("/main.pdf", []string{"/file1.pdf"}, "/output.pdf")
runningJob := factory.CreateRunningMergeJob("/main.pdf", []string{"/file1.pdf"}, "/output.pdf", 50.0)
completedJob := factory.CreateCompletedMergeJob("/main.pdf", []string{"/file1.pdf"}, "/output.pdf")
failedJob := factory.CreateFailedMergeJob("/main.pdf", []string{"/file1.pdf"}, "/output.pdf", "合并失败")
```

#### TestScenarioBuilder
用于构建测试场景的构建器。

```go
// 创建场景构建器
builder := test_utils.NewTestScenarioBuilder()

// 构建正常合并场景
normalScenario := builder.BuildNormalMergeScenario()

// 构建错误场景
errorScenario := builder.BuildErrorScenario("文件不存在")

// 构建性能测试场景
perfScenario := builder.BuildPerformanceScenario(10, 100*time.Millisecond)

// 构建并发测试场景
concurrencyScenario := builder.BuildConcurrencyScenario(5)
```

### 3. 测试助手 (Test Helpers)

#### TestHelper
提供各种测试辅助功能。

```go
// 创建测试助手
helper := test_utils.NewTestHelper(t)
defer helper.Cleanup()

// 创建测试文件
filePath := helper.CreateTestFile("test.txt", []byte("content"))
pdfPath := helper.CreateTestPDF("test.pdf")

// 创建测试目录
dirPath := helper.CreateTestDirectory("testdir")

// 断言
helper.AssertNoError(err)
helper.AssertEqual(expected, actual)
helper.AssertTrue(condition)
helper.AssertFileExists(filePath)

// 等待条件
helper.WaitForCondition(func() bool {
    return someCondition()
}, time.Second, "condition description")

// 等待任务完成
helper.WaitForJobCompletion(job, 10*time.Second)
```

#### TestRunner
用于运行测试场景的运行器。

```go
// 创建测试运行器
runner := test_utils.NewTestRunner(t)
defer runner.Cleanup()

// 添加场景
scenario := test_utils.TestScenario{
    Name: "测试场景",
    Description: "测试描述",
    Setup: func() interface{} {
        return "test data"
    },
    Execute: func(data interface{}) error {
        // 执行测试逻辑
        return nil
    },
    Verify: func(data interface{}, err error) bool {
        return err == nil
    },
}

runner.AddScenario(scenario)
runner.RunScenarios()
```

#### PerformanceProfiler
用于性能分析的工具。

```go
// 创建性能分析器
profiler := test_utils.NewPerformanceProfiler()

// 开始分析
profiler.Start()

// 执行被测试的代码
// ...

// 停止分析
profiler.Stop()

// 获取结果
duration := profiler.GetDuration()
profiler.SetMetric("operations", 1000)
profiler.SetMetric("memory_usage", 1024*1024)

allMetrics := profiler.GetAllMetrics()
```

## 使用示例

### 单元测试示例

```go
func TestPDFMerger(t *testing.T) {
    // 创建模拟服务
    pdfService := test_utils.NewMockPDFService()
    fileManager := test_utils.NewMockFileManager()
    
    // 设置测试数据
    fileManager.AddFile("/main.pdf", []byte("main content"))
    fileManager.AddFile("/file1.pdf", []byte("file1 content"))
    
    // 设置预期结果
    pdfService.SetMergeResult("/output.pdf", nil)
    
    // 创建被测试对象
    merger := NewPDFMerger(pdfService, fileManager)
    
    // 执行测试
    err := merger.Merge("/main.pdf", []string{"/file1.pdf"}, "/output.pdf")
    
    // 验证结果
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    
    // 验证调用
    if count := pdfService.GetCallCount("Merge"); count != 1 {
        t.Errorf("Expected 1 call to Merge, got %d", count)
    }
}
```

### 集成测试示例

```go
func TestIntegrationScenario(t *testing.T) {
    // 创建测试助手
    helper := test_utils.NewTestHelper(t)
    defer helper.Cleanup()
    
    // 创建测试文件
    mainFile := helper.CreateTestPDF("main.pdf")
    additionalFile := helper.CreateTestPDF("additional.pdf")
    
    // 创建输出路径
    outputFile := filepath.Join(helper.GetTempDir(), "output.pdf")
    
    // 执行集成测试
    err := performMerge(mainFile, []string{additionalFile}, outputFile)
    
    // 验证结果
    helper.AssertNoError(err)
    helper.AssertFileExists(outputFile)
}
```

### 性能测试示例

```go
func BenchmarkMergePerformance(b *testing.B) {
    // 创建基准测试运行器
    runner := test_utils.NewBenchmarkRunner()
    
    // 添加基准测试
    benchmark := test_utils.BenchmarkData{
        Name:     "PDF合并性能测试",
        DataSize: 10,
        Setup: func(size int) interface{} {
            factory := test_utils.NewTestDataFactory()
            return factory.CreateFileEntryList(size)
        },
        Operation: func(data interface{}) error {
            entries := data.([]*model.FileEntry)
            // 执行合并操作
            return performMergeOperation(entries)
        },
    }
    
    runner.AddBenchmark(benchmark)
    runner.RunBenchmarks(b)
}
```

## 最佳实践

1. **使用模拟对象隔离依赖**：在单元测试中使用模拟对象来隔离外部依赖。

2. **使用测试数据工厂创建一致的测试数据**：避免在测试中硬编码测试数据。

3. **使用测试助手简化测试代码**：利用断言和等待功能简化测试逻辑。

4. **使用场景构建器组织复杂测试**：对于复杂的测试场景，使用场景构建器来组织测试。

5. **使用性能分析器监控性能**：在性能测试中使用性能分析器来收集指标。

6. **及时清理资源**：使用defer语句确保测试资源得到正确清理。

## 扩展

这个测试工具包设计为可扩展的。你可以：

1. 添加新的模拟对象来支持新的依赖
2. 扩展测试数据工厂来创建新类型的测试数据
3. 添加新的测试助手功能
4. 创建自定义的测试场景构建器

通过这些工具，你可以编写更可靠、更易维护的测试代码。

# PDF合并工具 - 技术指南

## 🚀 快速开始

### 环境要求
- Go 1.21 或更高版本
- 支持的操作系统: macOS, Windows, Linux
- PDFCPU 命令行工具 (可选，用于高级功能)

### 安装与构建

```bash
# 克隆项目
git clone <repository-url>
cd pdf-merger

# 安装依赖
go mod download

# 构建应用程序
go build -o pdf-merger ./cmd/pdfmerger

# 运行应用程序
./pdf-merger
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./internal/controller -v

# 运行性能测试
go test ./tests -bench=. -benchmem

# 运行并发安全测试
go test ./... -race
```

## 🏗️ 架构设计

### 设计原则

1. **单一职责原则**: 每个组件只负责一个特定功能
2. **依赖倒置原则**: 高层模块不依赖低层模块
3. **接口隔离原则**: 使用小而专一的接口
4. **开闭原则**: 对扩展开放，对修改关闭

### 核心接口

#### PDFService 接口
```go
type PDFService interface {
    Merge(ctx context.Context, mainFile string, additionalFiles []string, 
          outputPath string, progressCallback func(float64)) error
    ValidateFile(filePath string) error
    GetFileInfo(filePath string) (*FileInfo, error)
    Close() error
}
```

#### FileManager 接口
```go
type FileManager interface {
    ValidateFile(filePath string) error
    CreateTempFile(prefix string) (string, error)
    CleanupTempFiles() error
    GetFileInfo(filePath string) (*FileInfo, error)
}
```

### 数据流

```
用户操作 → UI层 → 控制器层 → 服务层 → PDF处理库
    ↓         ↓        ↓         ↓          ↓
  事件处理 → 状态管理 → 业务逻辑 → 文件操作 → 底层处理
```

## 🔧 核心组件详解

### 1. 控制器层 (Controller)

**职责**: 协调各个组件，处理业务逻辑

**关键方法**:
- `StartMergeJob()`: 启动合并任务
- `CancelCurrentJob()`: 取消当前任务
- `ValidateFile()`: 验证文件

**设计特点**:
- 异步操作处理
- 错误恢复机制
- 资源自动清理

### 2. PDF处理层 (PDFService)

**职责**: 处理PDF相关操作

**实现方式**:
- **PDFCPUAdapter**: 使用PDFCPU库
- **StreamingMerger**: 流式处理大文件
- **BatchProcessor**: 批量处理多文件

**性能优化**:
- 内存流式处理
- 并发文件验证
- 智能缓存策略

### 3. 文件管理层 (FileManager)

**职责**: 管理文件操作和临时文件

**功能特性**:
- 临时文件自动清理
- 文件验证和信息获取
- 资源使用监控

### 4. UI层 (User Interface)

**职责**: 用户交互界面

**组件结构**:
- **FileListManager**: 文件列表管理
- **ProgressManager**: 进度显示
- **主界面**: 整体布局和事件处理

## 🧪 测试策略

### 测试分层

1. **单元测试**: 测试单个函数和方法
2. **集成测试**: 测试组件间交互
3. **性能测试**: 测试性能指标
4. **并发测试**: 测试线程安全

### 测试工具

```go
// 使用testify进行断言
func TestController_StartMergeJob(t *testing.T) {
    assert := assert.New(t)
    require := require.New(t)
    
    controller := NewController(mockService, mockFileManager)
    err := controller.StartMergeJob(job)
    
    require.NoError(err)
    assert.True(controller.IsJobRunning())
}
```

### Mock对象

```go
type MockPDFService struct {
    mergeDelay time.Duration
    shouldFail bool
}

func (m *MockPDFService) Merge(ctx context.Context, mainFile string, 
    additionalFiles []string, outputPath string, 
    progressCallback func(float64)) error {
    // 模拟合并过程
    time.Sleep(m.mergeDelay)
    if m.shouldFail {
        return errors.New("模拟错误")
    }
    return nil
}
```

## 🚀 性能优化

### 内存管理

1. **流式处理**: 避免将整个文件加载到内存
2. **及时释放**: 主动释放不需要的资源
3. **内存监控**: 实时监控内存使用情况

```go
func (sm *StreamingMerger) enableProgressiveGC() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if sm.isMemoryHigh() {
                    runtime.GC()
                }
            case <-sm.stopGC:
                return
            }
        }
    }()
}
```

### 并发优化

1. **异步操作**: 避免阻塞UI线程
2. **工作池**: 限制并发数量
3. **取消机制**: 支持操作取消

```go
func (we *WorkflowExecutor) executeWithCancellation(
    ctx context.Context, workflow Workflow) error {
    
    done := make(chan error, 1)
    
    go func() {
        done <- workflow.Execute(ctx)
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## 🔒 错误处理

### 错误分类

1. **用户错误**: 文件不存在、格式错误等
2. **系统错误**: 内存不足、磁盘空间不足等
3. **网络错误**: 文件下载失败等

### 错误恢复

```go
func (c *Controller) handleMergeError(err error) {
    switch {
    case errors.Is(err, ErrFileNotFound):
        c.ui.ShowError("文件未找到，请检查文件路径")
    case errors.Is(err, ErrInsufficientMemory):
        c.ui.ShowError("内存不足，请关闭其他应用程序")
    default:
        c.ui.ShowError(fmt.Sprintf("合并失败: %v", err))
    }
    
    // 清理资源
    c.cleanup()
}
```

## 📊 监控与日志

### 性能监控

```go
type PerformanceMonitor struct {
    startTime    time.Time
    memoryUsage  int64
    fileCount    int
}

func (pm *PerformanceMonitor) RecordMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("处理时间: %v, 内存使用: %d MB, 文件数: %d",
        time.Since(pm.startTime),
        m.Alloc/1024/1024,
        pm.fileCount)
}
```

### 日志记录

```go
import "log/slog"

func (s *PDFService) Merge(ctx context.Context, files []string) error {
    logger := slog.With("operation", "merge", "fileCount", len(files))
    
    logger.Info("开始合并PDF文件")
    
    if err := s.validateFiles(files); err != nil {
        logger.Error("文件验证失败", "error", err)
        return err
    }
    
    logger.Info("PDF文件合并完成")
    return nil
}
```

## 🔧 配置管理

### 配置结构

```go
type Config struct {
    OutputDirectory   string   `json:"output_directory"`
    TempDirectory    string   `json:"temp_directory"`
    MaxFileSize      int64    `json:"max_file_size"`
    CommonPasswords  []string `json:"common_passwords"`
    EnableLogging    bool     `json:"enable_logging"`
}
```

### 配置加载

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return DefaultConfig(), nil // 使用默认配置
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }
    
    return &config, nil
}
```

## 🚀 部署指南

### 构建发布版本

```bash
# 构建所有平台
GOOS=windows GOARCH=amd64 go build -o pdf-merger.exe ./cmd/pdfmerger
GOOS=darwin GOARCH=amd64 go build -o pdf-merger-mac ./cmd/pdfmerger
GOOS=linux GOARCH=amd64 go build -o pdf-merger-linux ./cmd/pdfmerger
```

### 打包资源

```bash
# 创建发布包
mkdir release
cp pdf-merger release/
cp -r docs release/
cp README.md release/
tar -czf pdf-merger-v1.0.0.tar.gz release/
```

## 📚 扩展开发

### 添加新的PDF操作

1. 在`PDFService`接口中添加新方法
2. 在`PDFCPUAdapter`中实现具体逻辑
3. 在控制器中添加业务逻辑
4. 在UI中添加用户界面
5. 编写相应的测试

### 添加新的文件格式支持

1. 创建新的适配器实现`PDFService`接口
2. 在工厂方法中注册新适配器
3. 更新文件验证逻辑
4. 添加相应的测试用例

---

**技术支持**: 如有技术问题，请查看项目文档或提交Issue
**更新日期**: 2025年7月27日

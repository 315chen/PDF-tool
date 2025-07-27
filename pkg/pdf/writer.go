package pdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PDFWriter 提供增强的PDF写入功能，使用pdfcpu
type PDFWriter struct {
	outputPath        string
	tempPath          string
	isOpen            bool
	mutex             sync.Mutex
	retryCount        int
	maxRetries        int
	retryDelay        time.Duration
	initialRetryDelay time.Duration
	maxRetryDelay     time.Duration
	backoffFactor     float64
	backupEnabled     bool
	adapter           *PDFCPUAdapter
	config            *PDFCPUConfig
	content           []byte // 存储要写入的内容
}

// WriterOptions PDF写入器选项
type WriterOptions struct {
	MaxRetries        int           // 最大重试次数
	RetryDelay        time.Duration // 重试延迟（兼容旧用法）
	InitialRetryDelay time.Duration // 初始重试延迟
	MaxRetryDelay     time.Duration // 最大重试延迟
	BackoffFactor     float64       // 指数退避因子
	BackupEnabled     bool          // 是否启用备份
	TempDirectory     string        // 临时文件目录
	ValidationMode    string        // pdfcpu验证模式
	WriteObjectStream bool          // 是否写入对象流
	WriteXRefStream   bool          // 是否写入交叉引用流
	EncryptUsingAES   bool          // 是否使用AES加密
	EncryptKeyLength  int           // 加密密钥长度
}

// WriteResult 写入结果
type WriteResult struct {
	OutputPath     string
	TempPath       string
	BackupPath     string
	FileSize       int64
	WriteTime      time.Duration
	RetryCount     int
	Success        bool
	ValidationTime time.Duration
}

// NewPDFWriter 创建新的PDF写入器
func NewPDFWriter(outputPath string, options *WriterOptions) (*PDFWriter, error) {
	if options == nil {
		options = &WriterOptions{
			MaxRetries:        3,
			RetryDelay:        time.Second * 2,
			BackupEnabled:     true,
			TempDirectory:     os.TempDir(),
			ValidationMode:    "relaxed",
			WriteObjectStream: true,
			WriteXRefStream:   true,
			EncryptUsingAES:   true,
			EncryptKeyLength:  256,
		}
	}

	// 验证输出路径
	if err := validateOutputPath(outputPath); err != nil {
		return nil, err
	}

	// 生成临时文件路径
	tempPath := generateTempPath(outputPath, options.TempDirectory)

	// 创建pdfcpu配置
	config := &PDFCPUConfig{
		ValidationMode:    options.ValidationMode,
		WriteObjectStream: options.WriteObjectStream,
		WriteXRefStream:   options.WriteXRefStream,
		EncryptUsingAES:   options.EncryptUsingAES,
		EncryptKeyLength:  options.EncryptKeyLength,
		TempDirectory:     options.TempDirectory,
	}

	// 创建pdfcpu适配器
	adapter, err := NewPDFCPUAdapter(config)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorProcessing,
			Message: "无法创建pdfcpu适配器",
			File:    outputPath,
			Cause:   err,
		}
	}

	writer := &PDFWriter{
		outputPath:        outputPath,
		tempPath:          tempPath,
		isOpen:            false,
		maxRetries:        options.MaxRetries,
		retryDelay:        options.RetryDelay,
		initialRetryDelay: options.InitialRetryDelay,
		maxRetryDelay:     options.MaxRetryDelay,
		backoffFactor:     options.BackoffFactor,
		backupEnabled:     options.BackupEnabled,
		adapter:           adapter,
		config:            config,
		content:           make([]byte, 0),
	}

	return writer, nil
}

// Open 打开PDF写入器
func (w *PDFWriter) Open() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isOpen {
		return nil
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(w.outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法创建输出目录",
			File:    w.outputPath,
			Cause:   err,
		}
	}

	// 确保临时目录存在
	tempDir := filepath.Dir(w.tempPath)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法创建临时目录",
			File:    w.tempPath,
			Cause:   err,
		}
	}

	w.isOpen = true
	return nil
}

// Close 关闭PDF写入器
func (w *PDFWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isOpen {
		return nil
	}

	// 关闭pdfcpu适配器
	if w.adapter != nil {
		w.adapter.Close()
	}

	w.isOpen = false
	w.content = nil

	// 清理临时文件
	if w.tempPath != "" && fileExists(w.tempPath) {
		os.Remove(w.tempPath)
	}

	return nil
}

// AddContent 添加内容到PDF写入器
func (w *PDFWriter) AddContent(content []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isOpen {
		return &PDFError{
			Type:    ErrorIO,
			Message: "PDF写入器未打开",
			File:    w.outputPath,
		}
	}

	// 追加内容
	w.content = append(w.content, content...)
	return nil
}

// Write 写入PDF文件（支持上下文取消和指数退避）
func (w *PDFWriter) Write(ctx context.Context, progressWriter io.Writer) (*WriteResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF写入器未打开",
			File:    w.outputPath,
		}
	}

	startTime := time.Now()
	result := &WriteResult{
		OutputPath: w.outputPath,
		TempPath:   w.tempPath,
	}

	// 创建备份（如果启用）
	var backupPath string
	var rollbackMgr *RollbackManager
	if w.backupEnabled && fileExists(w.outputPath) {
		backupDir := filepath.Dir(w.outputPath)
		rollbackMgr = NewRollbackManager(backupDir)
		backupPath, _ = rollbackMgr.BackupFile(w.outputPath)
		result.BackupPath = backupPath
		if progressWriter != nil && backupPath != "" {
			fmt.Fprintf(progressWriter, "已创建备份文件: %s\n", backupPath)
		}
	}

	// 尝试写入文件（带重试机制，支持指数退避和取消）
	var writeErr error
	delay := w.initialRetryDelay
	if delay == 0 {
		delay = w.retryDelay
	}
	if delay == 0 {
		delay = time.Millisecond * 100
	}
	maxDelay := w.maxRetryDelay
	if maxDelay == 0 {
		maxDelay = time.Second * 5
	}
	factor := w.backoffFactor
	if factor <= 1.0 {
		factor = 2.0
	}

	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		w.retryCount = attempt

		if progressWriter != nil && attempt > 0 {
			fmt.Fprintf(progressWriter, "重试写入文件 (第 %d/%d 次, 延迟: %v)...\n", attempt, w.maxRetries, delay)
		}

		writeErr = w.attemptWrite(progressWriter)
		if writeErr == nil {
			break
		}

		// 只对可恢复错误重试
		pdfErr, ok := writeErr.(*PDFError)
		if !ok || (pdfErr.Type != ErrorIO && pdfErr.Type != ErrorProcessing) {
			break
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < w.maxRetries {
			if progressWriter != nil {
				fmt.Fprintf(progressWriter, "写入失败，%v 后重试: %v\n", delay, writeErr)
			}
			select {
			case <-ctx.Done():
				if progressWriter != nil {
					fmt.Fprintf(progressWriter, "写入操作被取消: %v\n", ctx.Err())
				}
				result.RetryCount = w.retryCount
				result.WriteTime = time.Since(startTime)
				result.Success = false
				if rollbackMgr != nil && backupPath != "" {
					_ = rollbackMgr.RestoreFile(backupPath, w.outputPath)
				} else if backupPath != "" {
					w.restoreBackup(backupPath)
				}
				return result, ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * factor)
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}

	result.RetryCount = w.retryCount
	result.WriteTime = time.Since(startTime)

	if writeErr != nil {
		result.Success = false
		// 恢复备份（如果存在）
		if rollbackMgr != nil && backupPath != "" {
			_ = rollbackMgr.RestoreFile(backupPath, w.outputPath)
		} else if backupPath != "" {
			w.restoreBackup(backupPath)
		}
		return result, writeErr
	}

	// 获取文件大小
	if fileInfo, err := os.Stat(w.outputPath); err == nil {
		result.FileSize = fileInfo.Size()
	}

	result.Success = true

	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "PDF文件写入成功: %s (大小: %.2f MB, 用时: %v)\n",
			w.outputPath, float64(result.FileSize)/(1024*1024), result.WriteTime)
	}

	return result, nil
}

// attemptWrite 尝试写入文件
func (w *PDFWriter) attemptWrite(progressWriter io.Writer) error {
	// 首先写入临时文件
	if err := writeToTempFile(w); err != nil {
		return err
	}

	// 验证临时文件
	if err := w.validateTempFile(); err != nil {
		os.Remove(w.tempPath)
		return err
	}

	// 原子性地移动临时文件到最终位置
	if err := w.atomicMove(); err != nil {
		os.Remove(w.tempPath)
		return err
	}

	return nil
}

// 包级可替换的写入临时文件函数
var writeToTempFile func(*PDFWriter) error = realWriteToTempFile

// realWriteToTempFile 写入内容到临时文件（原 writeToTempFile 实现）
func realWriteToTempFile(w *PDFWriter) error {
	// 创建临时文件
	tempFile, err := os.Create(w.tempPath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法创建临时文件",
			File:    w.tempPath,
			Cause:   err,
		}
	}
	defer tempFile.Close()

	if len(w.content) == 0 {
		// 如果没有内容，创建一个空的PDF
		return w.createBasicPDF(tempFile)
	}

	// 写入内容
	if _, err := tempFile.Write(w.content); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "写入临时文件失败",
			File:    w.tempPath,
			Cause:   err,
		}
	}

	return nil
}

// createBasicPDF 创建基本的PDF文件
func (w *PDFWriter) createBasicPDF(file *os.File) error {
	// 使用pdfcpu创建基本PDF
	if w.adapter != nil && w.adapter.cliAdapter != nil {
		// 使用pdfcpu create命令创建基本PDF
		if err := w.adapter.cliAdapter.CreateTestPDF(w.tempPath, 1); err != nil {
			return &PDFError{
				Type:    ErrorProcessing,
				Message: "无法创建基本PDF文件",
				File:    w.tempPath,
				Cause:   err,
			}
		}
	} else {
		// 回退到创建简单的PDF内容
		basicPDF := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
72 720 Td
(Generated PDF) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000173 00000 n 
0000000300 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
400
%%EOF`

		if _, err := file.Write([]byte(basicPDF)); err != nil {
			return &PDFError{
				Type:    ErrorIO,
				Message: "无法写入基本PDF内容",
				File:    w.tempPath,
				Cause:   err,
			}
		}
	}

	return nil
}

// validateTempFile 验证临时文件
func (w *PDFWriter) validateTempFile() error {
	// 检查文件是否存在
	if !fileExists(w.tempPath) {
		return &PDFError{
			Type:    ErrorIO,
			Message: "临时文件不存在",
			File:    w.tempPath,
		}
	}

	// 检查文件大小
	fileInfo, err := os.Stat(w.tempPath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法获取临时文件信息",
			File:    w.tempPath,
			Cause:   err,
		}
	}

	if fileInfo.Size() == 0 {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "临时文件为空",
			File:    w.tempPath,
		}
	}

	// 如果有内容且不是通过pdfcpu创建的，跳过pdfcpu验证
	if len(w.content) > 0 {
		// 对于直接写入的内容，只进行基本验证
		return w.basicPDFValidation(w.tempPath)
	}

	// 使用pdfcpu验证PDF格式
	if w.adapter != nil {
		if err := w.adapter.ValidateFile(w.tempPath); err != nil {
			return &PDFError{
				Type:    ErrorCorrupted,
				Message: "生成的PDF文件格式无效",
				File:    w.tempPath,
				Cause:   err,
			}
		}
	} else {
		// 回退到基本验证
		validator := NewPDFValidator()
		if err := validator.ValidatePDFFile(w.tempPath); err != nil {
			return &PDFError{
				Type:    ErrorCorrupted,
				Message: "生成的PDF文件格式无效",
				File:    w.tempPath,
				Cause:   err,
			}
		}
	}

	return nil
}

// basicPDFValidation 基本PDF验证
func (w *PDFWriter) basicPDFValidation(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法打开文件进行验证",
			File:    filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 检查PDF头部
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法读取文件头部",
			File:    filePath,
			Cause:   err,
		}
	}

	if !strings.HasPrefix(string(header), "%PDF-") {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "无效的PDF头部",
			File:    filePath,
		}
	}

	return nil
}

// atomicMove 原子性地移动文件
func (w *PDFWriter) atomicMove() error {
	// 在Windows上，如果目标文件存在，需要先删除
	if fileExists(w.outputPath) {
		if err := os.Remove(w.outputPath); err != nil {
			return &PDFError{
				Type:    ErrorIO,
				Message: "无法删除现有输出文件",
				File:    w.outputPath,
				Cause:   err,
			}
		}
	}

	// 移动临时文件到最终位置
	if err := os.Rename(w.tempPath, w.outputPath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法移动临时文件到最终位置",
			File:    w.outputPath,
			Cause:   err,
		}
	}

	return nil
}

// createBackup 创建备份文件
func (w *PDFWriter) createBackup() string {
	if !fileExists(w.outputPath) {
		return ""
	}

	backupPath := w.outputPath + ".backup." + time.Now().Format("20060102-150405")
	if err := copyFile(w.outputPath, backupPath); err != nil {
		// 备份失败不是致命错误，只记录
		fmt.Printf("Warning: 备份文件创建失败: %v\n", err)
		return ""
	}

	return backupPath
}

// restoreBackup 恢复备份文件
func (w *PDFWriter) restoreBackup(backupPath string) error {
	if backupPath == "" || !fileExists(backupPath) {
		return nil
	}

	// 删除当前输出文件（如果存在）
	if fileExists(w.outputPath) {
		os.Remove(w.outputPath)
	}

	// 恢复备份
	return copyFile(backupPath, w.outputPath)
}

// GetOutputPath 获取输出路径
func (w *PDFWriter) GetOutputPath() string {
	return w.outputPath
}

// GetTempPath 获取临时文件路径
func (w *PDFWriter) GetTempPath() string {
	return w.tempPath
}

// IsOpen 检查写入器是否打开
func (w *PDFWriter) IsOpen() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.isOpen
}

// GetAdapter 获取pdfcpu适配器
func (w *PDFWriter) GetAdapter() *PDFCPUAdapter {
	return w.adapter
}

// GetConfig 获取pdfcpu配置
func (w *PDFWriter) GetConfig() *PDFCPUConfig {
	return w.config
}

// validateOutputPath 验证输出路径
func validateOutputPath(outputPath string) error {
	if outputPath == "" {
		return &PDFError{
			Type:    ErrorInvalidInput,
			Message: "输出路径不能为空",
		}
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(outputPath), ".pdf") {
		return &PDFError{
			Type:    ErrorInvalidInput,
			Message: "输出文件必须是PDF格式",
			File:    outputPath,
		}
	}

	// 检查目录是否可写
	outputDir := filepath.Dir(outputPath)
	if err := checkDirectoryWritable(outputDir); err != nil {
		return &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写",
			File:    outputPath,
			Cause:   err,
		}
	}

	return nil
}

// checkDirectoryWritable 检查目录是否可写
func checkDirectoryWritable(dir string) error {
	// 尝试创建测试文件
	testFile := filepath.Join(dir, ".test_write")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}

// generateTempPath 生成临时文件路径
func generateTempPath(outputPath, tempDir string) string {
	baseName := filepath.Base(outputPath)
	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)

	return filepath.Join(tempDir, fmt.Sprintf("%s_temp_%d%s",
		name, time.Now().UnixNano(), ext))
}

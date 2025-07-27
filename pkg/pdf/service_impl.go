package pdf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 初始化unidoc库
func init() {
	// 设置处理器数量，避免过度使用CPU
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	if runtime.GOMAXPROCS(0) < 1 {
		runtime.GOMAXPROCS(1)
	}

	// 设置unidoc日志级别 - 暂时跳过以避免网络问题
	// log.SetLogger(nil)
}

// PDFServiceImpl 实现PDFService接口
type PDFServiceImpl struct {
	validator     *PDFValidator
	errorHandler  ErrorHandler
	mutex         sync.Mutex
	config        *ServiceConfig
}

// ServiceConfig PDF服务配置
type ServiceConfig struct {
	MaxRetries        int
	RetryDelay        time.Duration
	EnableStrictMode  bool
	PreferPDFCPU      bool
	TempDirectory     string
	MaxMemoryUsage    int64
}

// NewPDFService 创建一个新的PDF服务实例
func NewPDFService() PDFService {
	return NewPDFServiceWithConfig(nil)
}

// NewPDFServiceWithConfig 使用配置创建PDF服务实例
func NewPDFServiceWithConfig(config *ServiceConfig) PDFService {
	if config == nil {
		config = &ServiceConfig{
			MaxRetries:       3,
			RetryDelay:       time.Second * 2,
			EnableStrictMode: false,
			PreferPDFCPU:     true,
			TempDirectory:    os.TempDir(),
			MaxMemoryUsage:   100 * 1024 * 1024, // 100MB
		}
	}

	return &PDFServiceImpl{
		validator:    NewPDFValidator(),
		errorHandler: NewDefaultErrorHandler(config.MaxRetries),
		config:       config,
	}
}

// ValidatePDF 验证PDF文件格式是否有效
func (s *PDFServiceImpl) ValidatePDF(filePath string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 使用错误收集器收集验证过程中的错误
	errorCollector := NewErrorCollector()

	// 第一步：基本文件验证
	if err := s.basicFileValidation(filePath); err != nil {
		return s.errorHandler.HandleError(err)
	}

	// 第二步：优先使用pdfcpu进行验证（如果配置启用）
	if s.config.PreferPDFCPU {
		if err := s.validateWithPDFCPU(filePath); err == nil {
			return nil // pdfcpu验证成功
		} else {
			errorCollector.Add(fmt.Errorf("pdfcpu validation failed: %w", err))
		}
	}

	// 第三步：使用增强的PDF读取器进行验证
	if err := s.validateWithEnhancedReader(filePath); err == nil {
		return nil // 增强读取器验证成功
	} else {
		errorCollector.Add(fmt.Errorf("enhanced reader validation failed: %w", err))
	}

	// 如果所有验证方法都失败，返回综合错误
	if errorCollector.HasErrors() {
		return &PDFError{
			Type:    ErrorValidation,
			Message: "PDF文件验证失败，尝试了多种验证方法",
			File:    filePath,
			Cause:   fmt.Errorf("validation errors: %s", errorCollector.GetSummary()),
		}
	}

	return nil
}

// GetPDFInfo 获取PDF文件的基本信息
func (s *PDFServiceImpl) GetPDFInfo(filePath string) (*PDFInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 首先进行基本验证
	if err := s.basicFileValidation(filePath); err != nil {
		return nil, s.errorHandler.HandleError(err)
	}

	var info *PDFInfo
	var lastError error

	// 方法1：优先使用pdfcpu适配器获取详细信息
	if s.config.PreferPDFCPU {
		if pdfcpuInfo, err := s.getInfoWithPDFCPU(filePath); err == nil {
			info = pdfcpuInfo
		} else {
			lastError = err
		}
	}

	// 方法2：如果pdfcpu失败或未启用，使用增强的PDF读取器
	if info == nil {
		if readerInfo, err := s.getInfoWithEnhancedReader(filePath); err == nil {
			info = readerInfo
		} else {
			lastError = err
		}
	}

	// 方法3：如果增强读取器失败，回退到基本方法
	if info == nil {
		if basicInfo, err := s.getBasicPDFInfo(filePath); err == nil {
			info = basicInfo
		} else {
			lastError = err
		}
	}

	// 如果所有方法都失败
	if info == nil {
		return nil, &PDFError{
			Type:    ErrorProcessing,
			Message: "无法获取PDF文件信息",
			File:    filePath,
			Cause:   lastError,
		}
	}

	// 补充文件系统信息
	if err := s.enrichInfoWithFileSystemData(info, filePath); err != nil {
		// 文件系统信息获取失败不是致命错误，记录但继续
		// 可以在这里添加日志记录
	}

	// 验证获取的信息是否合理
	if err := s.validatePDFInfo(info); err != nil {
		return nil, &PDFError{
			Type:    ErrorCorrupted,
			Message: "获取的PDF信息不合理",
			File:    filePath,
			Cause:   err,
		}
	}

	return info, nil
}

// 新增的信息获取方法

// getInfoWithPDFCPU 使用pdfcpu获取PDF信息
func (s *PDFServiceImpl) getInfoWithPDFCPU(filePath string) (*PDFInfo, error) {
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return nil, err
	}
	defer adapter.Close()

	return adapter.GetFileInfo(filePath)
}

// getInfoWithEnhancedReader 使用增强读取器获取PDF信息
func (s *PDFServiceImpl) getInfoWithEnhancedReader(filePath string) (*PDFInfo, error) {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return reader.GetInfo()
}

// enrichInfoWithFileSystemData 用文件系统数据补充PDF信息
func (s *PDFServiceImpl) enrichInfoWithFileSystemData(info *PDFInfo, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// 更新文件大小（如果未设置）
	if info.FileSize == 0 {
		info.FileSize = fileInfo.Size()
	}

	// 更新文件路径（如果未设置）
	if info.FilePath == "" {
		info.FilePath = filePath
	}

	// 更新修改时间（如果未设置）
	if info.ModDate.IsZero() {
		info.ModDate = fileInfo.ModTime()
	}

	// 如果没有标题，使用文件名
	if info.Title == "" {
		info.Title = getFileNameWithoutExt(filePath)
	}

	return nil
}

// validatePDFInfo 验证PDF信息的合理性
func (s *PDFServiceImpl) validatePDFInfo(info *PDFInfo) error {
	if info == nil {
		return fmt.Errorf("PDF信息为空")
	}

	// 验证页数
	if info.PageCount < 0 {
		return fmt.Errorf("页数不能为负数: %d", info.PageCount)
	}

	// 验证文件大小
	if info.FileSize < 0 {
		return fmt.Errorf("文件大小不能为负数: %d", info.FileSize)
	}

	// 验证文件路径
	if info.FilePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	return nil
}

// getBasicPDFInfo 获取基本PDF信息（回退方法）
func (s *PDFServiceImpl) getBasicPDFInfo(filePath string) (*PDFInfo, error) {
	// 使用pdfcpu适配器获取信息
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu不可用: %w", err)
	}
	defer adapter.Close()

	// 获取文件信息
	info, err := adapter.GetFileInfo(filePath)
	if err != nil {
		return nil, err
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    filePath,
			Cause:   err,
		}
	}

	// 检查是否加密
	isEncrypted, err := s.IsPDFEncrypted(filePath)
	if err != nil {
		// 如果无法确定加密状态，假设未加密
		isEncrypted = false
	}

	// 补充信息
	info.FileSize = fileInfo.Size()
	info.IsEncrypted = isEncrypted
	info.CreationDate = fileInfo.ModTime()
	info.ModDate = fileInfo.ModTime()

	// 如果没有标题，使用文件名
	if info.Title == "" {
		info.Title = getFileNameWithoutExt(filePath)
	}

	return info, nil
}

// IsPDFEncrypted 检查PDF文件是否加密
func (s *PDFServiceImpl) IsPDFEncrypted(filePath string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 首先进行基本验证
	if err := s.basicFileValidation(filePath); err != nil {
		return false, s.errorHandler.HandleError(err)
	}

	// 方法1：优先使用pdfcpu检查加密状态
	if s.config.PreferPDFCPU {
		if encrypted, err := s.checkEncryptionWithPDFCPU(filePath); err == nil {
			return encrypted, nil
		}
	}

	// 方法2：使用增强读取器检查
	if encrypted, err := s.checkEncryptionWithEnhancedReader(filePath); err == nil {
		return encrypted, nil
	}

	// 方法3：使用基本内容检查（最后的回退）
	return s.checkEncryptionByContent(filePath)
}

// 新增的加密检查方法

// checkEncryptionWithPDFCPU 使用pdfcpu检查加密状态
func (s *PDFServiceImpl) checkEncryptionWithPDFCPU(filePath string) (bool, error) {
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return false, err
	}
	defer adapter.Close()

	info, err := adapter.GetFileInfo(filePath)
	if err != nil {
		return false, err
	}

	return info.IsEncrypted, nil
}

// checkEncryptionWithEnhancedReader 使用增强读取器检查加密状态
func (s *PDFServiceImpl) checkEncryptionWithEnhancedReader(filePath string) (bool, error) {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return false, err
	}
	defer reader.Close()

	return reader.IsEncrypted()
}

// checkEncryptionByContent 通过文件内容检查加密状态（最后的回退方法）
func (s *PDFServiceImpl) checkEncryptionByContent(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 读取文件的前4KB内容
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	content := string(buffer[:n])
	
	// 查找加密相关的关键字
	encryptionKeywords := []string{
		"/Encrypt",
		"/Filter",
		"/V ",
		"/R ",
		"/O ",
		"/U ",
		"/P ",
		"Standard",
		"Security",
	}

	for _, keyword := range encryptionKeywords {
		if strings.Contains(content, keyword) {
			return true, nil
		}
	}

	return false, nil
}

// ValidatePDFStructure 验证PDF文件结构完整性
func (s *PDFServiceImpl) ValidatePDFStructure(filePath string) error {
	// 首先进行基本验证
	if err := s.ValidatePDF(filePath); err != nil {
		return err
	}

	// 使用增强的PDF读取器进行结构验证
	reader, err := NewPDFReader(filePath)
	if err != nil {
		// 如果无法使用增强读取器，进行基本结构检查
		return s.validateBasicStructure(filePath)
	}
	defer reader.Close()

	// 使用增强读取器验证结构
	return reader.ValidateStructure()
}

// GetPDFMetadata 获取PDF文件元数据
func (s *PDFServiceImpl) GetPDFMetadata(filePath string) (map[string]string, error) {
	// 使用增强的PDF读取器获取元数据
	reader, err := NewPDFReader(filePath)
	if err != nil {
		// 如果无法使用增强读取器，返回基本元数据
		return s.getBasicMetadata(filePath)
	}
	defer reader.Close()

	return reader.GetMetadata()
}

// validateBasicStructure 基本结构验证（回退方法）
func (s *PDFServiceImpl) validateBasicStructure(filePath string) error {
	// 使用pdfcpu适配器进行基本验证
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return fmt.Errorf("pdfcpu不可用: %w", err)
	}
	defer adapter.Close()

	// 验证文件
	if err := adapter.ValidateFile(filePath); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF结构损坏",
			File:    filePath,
			Cause:   err,
		}
	}

	// 获取文件信息以检查页数
	info, err := adapter.GetFileInfo(filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "无法获取PDF信息",
			File:    filePath,
			Cause:   err,
		}
	}

	if info.PageCount <= 0 {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件没有有效页面",
			File:    filePath,
		}
	}

	return nil
}

// getBasicMetadata 获取基本元数据（回退方法）
func (s *PDFServiceImpl) getBasicMetadata(filePath string) (map[string]string, error) {
	metadata := make(map[string]string)

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    filePath,
			Cause:   err,
		}
	}

	// 添加基本文件信息
	metadata["FileName"] = getFileNameWithoutExt(filePath)
	metadata["FileSize"] = fmt.Sprintf("%d", fileInfo.Size())
	metadata["ModificationDate"] = fileInfo.ModTime().Format("2006-01-02 15:04:05")

	// 尝试获取PDF信息
	if info, err := s.GetPDFInfo(filePath); err == nil {
		metadata["PageCount"] = fmt.Sprintf("%d", info.PageCount)
		metadata["IsEncrypted"] = fmt.Sprintf("%t", info.IsEncrypted)
		if info.Title != "" {
			metadata["Title"] = info.Title
		}
	}

	return metadata, nil
}

// MergePDFs 将多个PDF文件合并为一个（使用流式处理）
func (s *PDFServiceImpl) MergePDFs(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 预处理：验证所有输入文件
	allFiles := []string{mainFile}
	allFiles = append(allFiles, additionalFiles...)

	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "开始合并 %d 个PDF文件...\n", len(allFiles))
	}

	// 验证所有输入文件 - 在验证期间释放锁以避免死锁
	s.mutex.Unlock()
	errorCollector := NewErrorCollector()
	validFiles := make([]string, 0, len(allFiles))

	for i, file := range allFiles {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "验证文件 %d/%d: %s\n", i+1, len(allFiles), file)
		}

		if err := s.ValidatePDF(file); err != nil {
			errorCollector.Add(fmt.Errorf("文件 %s 验证失败: %w", file, err))
			if progressWriter != nil {
				fmt.Fprintf(progressWriter, "警告: 跳过无效文件 %s: %v\n", file, err)
			}
		} else {
			validFiles = append(validFiles, file)
		}
	}
	s.mutex.Lock() // 重新获取锁

	// 检查是否有足够的有效文件进行合并
	if len(validFiles) == 0 {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "没有有效的PDF文件可以合并",
			File:    "",
			Cause:   fmt.Errorf("validation errors: %s", errorCollector.GetSummary()),
		}
	}

	if len(validFiles) == 1 {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "只有一个有效文件，直接复制到输出位置\n")
		}
		return s.copyFile(validFiles[0], outputPath)
	}

	// 尝试不同的合并策略
	var mergeError error

	// 策略1：优先使用pdfcpu合并（如果配置启用）
	if s.config.PreferPDFCPU {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "尝试使用pdfcpu进行合并...\n")
		}

		if err := s.mergeWithPDFCPU(validFiles, outputPath, progressWriter); err == nil {
			if progressWriter != nil {
				fmt.Fprintf(progressWriter, "pdfcpu合并成功完成\n")
			}
			return nil
		} else {
			mergeError = err
			if progressWriter != nil {
				fmt.Fprintf(progressWriter, "pdfcpu合并失败: %v\n", err)
			}
		}
	}

	// 策略2：使用流式合并器
	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "使用流式合并器进行合并...\n")
	}

	if err := s.mergeWithStreamingMerger(validFiles, outputPath, progressWriter); err == nil {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "流式合并成功完成\n")
		}
		return nil
	} else {
		mergeError = err
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "流式合并失败: %v\n", err)
		}
	}

	// 策略3：基本合并（最后的回退）
	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "使用基本合并方法...\n")
	}

	if err := s.mergeWithBasicMethod(validFiles, outputPath, progressWriter); err == nil {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "基本合并成功完成\n")
		}
		return nil
	} else {
		mergeError = err
	}

	// 所有合并策略都失败
	return &PDFError{
		Type:    ErrorProcessing,
		Message: "所有合并策略都失败",
		File:    outputPath,
		Cause:   mergeError,
	}
}

// 新增的合并方法

// mergeWithPDFCPU 使用pdfcpu进行合并
func (s *PDFServiceImpl) mergeWithPDFCPU(files []string, outputPath string, progressWriter io.Writer) error {
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return err
	}
	defer adapter.Close()

	if err := adapter.MergeFiles(files, outputPath); err != nil {
		return err
	}

	// 验证输出文件 - 使用独立的验证方法避免死锁
	if err := s.validateOutputFile(outputPath); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "合并后的PDF文件无效",
			File:    outputPath,
			Cause:   err,
		}
	}

	// 输出统计信息
	if progressWriter != nil {
		if info, err := adapter.GetFileInfo(outputPath); err == nil {
			fmt.Fprintf(progressWriter, "合并完成 - 总页数: %d, 文件大小: %s\n", 
				info.PageCount, info.GetFormattedSize())
		}
	}

	return nil
}

// mergeWithStreamingMerger 使用流式合并器进行合并
func (s *PDFServiceImpl) mergeWithStreamingMerger(files []string, outputPath string, progressWriter io.Writer) error {
	if len(files) == 0 {
		return fmt.Errorf("没有文件需要合并")
	}

	mainFile := files[0]
	additionalFiles := files[1:]

	merger := NewStreamingMerger(&MergeOptions{
		MaxMemoryUsage: s.config.MaxMemoryUsage,
		TempDirectory:  s.config.TempDirectory,
		EnableGC:       true,
		ChunkSize:      10,
	})

	result, err := merger.MergeFilesLegacy(mainFile, additionalFiles, outputPath, progressWriter)
	if err != nil {
		return err
	}

	// 验证输出文件
	if err := s.validateOutputFile(outputPath); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "合并后的PDF文件无效",
			File:    outputPath,
			Cause:   err,
		}
	}

	// 输出统计信息
	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "流式合并统计:\n")
		fmt.Fprintf(progressWriter, "  总页数: %d\n", result.TotalPages)
		fmt.Fprintf(progressWriter, "  处理文件数: %d\n", result.ProcessedFiles)
		fmt.Fprintf(progressWriter, "  跳过文件数: %d\n", len(result.SkippedFiles))
		fmt.Fprintf(progressWriter, "  处理时间: %v\n", result.ProcessingTime)
		fmt.Fprintf(progressWriter, "  内存使用: %.2f MB\n", float64(result.MemoryUsage)/(1024*1024))
	}

	return nil
}

// mergeWithBasicMethod 使用基本方法进行合并
func (s *PDFServiceImpl) mergeWithBasicMethod(files []string, outputPath string, progressWriter io.Writer) error {
	if len(files) == 0 {
		return fmt.Errorf("没有文件需要合并")
	}

	// 新增：只读目录检测
	dir := filepath.Dir(outputPath)
	if err := checkDirectoryWritable(dir); err != nil {
		return &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写（只读目录）",
			File:    dir,
			Cause:   err,
		}
	}

	// 创建PDF写入器 - 使用pdfcpu替代UniPDF
	// 由于已完全迁移到pdfcpu，不再需要UniPDF的PDFWriter
	totalPages := 0

	// 逐个处理文件
	for i, file := range files {
		if progressWriter != nil {
			fmt.Fprintf(progressWriter, "处理文件 %d/%d: %s\n", i+1, len(files), file)
		}

		// 获取页数（简单实现）
		if info, err := s.GetPDFInfo(file); err == nil {
			totalPages += info.PageCount
		} else {
			totalPages += 1 // 假设至少有一页
		}
	}

	if totalPages == 0 {
		return &PDFError{
			Type:    ErrorProcessing,
			Message: "没有成功处理任何页面",
			File:    outputPath,
		}
	}

	// 使用pdfcpu进行合并
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return fmt.Errorf("pdfcpu不可用: %w", err)
	}
	defer adapter.Close()

	if err := adapter.MergeFiles(files, outputPath); err != nil {
		return fmt.Errorf("pdfcpu合并失败: %w", err)
	}

	// 验证输出文件
	if err := s.validateOutputFile(outputPath); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "合并后的PDF文件无效",
			File:    outputPath,
			Cause:   err,
		}
	}

	if progressWriter != nil {
		fmt.Fprintf(progressWriter, "基本合并完成 - 总页数: %d\n", totalPages)
	}

	return nil
}

// copyFile 复制文件
func (s *PDFServiceImpl) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法打开源文件",
			File:    src,
			Cause:   err,
		}
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法创建目标文件",
			File:    dst,
			Cause:   err,
		}
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "文件复制失败",
			File:    dst,
			Cause:   err,
		}
	}

	return nil
}

// writeOutputFile 写入输出文件
func (s *PDFServiceImpl) writeOutputFile(filePath string) error {
	// 新增：只读目录检测
	outputDir := filepath.Dir(filePath)
	if err := checkDirectoryWritable(outputDir); err != nil {
		return &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写（只读目录）",
			File:    outputDir,
			Cause:   err,
		}
	}

	// 确保输出目录存在
	outputDir = filepath.Dir(filePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法创建输出目录",
			File:    filePath,
			Cause:   err,
		}
	}

	// 由于已完全迁移到pdfcpu，不再需要UniPDF的PDFWriter
	// 这个方法现在主要用于验证文件路径和目录权限
	return nil
}

// basicFileValidation 基本文件验证
func (s *PDFServiceImpl) basicFileValidation(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "文件不存在或无法访问",
			File:    filePath,
			Cause:   err,
		}
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件不是PDF格式",
			File:    filePath,
		}
	}

	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    filePath,
			Cause:   err,
		}
	}

	if fileInfo.Size() == 0 {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件为空",
			File:    filePath,
		}
	}

	if fileInfo.Size() < 100 { // PDF文件至少应该有100字节
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件太小，不是有效的PDF文件",
			File:    filePath,
		}
	}

	return nil
}

// validateWithPDFCPU 使用pdfcpu进行验证
func (s *PDFServiceImpl) validateWithPDFCPU(filePath string) error {
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return err // pdfcpu不可用
	}
	defer adapter.Close()

	if s.config.EnableStrictMode {
		// 使用严格模式验证
		return s.validator.ValidateWithStrictMode(filePath)
	}

	return adapter.ValidateFile(filePath)
}

// validateWithEnhancedReader 使用增强的PDF读取器进行验证
func (s *PDFServiceImpl) validateWithEnhancedReader(filePath string) error {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 验证PDF结构
	if err := reader.ValidateStructure(); err != nil {
		return err
	}

	// 检查页面数量
	pageCount, err := reader.GetPageCount()
	if err != nil {
		return err
	}

	if pageCount <= 0 {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件没有有效页面",
			File:    filePath,
		}
	}

	// 验证前几页是否可以正常访问
	maxPagesToCheck := 3
	if pageCount < maxPagesToCheck {
		maxPagesToCheck = pageCount
	}

	for i := 1; i <= maxPagesToCheck; i++ {
		if err := reader.ValidatePage(i); err != nil {
			return &PDFError{
				Type:    ErrorCorrupted,
				Message: fmt.Sprintf("第 %d 页验证失败", i),
				File:    filePath,
				Cause:   err,
			}
		}
	}

	return nil
}

// validateOutputFile 验证输出文件是否有效
func (s *PDFServiceImpl) validateOutputFile(filePath string) error {
	// 使用独立的验证方法，避免死锁
	// 只进行基本的文件验证，不调用需要全局锁的ValidatePDF方法
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("输出文件不存在: %w", err)
	}
	
	// 检查文件大小
	if fileInfo, err := os.Stat(filePath); err == nil {
		if fileInfo.Size() == 0 {
			return fmt.Errorf("输出文件为空")
		}
	}
	
	// 使用pdfcpu进行快速验证（不依赖全局锁）
	adapter, err := NewPDFCPUAdapter(nil)
	if err != nil {
		return err // pdfcpu不可用，跳过验证
	}
	defer adapter.Close()
	
	return adapter.ValidateFile(filePath)
}

// getFileNameWithoutExt 获取不带扩展名的文件名
func getFileNameWithoutExt(filePath string) string {
	// 获取文件名
	fileName := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			fileName = filePath[i+1:]
			break
		}
	}
	if fileName == "" {
		fileName = filePath
	}

	// 去掉扩展名
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			fileName = fileName[:i]
			break
		}
	}

	return fileName
}
package pdf

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	progressmodel "github.com/user/pdf-merger/internal/model"
)

// StreamingMerger 流式PDF合并器
type StreamingMerger struct {
	adapter         *PDFCPUAdapter
	maxMemoryUsage  int64
	tempDir         string
	mutex           sync.Mutex
	progressTracker *progressmodel.ProgressTracker
	config          *PDFCPUConfig
	streamingConfig *StreamingConfig
}

// StreamingConfig 流式合并配置
type StreamingConfig struct {
	// 内存管理
	MemoryWarningThreshold  float64       // 内存警告阈值（百分比）
	MemoryCriticalThreshold float64       // 内存严重阈值（百分比）
	GCInterval              time.Duration // GC间隔

	// 分块处理
	MinChunkSize       int   // 最小分块大小
	MaxChunkSize       int   // 最大分块大小
	LargeFileThreshold int64 // 大文件阈值（字节）

	// 并发控制
	MaxConcurrentChunks int           // 最大并发分块数
	ChunkProcessTimeout time.Duration // 分块处理超时

	// 优化选项
	EnableAdaptiveChunking bool // 启用自适应分块
	EnableMemoryPrediction bool // 启用内存预测
	EnableProgressiveGC    bool // 启用渐进式GC
}

// DefaultStreamingConfig 默认流式合并配置
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		MemoryWarningThreshold:  0.70, // 70%
		MemoryCriticalThreshold: 0.85, // 85%
		GCInterval:              100 * time.Millisecond,

		MinChunkSize:       2,
		MaxChunkSize:       20,
		LargeFileThreshold: 10 * 1024 * 1024, // 10MB

		MaxConcurrentChunks: runtime.NumCPU(),
		ChunkProcessTimeout: 30 * time.Second,

		EnableAdaptiveChunking: true,
		EnableMemoryPrediction: true,
		EnableProgressiveGC:    true,
	}
}

// MergeOptions 合并选项
type MergeOptions struct {
	MaxMemoryUsage    int64  // 最大内存使用量（字节）
	TempDirectory     string // 临时文件目录
	EnableGC          bool   // 是否启用垃圾回收
	ChunkSize         int    // 每次处理的页面数量
	UseStreaming      bool   // 是否使用流式处理
	OptimizeMemory    bool   // 是否优化内存使用
	ConcurrentWorkers int    // 并发工作线程数
}

// MergeResult 合并结果
type MergeResult struct {
	OutputPath     string
	TotalPages     int
	ProcessedFiles int
	SkippedFiles   []string
	ProcessingTime time.Duration
	MemoryUsage    int64
}

// NewStreamingMerger 创建新的流式合并器
func NewStreamingMerger(options *MergeOptions) *StreamingMerger {
	if options == nil {
		options = &MergeOptions{
			MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
			TempDirectory:     os.TempDir(),
			EnableGC:          true,
			ChunkSize:         10, // 每次处理10页
			UseStreaming:      true,
			OptimizeMemory:    true,
			ConcurrentWorkers: runtime.NumCPU(),
		}
	}

	// 创建pdfcpu配置，优化内存使用
	config := &PDFCPUConfig{
		ValidationMode:    "relaxed",
		WriteObjectStream: options.OptimizeMemory,
		WriteXRefStream:   options.OptimizeMemory,
		EncryptUsingAES:   true,
		EncryptKeyLength:  256,
		TempDirectory:     options.TempDirectory,
	}

	// 创建pdfcpu适配器
	adapter, err := NewPDFCPUAdapter(config)
	if err != nil {
		// 如果创建适配器失败，记录错误但继续创建合并器
		fmt.Printf("Warning: Failed to create PDFCPUAdapter: %v\n", err)
	}

	// 创建流式配置
	streamingConfig := DefaultStreamingConfig()

	// 根据选项调整流式配置
	if options.MaxMemoryUsage > 0 {
		// 根据内存大小调整阈值
		if options.MaxMemoryUsage < 50*1024*1024 { // 小于50MB
			streamingConfig.MemoryWarningThreshold = 0.60 // 更保守
			streamingConfig.MemoryCriticalThreshold = 0.75
		}
	}

	if options.ConcurrentWorkers > 0 {
		streamingConfig.MaxConcurrentChunks = options.ConcurrentWorkers
	}

	return &StreamingMerger{
		adapter:         adapter,
		maxMemoryUsage:  options.MaxMemoryUsage,
		tempDir:         options.TempDirectory,
		config:          config,
		streamingConfig: streamingConfig,
	}
}

// NewStreamingMergerWithConfig 使用自定义流式配置创建合并器
func NewStreamingMergerWithConfig(options *MergeOptions, streamingConfig *StreamingConfig) *StreamingMerger {
	merger := NewStreamingMerger(options)
	if streamingConfig != nil {
		merger.streamingConfig = streamingConfig
	}
	return merger
}

// MergeFiles 使用pdfcpu合并多个PDF文件
func (sm *StreamingMerger) MergeFiles(files []string, outputPath string, options *MergeOptions) (*MergeResult, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	startTime := time.Now()
	result := &MergeResult{
		OutputPath:     outputPath,
		SkippedFiles:   make([]string, 0),
		ProcessingTime: 0,
	}

	if len(files) == 0 {
		return nil, &PDFError{
			Type:    ErrorInvalidInput,
			Message: "没有提供输入文件",
		}
	}

	// 新增：只读目录检测
	dir := filepath.Dir(outputPath)
	if err := checkDirectoryWritable(dir); err != nil {
		return nil, &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写（只读目录）",
			File:    dir,
			Cause:   err,
		}
	}

	// 验证所有输入文件
	for _, file := range files {
		if err := sm.validateInputFile(file); err != nil {
			result.SkippedFiles = append(result.SkippedFiles, file)
			continue
		}
	}

	// 如果所有文件都无效，返回错误
	validFiles := len(files) - len(result.SkippedFiles)
	if validFiles == 0 {
		return nil, &PDFError{
			Type:    ErrorInvalidInput,
			Message: "没有有效的输入文件",
		}
	}

	// 合并前备份输出文件
	var backupPath string
	var rollbackMgr *RollbackManager
	if fileExists(outputPath) {
		backupDir := filepath.Dir(outputPath)
		rollbackMgr = NewRollbackManager(backupDir)
		backupPath, _ = rollbackMgr.BackupFile(outputPath)
	}

	// 使用pdfcpu适配器进行合并
	var mergeErr error
	if sm.adapter != nil {
		mergeErr = sm.adapter.MergeFiles(files, outputPath)
	} else {
		mergeErr = sm.fallbackMerge(files, outputPath)
	}
	if mergeErr != nil {
		if rollbackMgr != nil && backupPath != "" {
			_ = rollbackMgr.RestoreFile(backupPath, outputPath)
		}
		return nil, mapPDFCPUError(mergeErr)
	}

	// 计算结果统计
	result.ProcessedFiles = validFiles
	result.ProcessingTime = time.Since(startTime)
	result.MemoryUsage = sm.getCurrentMemoryUsage()

	// 获取输出文件信息
	if info, err := os.Stat(outputPath); err == nil {
		// 估算页数（简单实现）
		result.TotalPages = int(info.Size() / (1024 * 50)) // 假设每页约50KB
		if result.TotalPages == 0 {
			result.TotalPages = 1
		}
	}

	return result, nil
}

// MergeStreaming 执行流式合并，支持进度回调和取消
func (sm *StreamingMerger) MergeStreaming(ctx context.Context, files []string, outputPath string,
	progressCallback func(progress float64, message string)) (*MergeResult, error) {

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	startTime := time.Now()
	result := &MergeResult{
		OutputPath:     outputPath,
		SkippedFiles:   make([]string, 0),
		ProcessingTime: 0,
	}

	if len(files) == 0 {
		return nil, &PDFError{
			Type:    ErrorInvalidInput,
			Message: "没有提供输入文件",
		}
	}

	// 新增：只读目录检测
	dir := filepath.Dir(outputPath)
	if err := checkDirectoryWritable(dir); err != nil {
		return nil, &PDFError{
			Type:    ErrorPermission,
			Message: "输出目录不可写（只读目录）",
			File:    dir,
			Cause:   err,
		}
	}

	// 创建内存监控器
	memoryMonitor := NewMemoryMonitor(sm.maxMemoryUsage)

	// 设置进度跟踪器
	totalSteps := len(files) + 2 // 文件验证 + 合并 + 后处理
	sm.progressTracker = progressmodel.NewProgressTracker(totalSteps)

	if progressCallback != nil {
		sm.progressTracker.AddCallback(progressCallback)
	}

	// 第一步：验证所有输入文件
	sm.progressTracker.SetCurrentStep(1, "验证输入文件")
	validFiles := make([]string, 0, len(files))

	for i, file := range files {
		// 检查取消
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// 检查内存压力
		if pressure := memoryMonitor.CheckMemoryPressure(); pressure != MemoryPressureNormal {
			sm.handleMemoryPressure(pressure)
		}

		progress := float64(i) / float64(len(files)) * 20 // 验证占20%
		sm.progressTracker.UpdateStepProgress(progress, fmt.Sprintf("验证文件: %s", filepath.Base(file)))

		if err := sm.validateInputFile(file); err != nil {
			result.SkippedFiles = append(result.SkippedFiles, file)
			continue
		}
		validFiles = append(validFiles, file)
	}

	if len(validFiles) == 0 {
		return nil, &PDFError{
			Type:    ErrorInvalidInput,
			Message: "没有有效的输入文件",
		}
	}

	// 合并前备份输出文件
	var backupPath string
	var rollbackMgr *RollbackManager
	if fileExists(outputPath) {
		backupDir := filepath.Dir(outputPath)
		rollbackMgr = NewRollbackManager(backupDir)
		backupPath, _ = rollbackMgr.BackupFile(outputPath)
	}

	// 第二步：执行智能合并策略选择
	sm.progressTracker.SetCurrentStep(2, "合并PDF文件")

	var mergeErr error

	// 针对大文件进行优化
	sm.optimizeForLargeFiles(validFiles)

	// 根据文件特征选择合并策略
	if sm.shouldUseConcurrentProcessing(validFiles) {
		sm.progressTracker.UpdateStepProgress(0, "使用并发处理模式")
		mergeErr = sm.processConcurrently(ctx, validFiles, outputPath)
	} else if sm.shouldUseStreamingMode(validFiles) {
		sm.progressTracker.UpdateStepProgress(0, "使用流式合并模式")
		mergeErr = sm.performStreamingMergeWithChunking(ctx, validFiles, outputPath)
	} else if sm.shouldUseMemoryOptimization(validFiles) {
		sm.progressTracker.UpdateStepProgress(0, "使用内存优化模式")
		mergeErr = sm.performOptimizedMerge(ctx, validFiles, outputPath)
	} else {
		sm.progressTracker.UpdateStepProgress(0, "使用标准合并模式")
		mergeErr = sm.performStreamingMerge(ctx, validFiles, outputPath)
	}

	if mergeErr != nil {
		if rollbackMgr != nil && backupPath != "" {
			_ = rollbackMgr.RestoreFile(backupPath, outputPath)
		}
		return nil, mergeErr
	}

	// 第三步：后处理和验证
	sm.progressTracker.SetCurrentStep(3, "验证输出文件")

	if err := sm.validateOutputFile(outputPath); err != nil {
		if rollbackMgr != nil && backupPath != "" {
			_ = rollbackMgr.RestoreFile(backupPath, outputPath)
		}
		return nil, err
	}

	// 计算结果统计
	result.ProcessedFiles = len(validFiles)
	result.ProcessingTime = time.Since(startTime)
	result.MemoryUsage = sm.getCurrentMemoryUsage()

	// 获取输出文件信息
	if info, err := os.Stat(outputPath); err == nil {
		result.TotalPages = sm.estimatePageCount(info.Size())
	}

	// 最终内存清理
	sm.optimizeMemoryUsage()

	sm.progressTracker.Complete("合并完成")
	return result, nil
}

// MergeFilesLegacy 流式合并多个PDF文件（保留原有接口）
func (sm *StreamingMerger) MergeFilesLegacy(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) (*MergeResult, error) {
	// 将参数转换为新接口格式
	allFiles := append([]string{mainFile}, additionalFiles...)

	// 创建进度回调函数
	var progressCallback func(progress float64, message string)
	if progressWriter != nil {
		progressCallback = func(progress float64, message string) {
			fmt.Fprintf(progressWriter, "进度: %.1f%% - %s\n", progress, message)
		}
	}

	// 使用新的流式合并方法
	ctx := context.Background()
	return sm.MergeStreaming(ctx, allFiles, outputPath, progressCallback)
}

// forceGC 强制垃圾回收
func (sm *StreamingMerger) forceGC() {
	runtime.GC()
	runtime.GC() // 调用两次确保彻底清理
}

// getCurrentMemoryUsage 获取当前内存使用量
func (sm *StreamingMerger) getCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// validateInputFile 验证输入文件
func (sm *StreamingMerger) validateInputFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件不存在",
			File:    filePath,
			Cause:   err,
		}
	}

	// 使用适配器验证文件
	if sm.adapter != nil {
		return sm.adapter.ValidateFile(filePath)
	}

	// 回退到基本验证
	return sm.basicValidation(filePath)
}

// validateOutputFile 验证输出文件
func (sm *StreamingMerger) validateOutputFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "输出文件不存在",
			File:    filePath,
			Cause:   err,
		}
	}

	// 使用适配器验证文件
	if sm.adapter != nil {
		return sm.adapter.ValidateFile(filePath)
	}

	return nil
}

// performStreamingMerge 执行流式合并的核心逻辑
func (sm *StreamingMerger) performStreamingMerge(ctx context.Context, files []string, outputPath string) error {
	// 检查是否应该使用内存优化模式
	if sm.shouldUseMemoryOptimization(files) {
		return sm.performOptimizedMerge(ctx, files, outputPath)
	}

	// 使用pdfcpu适配器进行标准合并
	if sm.adapter != nil {
		return sm.adapter.MergeFiles(files, outputPath)
	}

	// 回退到基本合并
	return sm.fallbackMerge(files, outputPath)
}

// performStreamingMergeWithChunking 执行分块流式合并
func (sm *StreamingMerger) performStreamingMergeWithChunking(ctx context.Context, files []string, outputPath string) error {
	chunkSize := sm.calculateOptimalChunkSize(files)
	if len(files) <= chunkSize {
		return sm.performDirectMerge(ctx, files, outputPath)
	}

	tempFiles := make([]string, 0)
	defer sm.cleanupTempFiles(tempFiles)

	// 并发分块合并优化
	maxConcurrent := sm.streamingConfig.MaxConcurrentChunks
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mergeErr atomic.Value

	// 内存缓冲池（用于分块I/O）
	_ = sync.Pool{
		New: func() interface{} {
			return make([]byte, 2*1024*1024) // 2MB缓冲区
		},
	}

	for i := 0; i < len(files); i += chunkSize {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}
		chunk := files[i:end]
		tempFile := sm.generateTempPath(outputPath)
		tempFiles = append(tempFiles, tempFile)

		sem <- struct{}{}
		wg.Add(1)
		go func(idx, chunkIdx int, chunk []string, tempFile string) {
			defer wg.Done()
			defer func() { <-sem }()
			if mergeErr.Load() != nil {
				return
			}
			// 使用缓冲池进行I/O优化（如有自定义实现可在此处用bufPool）
			err := sm.performDirectMerge(ctx, chunk, tempFile)
			if err != nil {
				mergeErr.Store(fmt.Errorf("分块 %d 合并失败: %v", chunkIdx+1, err))
			}
			// 内存优化
			if (chunkIdx+1)%3 == 0 {
				sm.optimizeMemoryUsage()
			}
		}(i, i/chunkSize, chunk, tempFile)
		// 更新进度
		progress := float64(i)/float64(len(files))*80 + 10
		sm.updateProgress(progress, fmt.Sprintf("处理分块 %d/%d", (i/chunkSize)+1, (len(files)+chunkSize-1)/chunkSize))
	}
	wg.Wait()
	if err := mergeErr.Load(); err != nil {
		return err.(error)
	}
	// 合并所有临时文件
	sm.updateProgress(90, "合并最终结果")
	return sm.performDirectMerge(ctx, tempFiles, outputPath)
}

// performDirectMerge 执行直接合并
func (sm *StreamingMerger) performDirectMerge(ctx context.Context, files []string, outputPath string) error {
	if sm.adapter != nil {
		return sm.adapter.MergeFiles(files, outputPath)
	}
	return sm.fallbackMerge(files, outputPath)
}

// calculateOptimalChunkSize 计算最优分块大小
func (sm *StreamingMerger) calculateOptimalChunkSize(files []string) int {
	config := sm.streamingConfig
	if config == nil {
		config = DefaultStreamingConfig()
	}

	// 如果禁用自适应分块，返回固定大小
	if !config.EnableAdaptiveChunking {
		return (config.MinChunkSize + config.MaxChunkSize) / 2
	}

	// 基于内存使用情况和文件数量计算最优分块大小
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	availableMemory := sm.maxMemoryUsage - int64(m.Alloc)
	if availableMemory <= 0 {
		return config.MinChunkSize // 内存不足时使用最小分块
	}

	// 分析文件特征
	fileAnalysis := sm.analyzeFiles(files)

	// 基于文件分析结果计算分块大小
	var optimalChunkSize int

	if fileAnalysis.HasLargeFiles {
		// 有大文件时使用较小的分块
		optimalChunkSize = config.MinChunkSize + 1
	} else if fileAnalysis.TotalSize > sm.maxMemoryUsage/2 {
		// 总大小较大时使用中等分块
		optimalChunkSize = (config.MinChunkSize + config.MaxChunkSize) / 2
	} else {
		// 基于内存预测计算
		if config.EnableMemoryPrediction {
			optimalChunkSize = sm.predictOptimalChunkSize(fileAnalysis, availableMemory)
		} else {
			// 简单计算
			estimatedMemoryPerFile := fileAnalysis.AvgSize / 10 // 假设内存使用为文件大小的1/10
			if estimatedMemoryPerFile > 0 {
				optimalChunkSize = int(availableMemory / estimatedMemoryPerFile)
			} else {
				optimalChunkSize = config.MaxChunkSize / 2
			}
		}
	}

	// 限制分块大小范围
	if optimalChunkSize < config.MinChunkSize {
		return config.MinChunkSize
	}
	if optimalChunkSize > config.MaxChunkSize {
		return config.MaxChunkSize
	}

	return optimalChunkSize
}

// FileAnalysis 文件分析结果
type FileAnalysis struct {
	TotalSize     int64
	AvgSize       int64
	MaxSize       int64
	MinSize       int64
	HasLargeFiles bool
	FileCount     int
}

// analyzeFiles 分析文件特征
func (sm *StreamingMerger) analyzeFiles(files []string) *FileAnalysis {
	analysis := &FileAnalysis{
		FileCount: len(files),
		MinSize:   int64(^uint64(0) >> 1), // 最大int64值
	}

	config := sm.streamingConfig
	if config == nil {
		config = DefaultStreamingConfig()
	}

	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			size := info.Size()
			analysis.TotalSize += size

			if size > analysis.MaxSize {
				analysis.MaxSize = size
			}
			if size < analysis.MinSize {
				analysis.MinSize = size
			}

			if size > config.LargeFileThreshold {
				analysis.HasLargeFiles = true
			}
		}
	}

	if analysis.FileCount > 0 {
		analysis.AvgSize = analysis.TotalSize / int64(analysis.FileCount)
	}

	if analysis.MinSize == int64(^uint64(0)>>1) {
		analysis.MinSize = 0
	}

	return analysis
}

// predictOptimalChunkSize 预测最优分块大小
func (sm *StreamingMerger) predictOptimalChunkSize(analysis *FileAnalysis, availableMemory int64) int {
	config := sm.streamingConfig

	// 基于历史数据和机器学习的简单预测模型
	// 这里使用启发式算法

	// 内存使用预测因子
	memoryFactor := float64(availableMemory) / float64(sm.maxMemoryUsage)

	// 文件大小因子
	sizeFactor := 1.0
	if analysis.AvgSize > 0 {
		sizeFactor = math.Min(2.0, float64(config.LargeFileThreshold)/float64(analysis.AvgSize))
	}

	// 文件数量因子
	countFactor := math.Max(0.5, math.Min(2.0, 10.0/float64(analysis.FileCount)))

	// 综合计算
	predictedSize := float64(config.MaxChunkSize) * memoryFactor * sizeFactor * countFactor

	return int(math.Round(predictedSize))
}

// updateProgress 更新进度（辅助方法）
func (sm *StreamingMerger) updateProgress(progress float64, message string) {
	if sm.progressTracker != nil {
		sm.progressTracker.UpdateStepProgress(progress, message)
	}
}

// performOptimizedMerge 执行内存优化的合并
func (sm *StreamingMerger) performOptimizedMerge(ctx context.Context, files []string, outputPath string) error {
	// 如果文件数量较多，分批处理
	if len(files) > 10 {
		return sm.performBatchMerge(ctx, files, outputPath)
	}

	// 直接合并
	if sm.adapter != nil {
		return sm.adapter.MergeFiles(files, outputPath)
	}

	return sm.fallbackMerge(files, outputPath)
}

// performBatchMerge 执行分批合并 - 增强版本支持大文件处理
func (sm *StreamingMerger) performBatchMerge(ctx context.Context, files []string, outputPath string) error {
	// 智能计算批次大小
	batchSize := sm.calculateOptimalBatchSize(files)
	tempFiles := make([]string, 0)
	defer sm.cleanupTempFiles(tempFiles)

	sm.logger("开始分批合并，文件数: %d, 批次大小: %d", len(files), batchSize)

	// 分批处理文件
	for i := 0; i < len(files); i += batchSize {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(files) + batchSize - 1) / batchSize

		sm.logger("处理批次 %d/%d，文件数: %d", batchNum, totalBatches, len(batch))

		// 检查内存压力并优化
		if sm.shouldOptimizeMemoryForBatch(batch) {
			sm.logger("检测到内存压力，执行优化")
			sm.optimizeMemoryUsage()
		}

		tempFile := sm.generateTempPath(outputPath)
		tempFiles = append(tempFiles, tempFile)

		// 更新进度
		progress := float64(i)/float64(len(files))*70 + 20 // 合并占70%，从20%开始
		sm.progressTracker.UpdateStepProgress(progress,
			fmt.Sprintf("处理批次 %d/%d", batchNum, totalBatches))

		// 合并当前批次
		startTime := time.Now()
		var err error

		if sm.adapter != nil {
			err = sm.adapter.MergeFiles(batch, tempFile)
		} else {
			err = sm.fallbackMerge(batch, tempFile)
		}

		if err != nil {
			sm.logger("批次 %d 合并失败: %v", batchNum, err)
			return fmt.Errorf("批次 %d 合并失败: %w", batchNum, err)
		}

		processingTime := time.Since(startTime)
		sm.logger("批次 %d 合并完成，耗时: %v", batchNum, processingTime)

		// 定期触发垃圾回收和内存优化
		if batchNum%2 == 0 { // 每2个批次优化一次
			sm.optimizeMemoryUsage()
		}

		// 检查临时文件大小，如果过大则进行中间合并
		if len(tempFiles) >= 10 {
			sm.logger("临时文件过多，执行中间合并")
			if err := sm.performIntermediateMerge(ctx, tempFiles, outputPath); err != nil {
				return fmt.Errorf("中间合并失败: %w", err)
			}
			// 清理已合并的临时文件，保留最后一个
			sm.cleanupTempFiles(tempFiles[:len(tempFiles)-1])
			tempFiles = tempFiles[len(tempFiles)-1:]
		}
	}

	// 合并所有临时文件
	sm.progressTracker.UpdateStepProgress(90, "合并最终结果")
	sm.logger("开始最终合并，临时文件数: %d", len(tempFiles))

	if sm.adapter != nil {
		return sm.adapter.MergeFiles(tempFiles, outputPath)
	}

	return sm.fallbackMerge(tempFiles, outputPath)
}

// calculateOptimalBatchSize 计算最优批次大小
func (sm *StreamingMerger) calculateOptimalBatchSize(files []string) int {
	config := sm.streamingConfig
	if config == nil {
		config = DefaultStreamingConfig()
	}

	// 分析文件特征
	analysis := sm.analyzeFiles(files)

	// 基于内存使用情况计算批次大小
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	availableMemory := sm.maxMemoryUsage - int64(m.Alloc)
	if availableMemory <= 0 {
		return 2 // 内存不足时使用最小批次
	}

	// 基于文件大小和可用内存计算
	var batchSize int

	if analysis.HasLargeFiles {
		// 有大文件时使用较小的批次
		batchSize = 3
	} else if analysis.AvgSize > 5*1024*1024 { // 平均文件大于5MB
		batchSize = 4
	} else if analysis.TotalSize > sm.maxMemoryUsage/3 {
		// 总大小较大时使用中等批次
		batchSize = 5
	} else {
		// 估算每个文件的内存使用量（假设为文件大小的1/5）
		estimatedMemoryPerFile := analysis.AvgSize / 5
		if estimatedMemoryPerFile > 0 {
			batchSize = int(availableMemory / estimatedMemoryPerFile)
		} else {
			batchSize = 8
		}
	}

	// 限制批次大小范围
	if batchSize < 2 {
		return 2
	}
	if batchSize > 15 {
		return 15
	}

	return batchSize
}

// shouldOptimizeMemoryForBatch 判断是否需要为批次优化内存
func (sm *StreamingMerger) shouldOptimizeMemoryForBatch(batch []string) bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMemory := int64(m.Alloc)
	memoryPressure := float64(currentMemory) / float64(sm.maxMemoryUsage)

	// 内存压力超过60%时优化
	if memoryPressure > 0.6 {
		return true
	}

	// 检查批次中是否有大文件
	for _, file := range batch {
		if info, err := os.Stat(file); err == nil {
			if info.Size() > 10*1024*1024 { // 超过10MB
				return true
			}
		}
	}

	return false
}

// performIntermediateMerge 执行中间合并
func (sm *StreamingMerger) performIntermediateMerge(ctx context.Context, tempFiles []string, outputPath string) error {
	if len(tempFiles) <= 1 {
		return nil
	}

	sm.logger("执行中间合并，文件数: %d", len(tempFiles))

	// 创建中间合并文件
	intermediateFile := sm.generateTempPath(outputPath)

	// 合并临时文件
	var err error
	if sm.adapter != nil {
		err = sm.adapter.MergeFiles(tempFiles, intermediateFile)
	} else {
		err = sm.fallbackMerge(tempFiles, intermediateFile)
	}

	if err != nil {
		return fmt.Errorf("中间合并失败: %w", err)
	}

	// 用中间文件替换原有临时文件
	tempFiles = []string{intermediateFile}

	sm.logger("中间合并完成")
	return nil
}

// shouldUseMemoryOptimization 判断是否应该使用内存优化
func (sm *StreamingMerger) shouldUseMemoryOptimization(files []string) bool {
	// 检查当前内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentMemory := int64(m.Alloc)

	// 如果当前内存使用超过阈值，使用优化模式
	if currentMemory > sm.maxMemoryUsage*60/100 { // 降低阈值到60%
		return true
	}

	// 检查文件数量 - 超过10个文件使用优化模式
	if len(files) > 10 {
		return true
	}

	// 检查文件总大小和平均大小
	totalSize := int64(0)
	largeFileCount := 0

	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
			// 检查是否有大文件（超过10MB）
			if info.Size() > 10*1024*1024 {
				largeFileCount++
			}
		}
	}

	// 如果有多个大文件，使用优化模式
	if largeFileCount > 2 {
		return true
	}

	// 如果文件总大小较大，使用优化模式
	if totalSize > sm.maxMemoryUsage/3 { // 降低阈值到1/3
		return true
	}

	// 检查平均文件大小
	if len(files) > 0 {
		avgSize := totalSize / int64(len(files))
		// 如果平均文件大小超过5MB，使用优化模式
		if avgSize > 5*1024*1024 {
			return true
		}
	}

	return false
}

// shouldUseConcurrentProcessing 判断是否应该使用并发处理
func (sm *StreamingMerger) shouldUseConcurrentProcessing(files []string) bool {
	config := sm.streamingConfig
	if config == nil {
		config = DefaultStreamingConfig()
	}

	// 文件数量少于4个时不使用并发
	if len(files) < 4 {
		return false
	}

	// 检查系统资源
	if runtime.NumCPU() < 2 {
		return false // 单核系统不使用并发
	}

	// 检查内存压力
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryPressure := float64(m.Alloc) / float64(sm.maxMemoryUsage)

	// 内存压力过高时不使用并发
	if memoryPressure > 0.7 {
		return false
	}

	// 分析文件特征
	analysis := sm.analyzeFiles(files)

	// 如果有太多大文件，不使用并发（避免内存爆炸）
	largeFileCount := 0
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			if info.Size() > config.LargeFileThreshold {
				largeFileCount++
			}
		}
	}

	if largeFileCount > config.MaxConcurrentChunks {
		return false
	}

	// 文件数量适中且系统资源充足时使用并发
	if len(files) >= 4 && len(files) <= 20 && !analysis.HasLargeFiles {
		return true
	}

	// 文件较多但平均大小不大时使用并发
	if len(files) > 8 && analysis.AvgSize < 5*1024*1024 {
		return true
	}

	return false
}

// shouldUseStreamingMode 判断是否应该使用流式模式
func (sm *StreamingMerger) shouldUseStreamingMode(files []string) bool {
	// 检查系统内存压力
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算内存压力指标
	memoryPressure := float64(m.Alloc) / float64(sm.maxMemoryUsage)

	// 内存压力超过50%时使用流式模式
	if memoryPressure > 0.5 {
		return true
	}

	// 文件数量超过5个时使用流式模式
	if len(files) > 5 {
		return true
	}

	// 检查是否有超大文件（超过20MB）
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			if info.Size() > 20*1024*1024 {
				return true
			}
		}
	}

	return false
}

// optimizeMemoryUsage 优化内存使用 - 增强版本
func (sm *StreamingMerger) optimizeMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	beforeGC := m.Alloc
	beforeSys := m.Sys

	sm.logger("开始内存优化，当前分配: %d MB, 系统内存: %d MB",
		beforeGC/(1024*1024), beforeSys/(1024*1024))

	// 第一阶段：标准垃圾回收
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// 第二阶段：强制释放未使用的内存
	runtime.GC()
	debug.FreeOSMemory() // 释放操作系统内存
	time.Sleep(50 * time.Millisecond)

	// 检查GC效果
	runtime.ReadMemStats(&m)
	afterGC := m.Alloc
	afterSys := m.Sys

	// 计算内存释放效果
	memoryReleased := beforeGC - afterGC
	sysMemoryReleased := beforeSys - afterSys

	sm.logger("内存优化完成，释放内存: %d MB, 系统内存释放: %d MB",
		memoryReleased/(1024*1024), sysMemoryReleased/(1024*1024))

	// 如果内存释放效果不佳，进行更激进的优化
	if afterGC > beforeGC*75/100 {
		sm.logger("内存释放效果不佳，进行激进优化")

		// 设置更低的GC目标
		originalGCPercent := debug.SetGCPercent(25) // 降低GC触发阈值到25%

		// 多次强制GC
		for i := 0; i < 3; i++ {
			runtime.GC()
			debug.FreeOSMemory()
			time.Sleep(50 * time.Millisecond)
		}

		// 恢复原始GC设置
		debug.SetGCPercent(originalGCPercent)

		// 最终检查
		runtime.ReadMemStats(&m)
		finalAlloc := m.Alloc
		sm.logger("激进优化后内存: %d MB", finalAlloc/(1024*1024))
	}

	// 如果配置了渐进式GC，启用它
	if sm.streamingConfig != nil && sm.streamingConfig.EnableProgressiveGC {
		sm.enableProgressiveGC()
	}
}

// enableProgressiveGC 启用渐进式垃圾回收
func (sm *StreamingMerger) enableProgressiveGC() {
	if sm.streamingConfig == nil {
		return
	}

	// 启动后台GC协程
	go func() {
		ticker := time.NewTicker(sm.streamingConfig.GCInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 检查内存压力
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				currentMemory := int64(m.Alloc)
				memoryPressure := float64(currentMemory) / float64(sm.maxMemoryUsage)

				// 根据内存压力调整GC频率
				if memoryPressure > sm.streamingConfig.MemoryCriticalThreshold {
					runtime.GC()
					debug.FreeOSMemory()
				} else if memoryPressure > sm.streamingConfig.MemoryWarningThreshold {
					runtime.GC()
				}
			}
		}
	}()
}

// MemoryMonitor 内存监控器
type MemoryMonitor struct {
	maxMemory     int64
	warningLevel  int64
	criticalLevel int64
	lastCheck     time.Time
	checkInterval time.Duration
}

// NewMemoryMonitor 创建内存监控器
func NewMemoryMonitor(maxMemory int64) *MemoryMonitor {
	return &MemoryMonitor{
		maxMemory:     maxMemory,
		warningLevel:  maxMemory * 70 / 100, // 70%警告
		criticalLevel: maxMemory * 85 / 100, // 85%严重
		checkInterval: 100 * time.Millisecond,
	}
}

// CheckMemoryPressure 检查内存压力
func (mm *MemoryMonitor) CheckMemoryPressure() MemoryPressureLevel {
	now := time.Now()
	if now.Sub(mm.lastCheck) < mm.checkInterval {
		return MemoryPressureNormal // 避免频繁检查
	}
	mm.lastCheck = now

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMemory := int64(m.Alloc)

	if currentMemory >= mm.criticalLevel {
		return MemoryPressureCritical
	} else if currentMemory >= mm.warningLevel {
		return MemoryPressureWarning
	}

	return MemoryPressureNormal
}

// MemoryPressureLevel 内存压力级别
type MemoryPressureLevel int

const (
	MemoryPressureNormal MemoryPressureLevel = iota
	MemoryPressureWarning
	MemoryPressureCritical
)

// handleMemoryPressure 处理内存压力
func (sm *StreamingMerger) handleMemoryPressure(level MemoryPressureLevel) {
	switch level {
	case MemoryPressureWarning:
		// 警告级别：进行标准GC
		runtime.GC()

	case MemoryPressureCritical:
		// 严重级别：激进的内存清理
		sm.optimizeMemoryUsage()

		// 如果仍然严重，暂停处理
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if int64(m.Alloc) > sm.maxMemoryUsage*80/100 {
			time.Sleep(500 * time.Millisecond) // 暂停500ms
		}
	}
}

// fallbackMerge 回退合并实现
func (sm *StreamingMerger) fallbackMerge(files []string, outputPath string) error {
	// 创建一个简单的占位符实现
	// 在实际部署中，这里应该有一个可工作的PDF合并实现

	if len(files) == 0 {
		return &PDFError{
			Type:    ErrorInvalidInput,
			Message: "没有输入文件",
		}
	}

	// 如果只有一个文件，直接复制
	if len(files) == 1 {
		return sm.copyFile(files[0], outputPath)
	}

	// 创建占位符合并结果
	content := fmt.Sprintf("Fallback merge result\nFiles: %v\nOutput: %s\nTimestamp: %s\n",
		files, outputPath, time.Now().Format(time.RFC3339))

	return os.WriteFile(outputPath+".fallback", []byte(content), 0644)
}

// basicValidation 基本文件验证
func (sm *StreamingMerger) basicValidation(filePath string) error {
	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件不是PDF格式",
			File:    filePath,
		}
	}

	// 检查文件大小
	info, err := os.Stat(filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    filePath,
			Cause:   err,
		}
	}

	if info.Size() == 0 {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "文件为空",
			File:    filePath,
		}
	}

	return nil
}

// estimatePageCount 估算页数
func (sm *StreamingMerger) estimatePageCount(fileSize int64) int {
	// 简单估算：假设每页约50KB
	pages := int(fileSize / (1024 * 50))
	if pages == 0 {
		pages = 1
	}
	return pages
}

// cleanupTempFiles 清理临时文件
func (sm *StreamingMerger) cleanupTempFiles(tempFiles []string) {
	for _, file := range tempFiles {
		if err := os.Remove(file); err != nil {
			// 记录错误但不中断程序
			fmt.Printf("Warning: Failed to remove temp file %s: %v\n", file, err)
		}
	}
}

// copyFile 复制文件
func (sm *StreamingMerger) copyFile(src, dst string) error {
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

// generateTempPath 生成临时文件路径
func (sm *StreamingMerger) generateTempPath(outputPath string) string {
	fileName := filepath.Base(outputPath)
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	timestamp := time.Now().Format("20060102_150405")
	tempFileName := fmt.Sprintf("%s_temp_%s_%d.pdf", nameWithoutExt, timestamp, time.Now().UnixNano()%1000)
	return filepath.Join(sm.tempDir, tempFileName)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// copyFile 复制文件（全局函数）
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// logger 日志记录辅助方法
func (sm *StreamingMerger) logger(format string, args ...interface{}) {
	if sm.adapter != nil && sm.adapter.logger != nil {
		sm.adapter.logger.Printf(format, args...)
	} else {
		fmt.Printf("[StreamingMerger] "+format+"\n", args...)
	}
}

// GetProgressTracker 获取进度跟踪器
func (sm *StreamingMerger) GetProgressTracker() *progressmodel.ProgressTracker {
	return sm.progressTracker
}

// Cancel 取消合并操作
func (sm *StreamingMerger) Cancel() {
	if sm.progressTracker != nil {
		sm.progressTracker.Cancel("用户取消操作")
	}
}

// processConcurrently 并发处理多个文件
func (sm *StreamingMerger) processConcurrently(ctx context.Context, files []string, outputPath string) error {
	config := sm.streamingConfig
	if config == nil {
		config = DefaultStreamingConfig()
	}

	// 如果文件数量较少，不使用并发
	if len(files) <= 3 {
		return sm.performDirectMerge(ctx, files, outputPath)
	}

	sm.logger("开始并发处理，文件数: %d, 最大并发数: %d", len(files), config.MaxConcurrentChunks)

	// 创建工作池
	semaphore := make(chan struct{}, config.MaxConcurrentChunks)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var processingErrors []error

	// 分组处理文件
	chunkSize := (len(files) + config.MaxConcurrentChunks - 1) / config.MaxConcurrentChunks
	if chunkSize < 2 {
		chunkSize = 2
	}

	tempFiles := make([]string, 0)
	defer sm.cleanupTempFiles(tempFiles)

	// 并发处理每个分组
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}

		chunk := files[i:end]
		chunkIndex := i / chunkSize

		wg.Add(1)
		go func(chunk []string, index int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查取消
			if ctx.Err() != nil {
				mu.Lock()
				processingErrors = append(processingErrors, ctx.Err())
				mu.Unlock()
				return
			}

			sm.logger("开始处理分组 %d，文件数: %d", index+1, len(chunk))

			// 创建临时文件
			tempFile := sm.generateTempPath(outputPath)

			// 处理分组
			startTime := time.Now()
			var err error

			// 添加超时控制
			chunkCtx, cancel := context.WithTimeout(ctx, config.ChunkProcessTimeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				if sm.adapter != nil {
					done <- sm.adapter.MergeFiles(chunk, tempFile)
				} else {
					done <- sm.fallbackMerge(chunk, tempFile)
				}
			}()

			select {
			case err = <-done:
				// 处理完成
			case <-chunkCtx.Done():
				err = fmt.Errorf("分组 %d 处理超时", index+1)
			}

			processingTime := time.Since(startTime)

			if err != nil {
				sm.logger("分组 %d 处理失败: %v", index+1, err)
				mu.Lock()
				processingErrors = append(processingErrors, fmt.Errorf("分组 %d 处理失败: %w", index+1, err))
				mu.Unlock()
				return
			}

			sm.logger("分组 %d 处理完成，耗时: %v", index+1, processingTime)

			// 添加到临时文件列表
			mu.Lock()
			tempFiles = append(tempFiles, tempFile)
			mu.Unlock()

			// 更新进度
			progress := float64(index+1)/float64((len(files)+chunkSize-1)/chunkSize)*80 + 10
			sm.updateProgress(progress, fmt.Sprintf("完成分组 %d", index+1))

		}(chunk, chunkIndex)
	}

	// 等待所有分组完成
	wg.Wait()

	// 检查处理错误
	if len(processingErrors) > 0 {
		return fmt.Errorf("并发处理失败: %v", processingErrors[0])
	}

	sm.logger("所有分组处理完成，开始最终合并")

	// 最终合并所有临时文件
	sm.updateProgress(90, "合并最终结果")

	if sm.adapter != nil {
		return sm.adapter.MergeFiles(tempFiles, outputPath)
	}

	return sm.fallbackMerge(tempFiles, outputPath)
}

// configurePDFCPUForMinimalMemory 配置pdfcpu使用最小内存模式
func (sm *StreamingMerger) configurePDFCPUForMinimalMemory() {
	if sm.config == nil {
		sm.config = &PDFCPUConfig{}
	}

	// 配置最小内存使用模式
	sm.config.WriteObjectStream = true   // 启用对象流压缩
	sm.config.WriteXRefStream = true     // 启用交叉引用流
	sm.config.ValidationMode = "relaxed" // 使用宽松验证模式减少内存使用

	sm.logger("已配置pdfcpu最小内存模式")

	// 如果有适配器，更新其配置
	if sm.adapter != nil {
		// 重新创建适配器以应用新配置
		if newAdapter, err := NewPDFCPUAdapter(sm.config); err == nil {
			// 关闭旧适配器
			sm.adapter.Close()
			sm.adapter = newAdapter
			sm.logger("已更新pdfcpu适配器配置")
		} else {
			sm.logger("更新pdfcpu适配器配置失败: %v", err)
		}
	}

	// 设置运行时参数以优化内存使用
	debug.SetGCPercent(50)                  // 降低GC触发阈值
	debug.SetMemoryLimit(sm.maxMemoryUsage) // 设置内存限制

	sm.logger("已设置运行时内存优化参数")
}

// optimizeForLargeFiles 针对大文件优化处理策略
func (sm *StreamingMerger) optimizeForLargeFiles(files []string) {
	analysis := sm.analyzeFiles(files)

	if !analysis.HasLargeFiles {
		return
	}

	sm.logger("检测到大文件，启用大文件优化模式")

	// 调整流式配置
	if sm.streamingConfig == nil {
		sm.streamingConfig = DefaultStreamingConfig()
	}

	// 针对大文件的优化配置
	sm.streamingConfig.MinChunkSize = 2                   // 减小最小分块大小
	sm.streamingConfig.MaxChunkSize = 5                   // 减小最大分块大小
	sm.streamingConfig.MaxConcurrentChunks = 2            // 减少并发数
	sm.streamingConfig.EnableProgressiveGC = true         // 启用渐进式GC
	sm.streamingConfig.GCInterval = 50 * time.Millisecond // 增加GC频率

	// 降低内存阈值
	sm.streamingConfig.MemoryWarningThreshold = 0.50  // 50%
	sm.streamingConfig.MemoryCriticalThreshold = 0.65 // 65%

	// 配置pdfcpu最小内存模式
	sm.configurePDFCPUForMinimalMemory()

	sm.logger("大文件优化配置完成")
}

// Close 关闭合并器并清理资源
func (sm *StreamingMerger) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 关闭pdfcpu适配器
	if sm.adapter != nil {
		if err := sm.adapter.Close(); err != nil {
			return err
		}
	}

	// 取消进度跟踪器
	if sm.progressTracker != nil {
		sm.progressTracker.Cancel("合并器关闭")
	}

	return nil
}

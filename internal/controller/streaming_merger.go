package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

// StreamingMerger 流式PDF合并器
type StreamingMerger struct {
	controller    *Controller
	chunkSize     int64
	maxMemory     int64
	tempFiles     []string
	tempMutex     sync.Mutex
}

// NewStreamingMerger 创建新的流式合并器
func NewStreamingMerger(controller *Controller) *StreamingMerger {
	return &StreamingMerger{
		controller: controller,
		chunkSize:  1024 * 1024, // 1MB chunks
		maxMemory:  controller.Config.MaxMemoryUsage,
		tempFiles:  make([]string, 0),
	}
}

// MergeStreaming 执行流式合并
func (sm *StreamingMerger) MergeStreaming(ctx context.Context, job *model.MergeJob, 
	progressWriter io.Writer) error {
	
	defer sm.cleanup()
	
	// 检查内存使用情况
	if !sm.shouldUseStreaming() {
		// 内存充足，使用标准合并
		return sm.controller.PDFService.MergePDFs(job.MainFile, job.AdditionalFiles, 
			job.OutputPath, progressWriter)
	}
	
	// 执行流式合并
	return sm.executeStreamingMerge(ctx, job, progressWriter)
}

// executeStreamingMerge 执行流式合并的核心逻辑
func (sm *StreamingMerger) executeStreamingMerge(ctx context.Context, job *model.MergeJob, 
	progressWriter io.Writer) error {
	
	allFiles := append([]string{job.MainFile}, job.AdditionalFiles...)
	totalFiles := len(allFiles)
	
	// 第一阶段：预处理文件，创建临时分块
	sm.notifyProgress(0.1, "预处理文件", "正在分析文件结构...")
	
	processedFiles := make([]string, 0, totalFiles)
	
	for i, filePath := range allFiles {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		progress := 0.1 + (0.3 * float64(i) / float64(totalFiles))
		sm.notifyProgress(progress, "预处理文件", 
			fmt.Sprintf("正在处理: %s (%d/%d)", filepath.Base(filePath), i+1, totalFiles))
		
		// 预处理单个文件
		processedFile, err := sm.preprocessFile(ctx, filePath)
		if err != nil {
			return fmt.Errorf("预处理文件 %s 失败: %v", filepath.Base(filePath), err)
		}
		
		processedFiles = append(processedFiles, processedFile)
		
		// 写入进度
		if progressWriter != nil {
			progressWriter.Write([]byte(fmt.Sprintf("processed:%s\n", filePath)))
		}
	}
	
	// 第二阶段：流式合并
	sm.notifyProgress(0.4, "流式合并", "开始流式合并处理...")
	
	err := sm.performStreamingMerge(ctx, processedFiles, job.OutputPath, progressWriter)
	if err != nil {
		return fmt.Errorf("流式合并失败: %v", err)
	}
	
	// 第三阶段：后处理和优化
	sm.notifyProgress(0.9, "后处理", "正在优化输出文件...")
	
	if err := sm.postProcessOutput(ctx, job.OutputPath); err != nil {
		return fmt.Errorf("后处理失败: %v", err)
	}
	
	return nil
}

// preprocessFile 预处理单个文件
func (sm *StreamingMerger) preprocessFile(ctx context.Context, filePath string) (string, error) {
	// 检查文件是否需要预处理
	fileInfo, err := sm.controller.FileManager.GetFileInfo(filePath)
	if err != nil {
		return "", err
	}
	
	// 如果文件较小，直接返回原文件
	if fileInfo.Size < sm.chunkSize*2 {
		return filePath, nil
	}
	
	// 创建临时文件进行预处理
	tempFile, err := sm.createTempFile("preprocessed_", ".pdf")
	if err != nil {
		return "", err
	}
	
	// 这里应该实现实际的PDF预处理逻辑
	// 目前简单地复制文件作为占位符
	if err := sm.controller.FileManager.CopyFile(filePath, tempFile); err != nil {
		return "", err
	}
	
	return tempFile, nil
}

// performStreamingMerge 执行流式合并
func (sm *StreamingMerger) performStreamingMerge(ctx context.Context, files []string, 
	outputPath string, progressWriter io.Writer) error {
	
	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer outputFile.Close()
	
	// 写入PDF头部
	if err := sm.writePDFHeader(outputFile); err != nil {
		return fmt.Errorf("写入PDF头部失败: %v", err)
	}
	
	totalFiles := len(files)
	
	// 逐个处理文件
	for i, filePath := range files {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		progress := 0.4 + (0.5 * float64(i) / float64(totalFiles))
		sm.notifyProgress(progress, "合并文件", 
			fmt.Sprintf("正在合并: %s (%d/%d)", filepath.Base(filePath), i+1, totalFiles))
		
		// 流式处理单个文件
		if err := sm.streamFile(ctx, filePath, outputFile); err != nil {
			return fmt.Errorf("处理文件 %s 失败: %v", filepath.Base(filePath), err)
		}
		
		// 写入进度
		if progressWriter != nil {
			progressWriter.Write([]byte(fmt.Sprintf("merged:%s\n", filePath)))
		}
		
		// 定期检查内存使用情况
		if i%5 == 0 {
			runtime.GC() // 触发垃圾回收
		}
	}
	
	// 写入PDF尾部
	if err := sm.writePDFFooter(outputFile); err != nil {
		return fmt.Errorf("写入PDF尾部失败: %v", err)
	}
	
	return nil
}

// streamFile 流式处理单个文件
func (sm *StreamingMerger) streamFile(ctx context.Context, filePath string, 
	outputFile *os.File) error {
	
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer inputFile.Close()
	
	// 分块读取和写入
	buffer := make([]byte, sm.chunkSize)
	
	for {
		// 检查取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// 读取数据块
		n, err := inputFile.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取文件失败: %v", err)
		}
		
		// 写入数据块
		if _, err := outputFile.Write(buffer[:n]); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
		
		// 检查内存使用情况
		if sm.isMemoryHigh() {
			runtime.GC()
			time.Sleep(10 * time.Millisecond) // 短暂暂停以释放内存
		}
	}
	
	return nil
}

// postProcessOutput 后处理输出文件
func (sm *StreamingMerger) postProcessOutput(ctx context.Context, outputPath string) error {
	// 验证输出文件
	if err := sm.controller.ValidateFile(outputPath); err != nil {
		return fmt.Errorf("输出文件验证失败: %v", err)
	}
	
	// 获取文件信息
	fileInfo, err := sm.controller.FileManager.GetFileInfo(outputPath)
	if err != nil {
		return fmt.Errorf("获取输出文件信息失败: %v", err)
	}
	
	sm.notifyProgress(0.95, "验证输出", 
		fmt.Sprintf("输出文件大小: %.2f MB", float64(fileInfo.Size)/(1024*1024)))
	
	return nil
}

// 辅助方法

// shouldUseStreaming 判断是否应该使用流式处理
func (sm *StreamingMerger) shouldUseStreaming() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	currentMemory := int64(m.Alloc)
	return currentMemory > sm.maxMemory*70/100 // 超过70%使用流式处理
}

// isMemoryHigh 检查内存使用是否过高
func (sm *StreamingMerger) isMemoryHigh() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	currentMemory := int64(m.Alloc)
	return currentMemory > sm.maxMemory*90/100 // 超过90%认为内存过高
}

// createTempFile 创建临时文件
func (sm *StreamingMerger) createTempFile(prefix, suffix string) (string, error) {
	tempFile, _, err := sm.controller.FileManager.CreateTempFileWithPrefix(prefix, suffix)
	if err != nil {
		return "", err
	}
	
	sm.tempMutex.Lock()
	sm.tempFiles = append(sm.tempFiles, tempFile)
	sm.tempMutex.Unlock()
	
	return tempFile, nil
}

// cleanup 清理临时文件
func (sm *StreamingMerger) cleanup() {
	sm.tempMutex.Lock()
	defer sm.tempMutex.Unlock()
	
	for _, tempFile := range sm.tempFiles {
		sm.controller.FileManager.RemoveTempFile(tempFile)
	}
	
	sm.tempFiles = sm.tempFiles[:0] // 清空切片
}

// notifyProgress 通知进度更新
func (sm *StreamingMerger) notifyProgress(progress float64, status, detail string) {
	sm.controller.notifyProgress(progress, status, detail)
}

// writePDFHeader 写入PDF头部
func (sm *StreamingMerger) writePDFHeader(outputFile *os.File) error {
	// 这里应该写入实际的PDF头部
	// 目前使用简单的占位符
	header := "%PDF-1.4\n"
	_, err := outputFile.WriteString(header)
	return err
}

// writePDFFooter 写入PDF尾部
func (sm *StreamingMerger) writePDFFooter(outputFile *os.File) error {
	// 这里应该写入实际的PDF尾部
	// 目前使用简单的占位符
	footer := "%%EOF\n"
	_, err := outputFile.WriteString(footer)
	return err
}

// BatchProcessor 批处理器，用于处理大量文件
type BatchProcessor struct {
	streamingMerger *StreamingMerger
	batchSize       int
	maxConcurrency  int
}

// NewBatchProcessor 创建新的批处理器
func NewBatchProcessor(streamingMerger *StreamingMerger) *BatchProcessor {
	return &BatchProcessor{
		streamingMerger: streamingMerger,
		batchSize:       10,
		maxConcurrency:  2,
	}
}

// ProcessBatch 批量处理文件
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, files []string, 
	outputPath string, progressWriter io.Writer) error {
	
	if len(files) <= bp.batchSize {
		// 文件数量较少，直接处理
		job := &model.MergeJob{
			MainFile:        files[0],
			AdditionalFiles: files[1:],
			OutputPath:      outputPath,
		}
		return bp.streamingMerger.MergeStreaming(ctx, job, progressWriter)
	}
	
	// 分批处理
	batches := bp.createBatches(files)
	tempOutputs := make([]string, len(batches))
	
	// 并发处理各个批次
	semaphore := make(chan struct{}, bp.maxConcurrency)
	errChan := make(chan error, len(batches))
	
	for i, batch := range batches {
		go func(batchIndex int, batchFiles []string) {
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			tempOutput, err := bp.streamingMerger.createTempFile(
				fmt.Sprintf("batch_%d_", batchIndex), ".pdf")
			if err != nil {
				errChan <- err
				return
			}
			
			tempOutputs[batchIndex] = tempOutput
			
			job := &model.MergeJob{
				MainFile:        batchFiles[0],
				AdditionalFiles: batchFiles[1:],
				OutputPath:      tempOutput,
			}
			
			err = bp.streamingMerger.MergeStreaming(ctx, job, progressWriter)
			errChan <- err
		}(i, batch)
	}
	
	// 等待所有批次完成
	for i := 0; i < len(batches); i++ {
		if err := <-errChan; err != nil {
			return fmt.Errorf("批次 %d 处理失败: %v", i, err)
		}
	}
	
	// 合并所有批次的输出
	finalJob := &model.MergeJob{
		MainFile:        tempOutputs[0],
		AdditionalFiles: tempOutputs[1:],
		OutputPath:      outputPath,
	}
	
	return bp.streamingMerger.MergeStreaming(ctx, finalJob, progressWriter)
}

// createBatches 创建文件批次
func (bp *BatchProcessor) createBatches(files []string) [][]string {
	batches := make([][]string, 0)
	
	for i := 0; i < len(files); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(files) {
			end = len(files)
		}
		batches = append(batches, files[i:end])
	}
	
	return batches
}
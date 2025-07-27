//go:build ignore
// +build ignore
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("🚀 PDF流式合并优化演示")
	fmt.Println("========================")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "streaming_demo")
	if err != nil {
		fmt.Printf("❌ 创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("📁 临时目录: %s\n", tempDir)

	// 创建演示文件
	files := createDemoFiles(tempDir)
	if len(files) == 0 {
		fmt.Println("❌ 无法创建演示文件")
		return
	}

	fmt.Printf("📄 创建了 %d 个演示文件\n", len(files))

	// 演示1: 基础流式合并
	demonstrateBasicStreaming(files, tempDir)

	// 演示2: 不同配置的流式合并
	demonstrateConfigurations(files, tempDir)

	// 演示3: 内存优化功能
	demonstrateMemoryOptimization(files, tempDir)

	// 演示4: 并发处理
	demonstrateConcurrentProcessing(files, tempDir)

	// 演示5: 自适应分块
	demonstrateAdaptiveChunking(files, tempDir)

	fmt.Println("\n✅ 演示完成！")
}

func createDemoFiles(tempDir string) []string {
	files := make([]string, 0)
	
	// 创建一些简单的文本文件作为演示（实际应用中应该是PDF文件）
	for i := 1; i <= 8; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("demo_%d.txt", i))
		content := fmt.Sprintf("演示文件 %d\n创建时间: %s\n内容: 这是一个演示文件，用于测试流式合并优化功能。", 
			i, time.Now().Format("2006-01-02 15:04:05"))
		
		if err := os.WriteFile(filename, []byte(content), 0644); err == nil {
			files = append(files, filename)
		}
	}
	
	return files
}

func demonstrateBasicStreaming(files []string, tempDir string) {
	fmt.Println("\n1. 基础流式合并演示...")

	// 创建标准配置的流式合并器
	options := &pdf.MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024, // 100MB
		TempDirectory:  tempDir,
		EnableGC:       true,
		UseStreaming:   true,
		OptimizeMemory: true,
	}

	merger := pdf.NewStreamingMerger(options)
	defer merger.Close()

	outputPath := filepath.Join(tempDir, "basic_output.txt")
	
	fmt.Printf("  输入文件: %d 个\n", len(files))
	fmt.Printf("  输出文件: %s\n", outputPath)
	fmt.Printf("  内存限制: %d MB\n", options.MaxMemoryUsage/(1024*1024))

	// 执行合并
	ctx := context.Background()
	startTime := time.Now()
	
	result, err := merger.MergeStreaming(ctx, files, outputPath, func(progress float64, message string) {
		fmt.Printf("    进度: %.1f%% - %s\n", progress, message)
	})

	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("  合并失败: %v\n", err)
	} else {
		fmt.Printf("  合并成功: 耗时=%v, 处理文件=%d, 内存使用=%s\n", 
			duration, result.ProcessedFiles, formatSize(result.MemoryUsage))
	}
}

func demonstrateConfigurations(files []string, tempDir string) {
	fmt.Println("\n2. 不同配置的流式合并演示...")

	configurations := []struct {
		name   string
		config *pdf.MergeOptions
	}{
		{
			"标准配置",
			&pdf.MergeOptions{
				MaxMemoryUsage: 100 * 1024 * 1024, // 100MB
				TempDirectory:  tempDir,
				EnableGC:       true,
			},
		},
		{
			"内存受限配置",
			&pdf.MergeOptions{
				MaxMemoryUsage: 10 * 1024 * 1024, // 10MB
				TempDirectory:  tempDir,
				EnableGC:       true,
				OptimizeMemory: true,
			},
		},
		{
			"高性能配置",
			&pdf.MergeOptions{
				MaxMemoryUsage:    200 * 1024 * 1024, // 200MB
				TempDirectory:     tempDir,
				EnableGC:          false,
				ConcurrentWorkers: 4,
			},
		},
	}

	for i, config := range configurations {
		fmt.Printf("\n  配置 %d: %s\n", i+1, config.name)
		
		// 创建流式合并器
		merger := pdf.NewStreamingMerger(config.config)
		
		// 执行合并
		outputPath := filepath.Join(tempDir, fmt.Sprintf("output_%d.txt", i+1))
		
		startTime := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		result, err := merger.MergeStreaming(ctx, files, outputPath, func(progress float64, message string) {
			fmt.Printf("    进度: %.1f%% - %s\n", progress, message)
		})
		
		cancel()
		
		if err != nil {
			fmt.Printf("    合并失败: %v\n", err)
		} else {
			duration := time.Since(startTime)
			fmt.Printf("    合并成功: 耗时=%v, 处理文件=%d, 内存使用=%s\n", 
				duration, result.ProcessedFiles, formatSize(result.MemoryUsage))
		}
		
		merger.Close()
	}
}

func demonstrateMemoryOptimization(files []string, tempDir string) {
	fmt.Println("\n3. 内存优化功能演示...")

	// 创建内存受限的合并器
	merger := pdf.NewStreamingMerger(&pdf.MergeOptions{
		MaxMemoryUsage: 20 * 1024 * 1024, // 20MB限制
		TempDirectory:  tempDir,
		OptimizeMemory: true,
	})
	defer merger.Close()

	// 创建内存监控器
	monitor := pdf.NewMemoryMonitor(20 * 1024 * 1024)

	fmt.Println("  检查内存压力...")
	pressure := monitor.CheckMemoryPressure()
	fmt.Printf("  当前内存压力级别: %v\n", pressure)

	// 分析文件特征
	fmt.Println("  分析文件特征...")
	totalSize := int64(0)
	largeFileCount := 0
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
			if info.Size() > 5*1024*1024 { // 5MB以上算大文件
				largeFileCount++
			}
		}
	}

	fmt.Printf("  文件总大小: %s\n", formatSize(totalSize))
	fmt.Printf("  大文件数量: %d\n", largeFileCount)

	// 执行优化合并
	outputPath := filepath.Join(tempDir, "optimized_output.txt")
	
	fmt.Println("  执行内存优化合并...")
	startTime := time.Now()
	ctx := context.Background()
	
	result, err := merger.MergeStreaming(ctx, files, outputPath, func(progress float64, message string) {
		fmt.Printf("    进度: %.1f%% - %s\n", progress, message)
	})
	
	if err != nil {
		fmt.Printf("  优化合并失败: %v\n", err)
		return
	}
	
	duration := time.Since(startTime)
	fmt.Printf("  优化合并成功: 耗时=%v, 内存使用=%s\n", 
		duration, formatSize(result.MemoryUsage))
}

func demonstrateConcurrentProcessing(files []string, tempDir string) {
	fmt.Println("\n4. 并发处理演示...")

	if runtime.NumCPU() < 2 {
		fmt.Println("  跳过并发演示：系统只有一个CPU核心")
		return
	}

	// 创建支持并发的配置
	options := &pdf.MergeOptions{
		MaxMemoryUsage:    100 * 1024 * 1024, // 100MB
		TempDirectory:     tempDir,
		ConcurrentWorkers: runtime.NumCPU(),
		UseStreaming:      true,
	}

	merger := pdf.NewStreamingMerger(options)
	defer merger.Close()

	outputPath := filepath.Join(tempDir, "concurrent_output.txt")
	
	fmt.Printf("  系统CPU核心数: %d\n", runtime.NumCPU())
	fmt.Printf("  并发工作线程数: %d\n", options.ConcurrentWorkers)

	// 检查是否应该使用并发处理
	shouldUseConcurrent := merger.ShouldUseConcurrentProcessing(files)
	fmt.Printf("  是否应该使用并发处理: %v\n", shouldUseConcurrent)

	// 执行并发合并
	ctx := context.Background()
	startTime := time.Now()
	
	result, err := merger.MergeStreaming(ctx, files, outputPath, func(progress float64, message string) {
		fmt.Printf("    进度: %.1f%% - %s\n", progress, message)
	})
	
	if err != nil {
		fmt.Printf("  并发合并失败: %v\n", err)
	} else {
		duration := time.Since(startTime)
		fmt.Printf("  并发合并成功: 耗时=%v, 处理文件=%d\n", 
			duration, result.ProcessedFiles)
	}
}

func demonstrateAdaptiveChunking(files []string, tempDir string) {
	fmt.Println("\n5. 自适应分块演示...")

	// 创建支持自适应分块的配置
	streamingConfig := pdf.DefaultStreamingConfig()
	streamingConfig.EnableAdaptiveChunking = true
	streamingConfig.EnableMemoryPrediction = true
	streamingConfig.MinChunkSize = 2
	streamingConfig.MaxChunkSize = 10

	options := &pdf.MergeOptions{
		MaxMemoryUsage: 50 * 1024 * 1024, // 50MB
		TempDirectory:  tempDir,
		UseStreaming:   true,
	}

	merger := pdf.NewStreamingMergerWithConfig(options, streamingConfig)
	defer merger.Close()

	// 分析文件特征
	fmt.Println("  分析文件特征...")
	analysis := merger.AnalyzeFiles(files)
	
	fmt.Printf("  文件总数: %d\n", analysis.FileCount)
	fmt.Printf("  总大小: %s\n", formatSize(analysis.TotalSize))
	fmt.Printf("  平均大小: %s\n", formatSize(analysis.AvgSize))
	fmt.Printf("  最大文件: %s\n", formatSize(analysis.MaxSize))
	fmt.Printf("  最小文件: %s\n", formatSize(analysis.MinSize))
	fmt.Printf("  包含大文件: %v\n", analysis.HasLargeFiles)

	// 计算最优分块大小
	chunkSize := merger.CalculateOptimalChunkSize(files)
	fmt.Printf("  计算的最优分块大小: %d\n", chunkSize)

	// 执行自适应分块合并
	outputPath := filepath.Join(tempDir, "adaptive_output.txt")
	
	fmt.Println("  执行自适应分块合并...")
	ctx := context.Background()
	startTime := time.Now()
	
	result, err := merger.MergeStreaming(ctx, files, outputPath, func(progress float64, message string) {
		fmt.Printf("    进度: %.1f%% - %s\n", progress, message)
	})
	
	if err != nil {
		fmt.Printf("  自适应分块合并失败: %v\n", err)
	} else {
		duration := time.Since(startTime)
		fmt.Printf("  自适应分块合并成功: 耗时=%v, 处理文件=%d\n", 
			duration, result.ProcessedFiles)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
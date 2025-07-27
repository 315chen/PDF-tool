//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF流式合并示例 ===")

	// 检查命令行参数
	if len(os.Args) < 4 {
		fmt.Println("用法: go run streaming_merger_example.go <主文件> <输出文件> <附加文件1> [附加文件2] ...")
		fmt.Println("示例: go run streaming_merger_example.go main.pdf output.pdf file1.pdf file2.pdf")
		os.Exit(1)
	}

	mainFile := os.Args[1]
	outputFile := os.Args[2]
	additionalFiles := os.Args[3:]

	// 验证输入文件存在
	if !fileExists(mainFile) {
		fmt.Printf("错误: 主文件不存在: %s\n", mainFile)
		os.Exit(1)
	}

	for _, file := range additionalFiles {
		if !fileExists(file) {
			fmt.Printf("警告: 附加文件不存在: %s\n", file)
		}
	}

	// 创建流式合并器
	options := &pdf.MergeOptions{
		MaxMemoryUsage: 100 * 1024 * 1024, // 100MB内存限制
		TempDirectory:  os.TempDir(),
		EnableGC:       true,
		ChunkSize:      10, // 每次处理10页
	}

	merger := pdf.NewStreamingMerger(options)

	fmt.Printf("开始合并PDF文件...\n")
	fmt.Printf("主文件: %s\n", mainFile)
	fmt.Printf("附加文件: %v\n", additionalFiles)
	fmt.Printf("输出文件: %s\n", outputFile)
	fmt.Printf("内存限制: %.2f MB\n", float64(options.MaxMemoryUsage)/(1024*1024))
	fmt.Println()

	// 执行合并
	result, err := merger.MergeFiles(mainFile, additionalFiles, outputFile, os.Stdout)
	
	if err != nil {
		fmt.Printf("合并失败: %v\n", err)
		os.Exit(1)
	}

	// 显示结果
	fmt.Println("\n=== 合并完成 ===")
	fmt.Printf("输出文件: %s\n", result.OutputPath)
	fmt.Printf("总页数: %d\n", result.TotalPages)
	fmt.Printf("处理文件数: %d\n", result.ProcessedFiles)
	fmt.Printf("跳过文件数: %d\n", len(result.SkippedFiles))
	fmt.Printf("处理时间: %v\n", result.ProcessingTime)
	fmt.Printf("内存使用: %.2f MB\n", float64(result.MemoryUsage)/(1024*1024))

	if len(result.SkippedFiles) > 0 {
		fmt.Println("\n跳过的文件:")
		for _, file := range result.SkippedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	// 验证输出文件
	if fileExists(result.OutputPath) {
		fileInfo, err := os.Stat(result.OutputPath)
		if err == nil {
			fmt.Printf("\n输出文件大小: %.2f MB\n", float64(fileInfo.Size())/(1024*1024))
		}
	}

	fmt.Println("\n合并成功完成！")
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
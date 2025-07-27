//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF写入和输出管理示例 ===")

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run pdf_writer_example.go <输出文件路径>")
		fmt.Println("示例: go run pdf_writer_example.go output.pdf")
		os.Exit(1)
	}

	outputPath := os.Args[1]

	// 演示输出路径管理
	demonstrateOutputManager(outputPath)

	// 演示PDF写入器
	demonstratePDFWriter(outputPath)
}

func demonstrateOutputManager(requestedPath string) {
	fmt.Println("\n=== 输出路径管理演示 ===")

	// 创建输出管理器
	manager := pdf.NewOutputManager(&pdf.OutputOptions{
		BaseDirectory:   ".",
		DefaultFileName: "default_output.pdf",
		AutoIncrement:   true,
		TimestampSuffix: false,
		BackupEnabled:   true,
	})

	// 解析输出路径
	outputInfo, err := manager.ResolveOutputPath(requestedPath)
	if err != nil {
		fmt.Printf("解析输出路径失败: %v\n", err)
		return
	}

	fmt.Printf("原始路径: %s\n", outputInfo.OriginalPath)
	fmt.Printf("最终路径: %s\n", outputInfo.FinalPath)
	fmt.Printf("是否递增: %t\n", outputInfo.IsIncremented)
	fmt.Printf("是否有时间戳: %t\n", outputInfo.HasTimestamp)
	if outputInfo.BackupPath != "" {
		fmt.Printf("备份路径: %s\n", outputInfo.BackupPath)
	}

	// 演示建议路径功能
	inputFiles := []string{"document1.pdf", "document2.pdf", "document3.pdf"}
	suggestedPath := manager.GetSuggestedPath(inputFiles)
	fmt.Printf("建议路径: %s\n", suggestedPath)

	// 验证路径
	if err := manager.ValidateOutputPath(outputInfo.FinalPath); err != nil {
		fmt.Printf("路径验证失败: %v\n", err)
	} else {
		fmt.Println("路径验证通过")
	}
}

func demonstratePDFWriter(outputPath string) {
	fmt.Println("\n=== PDF写入器演示 ===")

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "pdf_writer_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 使用临时目录中的输出路径
	demoOutputPath := filepath.Join(tempDir, "demo_output.pdf")

	// 创建PDF写入器
	writer, err := pdf.NewPDFWriter(demoOutputPath, &pdf.WriterOptions{
		MaxRetries:    3,
		RetryDelay:    1000000000, // 1秒
		BackupEnabled: true,
		TempDirectory: tempDir,
	})
	if err != nil {
		fmt.Printf("创建PDF写入器失败: %v\n", err)
		return
	}

	fmt.Printf("输出路径: %s\n", writer.GetOutputPath())
	fmt.Printf("临时路径: %s\n", writer.GetTempPath())

	// 打开写入器
	if err := writer.Open(); err != nil {
		fmt.Printf("打开写入器失败: %v\n", err)
		return
	}
	defer writer.Close()

	fmt.Printf("写入器状态: 已打开 = %t\n", writer.IsOpen())

	// 注意：由于unidoc许可证限制，实际写入可能会失败
	// 但我们可以演示写入器的结构和错误处理
	fmt.Println("\n尝试写入PDF文件...")
	result, err := writer.Write(os.Stdout)

	if err != nil {
		fmt.Printf("写入失败: %v\n", err)
		if result != nil {
			fmt.Printf("重试次数: %d\n", result.RetryCount)
			fmt.Printf("写入时间: %v\n", result.WriteTime)
		}
	} else {
		fmt.Println("写入成功！")
		fmt.Printf("输出文件: %s\n", result.OutputPath)
		fmt.Printf("文件大小: %d 字节\n", result.FileSize)
		fmt.Printf("写入时间: %v\n", result.WriteTime)
		fmt.Printf("重试次数: %d\n", result.RetryCount)
		
		if result.BackupPath != "" {
			fmt.Printf("备份文件: %s\n", result.BackupPath)
		}
	}
}

func demonstrateErrorHandling() {
	fmt.Println("\n=== 错误处理演示 ===")

	// 演示各种错误情况
	errorCases := []struct {
		name        string
		outputPath  string
		description string
	}{
		{
			name:        "无效扩展名",
			outputPath:  "output.txt",
			description: "尝试使用非PDF扩展名",
		},
		{
			name:        "空路径",
			outputPath:  "",
			description: "尝试使用空输出路径",
		},
		{
			name:        "无效目录",
			outputPath:  "/invalid/path/output.pdf",
			description: "尝试使用无效目录",
		},
	}

	for _, errorCase := range errorCases {
		fmt.Printf("\n测试场景: %s\n", errorCase.name)
		fmt.Printf("描述: %s\n", errorCase.description)
		fmt.Printf("路径: %s\n", errorCase.outputPath)

		_, err := pdf.NewPDFWriter(errorCase.outputPath, nil)
		if err != nil {
			fmt.Printf("结果: 正确捕获错误 - %v\n", err)
		} else {
			fmt.Println("结果: 意外成功")
		}
	}
}

func demonstrateRetryMechanism() {
	fmt.Println("\n=== 重试机制演示 ===")

	tempDir := filepath.Join(os.TempDir(), "retry_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "retry_test.pdf")

	// 创建具有重试机制的写入器
	writer, err := pdf.NewPDFWriter(outputPath, &pdf.WriterOptions{
		MaxRetries:    5,
		RetryDelay:    500000000, // 0.5秒
		BackupEnabled: false,
		TempDirectory: tempDir,
	})
	if err != nil {
		fmt.Printf("创建写入器失败: %v\n", err)
		return
	}

	fmt.Printf("最大重试次数: 5\n")
	fmt.Printf("重试延迟: 0.5秒\n")
	fmt.Printf("输出路径: %s\n", outputPath)

	if err := writer.Open(); err != nil {
		fmt.Printf("打开写入器失败: %v\n", err)
		return
	}
	defer writer.Close()

	// 尝试写入（可能会因为许可证问题失败，但会演示重试机制）
	fmt.Println("\n开始写入（将演示重试机制）...")
	result, err := writer.Write(os.Stdout)

	if result != nil {
		fmt.Printf("实际重试次数: %d\n", result.RetryCount)
		fmt.Printf("总写入时间: %v\n", result.WriteTime)
		fmt.Printf("写入成功: %t\n", result.Success)
	}

	if err != nil {
		fmt.Printf("最终错误: %v\n", err)
	}
}
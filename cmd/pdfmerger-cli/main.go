package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/pdf-merger/internal/controller"
	"github.com/user/pdf-merger/internal/model"
	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

var (
	Version   = "v1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	var (
		inputFiles  = flag.String("input", "", "输入PDF文件路径，用逗号分隔")
		outputFile  = flag.String("output", "merged.pdf", "输出PDF文件路径")
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("PDF合并工具 (命令行版本) %s\n", Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
		fmt.Printf("Git提交: %s\n", GitCommit)
		return
	}

	if *showHelp || *inputFiles == "" {
		showUsage()
		return
	}

	// 解析输入文件
	files := strings.Split(*inputFiles, ",")
	for i, file := range files {
		files[i] = strings.TrimSpace(file)
	}

	if len(files) < 2 {
		fmt.Println("错误: 至少需要两个PDF文件进行合并")
		os.Exit(1)
	}

	// 验证输入文件
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("错误: 文件不存在: %s\n", file)
			os.Exit(1)
		}
	}

	// 创建输出目录
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("错误: 无法创建输出目录: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("开始合并 %d 个PDF文件...\n", len(files))
	fmt.Printf("输出文件: %s\n", *outputFile)
	fmt.Println()

	// 执行合并
	if err := mergePDFs(files, *outputFile); err != nil {
		fmt.Printf("合并失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ PDF合并完成！")
}

func showUsage() {
	fmt.Println("PDF合并工具 (命令行版本)")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  pdf-merger-cli -input file1.pdf,file2.pdf,file3.pdf -output merged.pdf")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -input   输入PDF文件路径，用逗号分隔 (必需)")
	fmt.Println("  -output  输出PDF文件路径 (默认: merged.pdf)")
	fmt.Println("  -version 显示版本信息")
	fmt.Println("  -help    显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  pdf-merger-cli -input doc1.pdf,doc2.pdf -output combined.pdf")
	fmt.Println("  pdf-merger-cli -input *.pdf -output all.pdf")
	fmt.Println("  pdf-merger-cli -version")
}

func mergePDFs(inputFiles []string, outputFile string) error {
	// 创建配置
	config := model.DefaultConfig()

	// 创建PDF服务
	pdfService := pdf.NewPDFService()

	// 创建文件管理器
	fileManager := file.NewFileManager(config.TempDirectory)

	// 创建控制器
	ctrl := controller.NewController(pdfService, fileManager, config)

	// 设置进度回调
	ctrl.SetProgressCallback(func(progress float64, status, detail string) {
		percentage := int(progress * 100)
		fmt.Printf("\r进度: %d%% - %s: %s", percentage, status, detail)
		if progress >= 1.0 {
			fmt.Println()
		}
	})

	// 设置错误回调
	errorChan := make(chan error, 1)
	ctrl.SetErrorCallback(func(err error) {
		errorChan <- err
	})

	// 设置完成回调
	completionChan := make(chan string, 1)
	ctrl.SetCompletionCallback(func(outputPath string) {
		completionChan <- outputPath
	})

	// 验证文件
	for _, file := range inputFiles {
		if err := ctrl.ValidateFile(file); err != nil {
			return fmt.Errorf("文件验证失败 %s: %v", file, err)
		}
	}

	// 启动合并任务 (主文件 + 附加文件)
	mainFile := inputFiles[0]
	additionalFiles := inputFiles[1:]

	if err := ctrl.StartMergeJob(mainFile, additionalFiles, outputFile); err != nil {
		return err
	}

	// 等待结果
	select {
	case err := <-errorChan:
		return err
	case outputPath := <-completionChan:
		fmt.Printf("合并完成，输出文件: %s\n", outputPath)
		return nil
	}
}

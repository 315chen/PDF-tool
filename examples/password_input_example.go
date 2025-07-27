//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/user/pdf-merger/internal/ui"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF密码输入处理示例 ===")

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  go run password_input_example.go console <PDF文件路径>  # 控制台模式")
		fmt.Println("  go run password_input_example.go gui <PDF文件路径>      # GUI模式")
		fmt.Println("  go run password_input_example.go batch               # 批量密码演示")
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "console":
		if len(os.Args) < 3 {
			fmt.Println("控制台模式需要指定PDF文件路径")
			os.Exit(1)
		}
		demonstrateConsoleMode(os.Args[2])
	case "gui":
		if len(os.Args) < 3 {
			fmt.Println("GUI模式需要指定PDF文件路径")
			os.Exit(1)
		}
		demonstrateGUIMode(os.Args[2])
	case "batch":
		demonstrateBatchMode()
	default:
		fmt.Printf("未知模式: %s\n", mode)
		os.Exit(1)
	}
}

func demonstrateConsoleMode(pdfFile string) {
	fmt.Printf("\n=== 控制台密码输入演示 ===\n")
	fmt.Printf("目标文件: %s\n", pdfFile)

	// 验证文件存在
	if !fileExists(pdfFile) {
		fmt.Printf("错误: 文件不存在: %s\n", pdfFile)
		os.Exit(1)
	}

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "password_input_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
	})
	defer decryptor.CleanupTempFiles()

	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		MaxCacheSize:   10,
		PasswordPrompt: pdf.CreateConsolePasswordPrompt(os.Stdout),
		ValidationFunc: pdf.CreateValidationFunc(decryptor),
	})

	// 演示密码输入处理
	fmt.Println("\n开始密码输入处理...")
	result, err := passwordManager.GetPasswordForFile(pdfFile)

	// 显示结果
	fmt.Println("\n=== 处理结果 ===")
	if err != nil {
		fmt.Printf("处理失败: %v\n", err)
	}

	if result != nil {
		fmt.Printf("成功: %t\n", result.Success)
		fmt.Printf("密码: %s\n", result.Password)
		fmt.Printf("来自缓存: %t\n", result.FromCache)
		fmt.Printf("尝试次数: %d\n", result.AttemptCount)
		fmt.Printf("用户取消: %t\n", result.UserCanceled)

		if result.Success {
			fmt.Println("密码已缓存，下次访问同一文件将自动使用")
		}
	}

	// 演示缓存功能
	if result != nil && result.Success {
		fmt.Println("\n=== 缓存功能演示 ===")
		fmt.Println("再次获取密码（应该来自缓存）...")

		result2, err := passwordManager.GetPasswordForFile(pdfFile)
		if err != nil {
			fmt.Printf("从缓存获取失败: %v\n", err)
		} else if result2.Success && result2.FromCache {
			fmt.Printf("成功从缓存获取密码: %s\n", result2.Password)
		}

		// 显示缓存信息
		fmt.Printf("缓存大小: %d\n", passwordManager.GetCacheSize())
		cachedFiles := passwordManager.GetCachedFiles()
		fmt.Printf("缓存文件数: %d\n", len(cachedFiles))
	}
}

func demonstrateGUIMode(pdfFile string) {
	fmt.Printf("\n=== GUI密码输入演示 ===\n")
	fmt.Printf("目标文件: %s\n", pdfFile)

	// 验证文件存在
	if !fileExists(pdfFile) {
		fmt.Printf("错误: 文件不存在: %s\n", pdfFile)
		os.Exit(1)
	}

	// 创建Fyne应用
	myApp := app.New()

	window := myApp.NewWindow("PDF密码输入演示")
	window.Resize(fyne.NewSize(600, 400))

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "password_input_gui_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
	})
	defer decryptor.CleanupTempFiles()

	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		MaxCacheSize:   10,
		PasswordPrompt: ui.CreateGUIPasswordPrompt(window),
		ValidationFunc: pdf.CreateValidationFunc(decryptor),
	})

	// 创建界面
	resultLabel := widget.NewLabel("点击按钮开始密码输入演示")
	resultLabel.Wrapping = fyne.TextWrapWord

	testButton := widget.NewButton("测试密码输入", func() {
		resultLabel.SetText("正在处理密码输入...")

		// 在goroutine中处理密码输入，避免阻塞UI
		go func() {
			result, err := passwordManager.GetPasswordForFile(pdfFile)

			// 更新UI（需要在主线程中执行）
			var resultText string
			if err != nil {
				resultText = fmt.Sprintf("处理失败: %v", err)
			} else if result != nil {
				resultText = fmt.Sprintf(
					"处理结果:\n成功: %t\n密码: %s\n来自缓存: %t\n尝试次数: %d\n用户取消: %t",
					result.Success, result.Password, result.FromCache,
					result.AttemptCount, result.UserCanceled,
				)
			}

			resultLabel.SetText(resultText)
		}()
	})

	cacheButton := widget.NewButton("管理密码缓存", func() {
		cachedFiles := passwordManager.GetCachedFiles()
		cacheDialog := ui.NewPasswordCacheDialog(window, cachedFiles, func(filePath string) {
			passwordManager.RemovePasswordFromCache(filePath)
			fmt.Printf("已清除文件密码缓存: %s\n", filepath.Base(filePath))
		})
		cacheDialog.Show()
	})

	batchButton := widget.NewButton("批量密码设置", func() {
		initialPasswords := []string{"password123", "123456", "admin"}
		batchDialog := ui.NewPasswordBatchDialog(window, initialPasswords)
		passwords := batchDialog.ShowAndWait()

		if passwords != nil {
			fmt.Printf("设置的批量密码: %v\n", passwords)
			resultLabel.SetText(fmt.Sprintf("设置了 %d 个批量密码", len(passwords)))
		} else {
			resultLabel.SetText("批量密码设置已取消")
		}
	})

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("文件: %s", filepath.Base(pdfFile))),
		testButton,
		cacheButton,
		batchButton,
		widget.NewSeparator(),
		resultLabel,
	)

	window.SetContent(content)
	window.ShowAndRun()
}

func demonstrateBatchMode() {
	fmt.Printf("\n=== 批量密码处理演示 ===\n")

	// 创建临时目录和测试文件
	tempDir := filepath.Join(os.TempDir(), "batch_password_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建多个测试文件
	testFiles := []string{
		filepath.Join(tempDir, "file1.pdf"),
		filepath.Join(tempDir, "file2.pdf"),
		filepath.Join(tempDir, "file3.pdf"),
	}

	for i, testFile := range testFiles {
		err := os.WriteFile(testFile, []byte(fmt.Sprintf("test content %d", i)), 0644)
		if err != nil {
			fmt.Printf("创建测试文件失败: %v\n", err)
			return
		}
	}

	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
	})
	defer decryptor.CleanupTempFiles()

	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		MaxCacheSize: 10,
		PasswordPrompt: func(filePath string, attempt int, lastError error) (string, bool) {
			// 模拟用户输入不同的密码
			fileName := filepath.Base(filePath)
			password := fmt.Sprintf("password_%s", fileName)
			fmt.Printf("为文件 %s 提供密码: %s\n", fileName, password)
			return password, true
		},
		ValidationFunc: func(filePath, password string) error {
			// 模拟验证总是成功
			return nil
		},
	})

	// 创建批量密码管理器
	batchPasswords := []string{"common1", "common2", "common3"}
	batchManager := pdf.NewBatchPasswordManager(passwordManager, batchPasswords)

	fmt.Printf("批量密码: %v\n", batchPasswords)
	fmt.Printf("测试文件: %d 个\n", len(testFiles))

	// 处理每个文件
	for i, testFile := range testFiles {
		fmt.Printf("\n--- 处理文件 %d: %s ---\n", i+1, filepath.Base(testFile))

		result, err := batchManager.ProcessFileWithBatch(testFile)
		if err != nil {
			fmt.Printf("处理失败: %v\n", err)
		} else if result != nil {
			fmt.Printf("处理成功: 密码=%s, 尝试次数=%d, 来自缓存=%t\n",
				result.Password, result.AttemptCount, result.FromCache)
		}
	}

	// 显示最终缓存状态
	fmt.Printf("\n=== 最终状态 ===\n")
	fmt.Printf("缓存大小: %d\n", passwordManager.GetCacheSize())
	cachedFiles := passwordManager.GetCachedFiles()
	fmt.Printf("缓存文件:\n")
	for _, file := range cachedFiles {
		fmt.Printf("  - %s\n", filepath.Base(file))
	}
}

func demonstratePasswordMemory() {
	fmt.Println("\n=== 密码记忆功能演示 ===")

	tempDir := filepath.Join(os.TempDir(), "password_memory_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "memory_test.pdf")
	err = os.WriteFile(testFile, []byte("test content for memory"), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}

	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		MaxCacheSize: 5,
	})

	// 模拟第一次输入密码
	testPassword := "remembered_password"
	passwordManager.SetPasswordPrompt(func(filePath string, attempt int, lastError error) (string, bool) {
		fmt.Printf("用户输入密码: %s\n", testPassword)
		return testPassword, true
	})

	passwordManager.SetValidationFunc(func(filePath, password string) error {
		if password == testPassword {
			return nil
		}
		return fmt.Errorf("密码错误")
	})

	// 第一次获取密码
	fmt.Println("第一次获取密码:")
	result1, err := passwordManager.GetPasswordForFile(testFile)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}

	fmt.Printf("成功: %t, 密码: %s, 来自缓存: %t\n",
		result1.Success, result1.Password, result1.FromCache)

	// 第二次获取密码（应该来自缓存）
	fmt.Println("\n第二次获取密码（应该来自缓存）:")
	result2, err := passwordManager.GetPasswordForFile(testFile)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}

	fmt.Printf("成功: %t, 密码: %s, 来自缓存: %t\n",
		result2.Success, result2.Password, result2.FromCache)

	// 验证密码记忆功能
	if result2.FromCache && result2.Password == testPassword {
		fmt.Println("✓ 密码记忆功能正常工作")
	} else {
		fmt.Println("✗ 密码记忆功能异常")
	}
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
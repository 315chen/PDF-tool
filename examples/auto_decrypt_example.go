//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF自动解密示例 ===")

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run auto_decrypt_example.go <PDF文件路径> [自定义密码1] [自定义密码2] ...")
		fmt.Println("示例: go run auto_decrypt_example.go encrypted.pdf mypassword 123456")
		os.Exit(1)
	}

	pdfFile := os.Args[1]
	customPasswords := os.Args[2:]

	// 验证输入文件存在
	if !fileExists(pdfFile) {
		fmt.Printf("错误: PDF文件不存在: %s\n", pdfFile)
		os.Exit(1)
	}

	// 演示自动解密功能
	demonstrateAutoDecrypt(pdfFile, customPasswords)

	// 演示密码管理功能
	demonstratePasswordManagement()

	// 演示临时文件管理
	demonstrateTempFileManagement()
}

func demonstrateAutoDecrypt(pdfFile string, customPasswords []string) {
	fmt.Printf("\n=== 自动解密演示 ===\n")
	fmt.Printf("目标文件: %s\n", pdfFile)

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "auto_decrypt_demo")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
		MaxAttempts:   20, // 限制尝试次数以加快演示
		AttemptDelay:  time.Millisecond * 100,
	})
	defer decryptor.CleanupTempFiles()

	// 如果提供了自定义密码，添加到常用密码列表前面
	if len(customPasswords) > 0 {
		fmt.Printf("添加自定义密码: %v\n", customPasswords)
		currentPasswords := decryptor.GetCommonPasswords()
		// 将自定义密码添加到列表前面
		newPasswords := append(customPasswords, currentPasswords...)
		decryptor.SetCommonPasswords(newPasswords)
	}

	fmt.Printf("最大尝试次数: %d\n", decryptor.GetMaxAttempts())
	fmt.Printf("尝试延迟: %v\n", decryptor.GetAttemptDelay())

	// 执行自动解密（带进度显示）
	fmt.Println("\n开始自动解密...")
	result, err := decryptor.DecryptWithProgress(pdfFile, os.Stdout)

	// 显示结果
	fmt.Println("\n=== 解密结果 ===")
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
	}

	if result != nil {
		fmt.Printf("成功: %t\n", result.Success)
		fmt.Printf("输出路径: %s\n", result.DecryptedPath)
		fmt.Printf("使用密码: %s\n", result.UsedPassword)
		fmt.Printf("尝试次数: %d\n", result.AttemptCount)
		fmt.Printf("处理时间: %v\n", result.ProcessingTime)
		fmt.Printf("原始文件: %t\n", result.IsOriginalFile)

		if result.Success && !result.IsOriginalFile {
			// 验证解密后的文件
			if fileExists(result.DecryptedPath) {
				fileInfo, err := os.Stat(result.DecryptedPath)
				if err == nil {
					fmt.Printf("解密文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
				}
			}
		}
	}

	// 显示临时文件信息
	tempFiles := decryptor.GetTempFiles()
	if len(tempFiles) > 0 {
		fmt.Printf("\n临时文件: %d 个\n", len(tempFiles))
		for _, tempFile := range tempFiles {
			fmt.Printf("  - %s\n", tempFile)
		}
	}
}

func demonstratePasswordManagement() {
	fmt.Println("\n=== 密码管理演示 ===")

	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(nil)

	// 显示默认密码列表（前10个）
	defaultPasswords := decryptor.GetCommonPasswords()
	fmt.Printf("默认密码列表包含 %d 个密码\n", len(defaultPasswords))
	fmt.Println("前10个密码:")
	for i, password := range defaultPasswords {
		if i >= 10 {
			break
		}
		if password == "" {
			fmt.Printf("  %d. (空密码)\n", i+1)
		} else {
			fmt.Printf("  %d. %s\n", i+1, password)
		}
	}

	// 演示密码管理操作
	fmt.Println("\n密码管理操作:")

	// 添加自定义密码
	customPasswords := []string{"mypassword", "secret123", "admin2024"}
	for _, password := range customPasswords {
		decryptor.AddCommonPassword(password)
		fmt.Printf("添加密码: %s\n", password)
	}

	// 显示更新后的密码数量
	updatedPasswords := decryptor.GetCommonPasswords()
	fmt.Printf("更新后密码列表包含 %d 个密码\n", len(updatedPasswords))

	// 移除一个密码
	decryptor.RemoveCommonPassword("secret123")
	fmt.Println("移除密码: secret123")

	finalPasswords := decryptor.GetCommonPasswords()
	fmt.Printf("最终密码列表包含 %d 个密码\n", len(finalPasswords))

	// 演示设置完全自定义的密码列表
	customList := []string{"password1", "password2", "password3"}
	decryptor.SetCommonPasswords(customList)
	fmt.Printf("设置自定义密码列表: %v\n", customList)

	currentList := decryptor.GetCommonPasswords()
	fmt.Printf("当前密码列表: %v\n", currentList)
}

func demonstrateTempFileManagement() {
	fmt.Println("\n=== 临时文件管理演示 ===")

	tempDir := filepath.Join(os.TempDir(), "temp_file_demo")
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

	fmt.Printf("临时文件目录: %s\n", tempDir)

	// 显示初始状态
	initialFiles := decryptor.GetTempFiles()
	fmt.Printf("初始临时文件数量: %d\n", len(initialFiles))

	// 模拟创建一些临时文件
	tempFiles := []string{
		filepath.Join(tempDir, "temp1.pdf"),
		filepath.Join(tempDir, "temp2.pdf"),
		filepath.Join(tempDir, "temp3.pdf"),
	}

	for i, tempFile := range tempFiles {
		// 创建实际文件
		content := fmt.Sprintf("临时文件内容 %d", i+1)
		err := os.WriteFile(tempFile, []byte(content), 0644)
		if err != nil {
			fmt.Printf("创建临时文件失败: %v\n", err)
			continue
		}

		// 注意：addTempFile是私有方法，这里只是演示概念
		// 在实际使用中，临时文件会在解密过程中自动管理
		fmt.Printf("创建临时文件: %s\n", filepath.Base(tempFile))
	}

	// 显示临时文件列表（在实际解密后会有临时文件）
	currentFiles := decryptor.GetTempFiles()
	fmt.Printf("当前临时文件数量: %d\n", len(currentFiles))

	fmt.Println("\n注意：临时文件管理在实际解密过程中自动进行")
	fmt.Println("解密器会自动跟踪和清理解密过程中创建的临时文件")

	// 演示清理功能
	fmt.Println("\n清理临时文件...")
	err = decryptor.CleanupTempFiles()
	if err != nil {
		fmt.Printf("清理失败: %v\n", err)
	} else {
		fmt.Println("清理成功")
	}

	// 手动清理我们创建的演示文件
	fmt.Println("清理演示文件:")
	for _, tempFile := range tempFiles {
		if err := os.Remove(tempFile); err == nil {
			fmt.Printf("  ✓ 删除 %s\n", filepath.Base(tempFile))
		} else {
			fmt.Printf("  ✗ 删除 %s 失败: %v\n", filepath.Base(tempFile), err)
		}
	}
}

func demonstrateProgressCallback() {
	fmt.Println("\n=== 进度回调演示 ===")

	// 创建进度回调函数
	progressCallback := func(current, total int, password string) {
		percentage := float64(current) / float64(total) * 100
		if password == "" {
			fmt.Printf("[%.1f%%] 尝试空密码 (%d/%d)\n", percentage, current, total)
		} else {
			fmt.Printf("[%.1f%%] 尝试密码: %s (%d/%d)\n", percentage, password, current, total)
		}
	}

	// 创建带进度回调的解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory:    os.TempDir(),
		MaxAttempts:      5,
		AttemptDelay:     time.Millisecond * 200,
		ProgressCallback: progressCallback,
	})

	fmt.Println("进度回调已设置，将在实际解密时显示进度")
	fmt.Printf("最大尝试次数: %d\n", decryptor.GetMaxAttempts())
	fmt.Printf("尝试延迟: %v\n", decryptor.GetAttemptDelay())
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
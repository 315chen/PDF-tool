//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF自动解密功能演示 ===\n")

	// 1. 演示PDF解密器基本功能
	demonstratePDFDecryptorBasics()

	// 2. 演示密码管理器功能
	demonstratePasswordManager()

	// 3. 演示自动解密功能
	demonstrateAutoDecrypt()

	// 4. 演示进度跟踪功能
	demonstrateProgressTracking()

	// 5. 演示批量解密功能
	demonstrateBatchDecrypt()

	// 6. 演示密码强度分析
	demonstratePasswordStrengthAnalysis()

	// 7. 演示完整的解密流程
	demonstrateCompleteDecryptFlow()

	fmt.Println("\n=== PDF自动解密演示完成 ===")
}

func demonstratePDFDecryptorBasics() {
	fmt.Println("1. PDF解密器基本功能演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "decryptor-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建解密器
	fmt.Printf("   创建PDF解密器，临时目录: %s\n", tempDir)
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
		MaxAttempts:   10,
		AttemptDelay:  time.Millisecond * 100,
	})
	defer decryptor.CleanupTempFiles()
	
	// 1.1 显示解密器配置
	fmt.Println("\n   1.1 解密器配置:")
	fmt.Printf("   - 最大尝试次数: %d\n", decryptor.GetMaxAttempts())
	fmt.Printf("   - 尝试延迟: %v\n", decryptor.GetAttemptDelay())
	fmt.Printf("   - 临时目录: %s\n", tempDir)
	
	// 1.2 显示默认密码列表
	fmt.Println("\n   1.2 默认密码列表:")
	defaultPasswords := decryptor.GetCommonPasswords()
	fmt.Printf("   - 默认密码数量: %d\n", len(defaultPasswords))
	fmt.Printf("   - 前10个密码: ")
	for i, password := range defaultPasswords {
		if i >= 10 {
			break
		}
		if password == "" {
			fmt.Printf("(空), ")
		} else {
			fmt.Printf("%s, ", password)
		}
	}
	fmt.Println()
	
	// 1.3 创建测试PDF文件
	fmt.Println("\n   1.3 创建测试PDF文件:")
	testFiles := createTestPDFFiles(tempDir)
	for name, path := range testFiles {
		fmt.Printf("   - %s: %s\n", name, filepath.Base(path))
	}
	
	// 1.4 检查加密状态
	fmt.Println("\n   1.4 检查PDF加密状态:")
	for name, path := range testFiles {
		isEncrypted, err := decryptor.IsPDFEncrypted(path)
		if err != nil {
			fmt.Printf("   - %s: 检查失败 - %v\n", name, err)
		} else {
			status := "未加密"
			if isEncrypted {
				status = "已加密"
			}
			fmt.Printf("   - %s: %s\n", name, status)
		}
	}
	
	fmt.Println()
}

func demonstratePasswordManager() {
	fmt.Println("2. 密码管理器功能演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "password-manager-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建密码管理器
	fmt.Printf("   创建密码管理器，缓存目录: %s\n", tempDir)
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"password", "123456", "admin", "secret"},
		EnableCache:     true,
		EnableStats:     true,
	})
	
	// 2.1 基本密码操作
	fmt.Println("\n   2.1 基本密码操作:")
	
	// 添加常用密码
	newPasswords := []string{"mypassword", "test123", "admin2024"}
	for _, password := range newPasswords {
		passwordManager.AddCommonPassword(password)
		fmt.Printf("   - 添加密码: %s\n", password)
	}
	
	// 显示当前密码列表
	currentPasswords := passwordManager.GetCommonPasswords()
	fmt.Printf("   - 当前密码数量: %d\n", len(currentPasswords))
	
	// 移除一个密码
	passwordManager.RemoveCommonPassword("test123")
	fmt.Printf("   - 移除密码: test123\n")
	
	updatedPasswords := passwordManager.GetCommonPasswords()
	fmt.Printf("   - 更新后密码数量: %d\n", len(updatedPasswords))
	
	// 2.2 密码缓存功能
	fmt.Println("\n   2.2 密码缓存功能:")
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	os.WriteFile(testFile, []byte("test pdf content"), 0644)
	
	// 设置密码缓存
	testPassword := "cached_password"
	passwordManager.SetPassword(testFile, testPassword)
	fmt.Printf("   - 为文件设置缓存密码: %s\n", testPassword)
	
	// 获取缓存密码
	if cachedPassword, exists := passwordManager.GetPassword(testFile); exists {
		fmt.Printf("   - 从缓存获取密码: %s ✓\n", cachedPassword)
	} else {
		fmt.Printf("   - 缓存中未找到密码 ✗\n")
	}
	
	// 移除缓存
	passwordManager.RemovePassword(testFile)
	fmt.Printf("   - 移除密码缓存\n")
	
	if _, exists := passwordManager.GetPassword(testFile); !exists {
		fmt.Printf("   - 缓存已清除 ✓\n")
	}
	
	// 2.3 密码统计功能
	fmt.Println("\n   2.3 密码统计功能:")
	
	// 模拟密码使用
	testPasswords := []string{"password", "123456", "password", "admin", "password"}
	for _, password := range testPasswords {
		passwordManager.SetPassword(fmt.Sprintf("file_%s.pdf", password), password)
	}
	
	// 获取优化的密码列表
	optimizedList := passwordManager.GetOptimizedPasswordList()
	fmt.Printf("   - 优化后密码列表 (按使用频率排序):\n")
	for i, password := range optimizedList {
		if i >= 5 { // 只显示前5个
			break
		}
		fmt.Printf("     %d. %s\n", i+1, password)
	}
	
	fmt.Println()
}

func demonstrateAutoDecrypt() {
	fmt.Println("3. 自动解密功能演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "auto-decrypt-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
		MaxAttempts:   5,
		AttemptDelay:  time.Millisecond * 50,
	})
	defer decryptor.CleanupTempFiles()
	
	// 创建测试文件
	testFiles := createTestPDFFiles(tempDir)
	
	// 3.1 自动解密未加密文件
	fmt.Println("\n   3.1 自动解密未加密文件:")
	unencryptedFile := testFiles["未加密PDF"]
	
	result, err := decryptor.AutoDecrypt(unencryptedFile)
	if err != nil {
		fmt.Printf("   - 解密失败: %v\n", err)
	} else {
		fmt.Printf("   - 解密结果: 成功=%t, 原文件=%t, 用时=%v\n", 
			result.Success, result.IsOriginalFile, result.ProcessingTime)
		if result.Success && result.IsOriginalFile {
			fmt.Printf("   - 文件未加密，无需解密 ✓\n")
		}
	}
	
	// 3.2 尝试解密模拟加密文件
	fmt.Println("\n   3.2 尝试解密模拟加密文件:")
	encryptedFile := testFiles["模拟加密PDF"]
	
	result, err = decryptor.AutoDecrypt(encryptedFile)
	if err != nil {
		fmt.Printf("   - 解密失败 (预期): %v\n", err)
		fmt.Printf("   - 尝试次数: %d, 用时: %v\n", result.AttemptCount, result.ProcessingTime)
	} else {
		fmt.Printf("   - 解密成功: 使用密码=%s, 尝试次数=%d\n", 
			result.UsedPassword, result.AttemptCount)
	}
	
	// 3.3 使用自定义密码列表
	fmt.Println("\n   3.3 使用自定义密码列表:")
	customPasswords := []string{"custom1", "custom2", "mypassword"}
	
	result, err = decryptor.TryDecryptWithPasswords(unencryptedFile, customPasswords)
	if err != nil {
		fmt.Printf("   - 解密失败: %v\n", err)
	} else {
		fmt.Printf("   - 解密结果: 成功=%t, 尝试次数=%d\n", 
			result.Success, result.AttemptCount)
	}
	
	fmt.Println()
}

func demonstrateProgressTracking() {
	fmt.Println("4. 进度跟踪功能演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "progress-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建带进度回调的解密器
	var progressLog []string
	progressCallback := func(current, total int, password string) {
		percentage := float64(current) / float64(total) * 100
		message := fmt.Sprintf("[%.1f%%] 尝试密码 %d/%d", percentage, current, total)
		if password == "" {
			message += " (空密码)"
		} else {
			message += fmt.Sprintf(": %s", password)
		}
		progressLog = append(progressLog, message)
	}
	
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory:    tempDir,
		MaxAttempts:      3,
		AttemptDelay:     time.Millisecond * 10,
		ProgressCallback: progressCallback,
	})
	defer decryptor.CleanupTempFiles()
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "progress_test.pdf")
	os.WriteFile(testFile, []byte("%PDF-1.4\ntest content\n%%EOF"), 0644)
	
	fmt.Printf("   创建带进度回调的解密器\n")
	fmt.Printf("   - 最大尝试次数: %d\n", decryptor.GetMaxAttempts())
	
	// 执行解密（会触发进度回调）
	fmt.Println("\n   执行解密并跟踪进度:")
	result, err := decryptor.AutoDecrypt(testFile)
	
	// 显示进度日志
	fmt.Printf("   - 进度回调次数: %d\n", len(progressLog))
	for _, log := range progressLog {
		fmt.Printf("     %s\n", log)
	}
	
	// 显示最终结果
	if err != nil {
		fmt.Printf("   - 最终结果: 失败 - %v\n", err)
	} else {
		fmt.Printf("   - 最终结果: 成功=%t, 尝试次数=%d\n", 
			result.Success, result.AttemptCount)
	}
	
	fmt.Println()
}

func demonstrateBatchDecrypt() {
	fmt.Println("5. 批量解密功能演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "batch-decrypt-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
		MaxAttempts:   3,
		AttemptDelay:  time.Millisecond * 20,
	})
	defer decryptor.CleanupTempFiles()
	
	// 创建多个测试文件
	testFiles := []string{}
	for i := 1; i <= 4; i++ {
		fileName := fmt.Sprintf("batch_test_%d.pdf", i)
		filePath := filepath.Join(tempDir, fileName)
		content := fmt.Sprintf("%%PDF-1.4\nBatch test file %d\n%%%%EOF", i)
		os.WriteFile(filePath, []byte(content), 0644)
		testFiles = append(testFiles, filePath)
	}
	
	fmt.Printf("   创建了 %d 个测试文件\n", len(testFiles))
	
	// 5.1 批量检查加密状态
	fmt.Println("\n   5.1 批量检查加密状态:")
	encryptedCount := 0
	for i, file := range testFiles {
		isEncrypted, err := decryptor.IsPDFEncrypted(file)
		if err != nil {
			fmt.Printf("   - 文件%d: 检查失败 - %v\n", i+1, err)
		} else {
			status := "未加密"
			if isEncrypted {
				status = "已加密"
				encryptedCount++
			}
			fmt.Printf("   - 文件%d: %s\n", i+1, status)
		}
	}
	
	fmt.Printf("   - 加密文件数量: %d/%d\n", encryptedCount, len(testFiles))
	
	// 5.2 批量解密处理
	fmt.Println("\n   5.2 批量解密处理:")
	successCount := 0
	totalTime := time.Duration(0)
	
	for i, file := range testFiles {
		fmt.Printf("   处理文件%d: %s\n", i+1, filepath.Base(file))
		
		result, err := decryptor.AutoDecrypt(file)
		totalTime += result.ProcessingTime
		
		if err != nil {
			fmt.Printf("     - 解密失败: %v\n", err)
		} else {
			if result.Success {
				successCount++
				if result.IsOriginalFile {
					fmt.Printf("     - 文件未加密，无需解密\n")
				} else {
					fmt.Printf("     - 解密成功: 密码=%s, 尝试=%d\n", 
						result.UsedPassword, result.AttemptCount)
				}
			} else {
				fmt.Printf("     - 解密失败: 尝试=%d\n", result.AttemptCount)
			}
		}
		fmt.Printf("     - 处理时间: %v\n", result.ProcessingTime)
	}
	
	// 5.3 批量处理统计
	fmt.Println("\n   5.3 批量处理统计:")
	fmt.Printf("   - 处理文件数: %d\n", len(testFiles))
	fmt.Printf("   - 成功文件数: %d\n", successCount)
	fmt.Printf("   - 成功率: %.1f%%\n", float64(successCount)/float64(len(testFiles))*100)
	fmt.Printf("   - 总处理时间: %v\n", totalTime)
	fmt.Printf("   - 平均处理时间: %v\n", totalTime/time.Duration(len(testFiles)))
	
	fmt.Println()
}

func demonstratePasswordStrengthAnalysis() {
	fmt.Println("6. 密码强度分析演示:")
	
	// 测试不同强度的密码
	testPasswords := []string{
		"",                    // 空密码
		"123",                 // 弱密码
		"password",            // 常见密码
		"Password123",         // 中等密码
		"MyStr0ng!P@ssw0rd",   // 强密码
		"VeryLongPasswordWithManyCharacters123!@#", // 超长密码
	}
	
	fmt.Println("   密码强度分析结果:")
	for i, password := range testPasswords {
		score := analyzePasswordStrength(password)
		level := getPasswordLevel(score)

		displayPassword := password
		if password == "" {
			displayPassword = "(空密码)"
		} else if len(password) > 20 {
			displayPassword = password[:17] + "..."
		}

		fmt.Printf("   %d. %-25s 分数: %3d, 级别: %s\n",
			i+1, displayPassword, score, level)
	}
	
	// 生成密码建议
	fmt.Println("\n   密码安全建议:")
	suggestions := []string{
		"使用至少8个字符的密码",
		"包含大小写字母、数字和特殊字符",
		"避免使用常见密码如'password'、'123456'",
		"定期更换密码",
		"不要在多个文件中使用相同密码",
	}
	
	for i, suggestion := range suggestions {
		fmt.Printf("   %d. %s\n", i+1, suggestion)
	}
	
	fmt.Println()
}

func demonstrateCompleteDecryptFlow() {
	fmt.Println("7. 完整解密流程演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "complete-decrypt-demo")
	defer os.RemoveAll(tempDir)
	
	// 7.1 初始化组件
	fmt.Println("   7.1 初始化解密组件:")
	
	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"password", "123456", "admin", "secret", "test"},
		EnableCache:     true,
		EnableStats:     true,
	})
	
	// 创建解密器
	decryptor := pdf.NewPDFDecryptor(&pdf.DecryptorOptions{
		TempDirectory: tempDir,
		MaxAttempts:   8,
		AttemptDelay:  time.Millisecond * 30,
	})
	defer decryptor.CleanupTempFiles()
	
	fmt.Printf("   - 密码管理器初始化完成\n")
	fmt.Printf("   - PDF解密器初始化完成\n")
	fmt.Printf("   - 常用密码数量: %d\n", len(passwordManager.GetCommonPasswords()))
	
	// 7.2 创建测试文件
	fmt.Println("\n   7.2 创建测试文件:")
	testFile := filepath.Join(tempDir, "complete_test.pdf")
	testContent := `%PDF-1.4
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
>>
endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000074 00000 n 
0000000120 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
179
%%EOF`
	
	os.WriteFile(testFile, []byte(testContent), 0644)
	fmt.Printf("   - 测试文件创建: %s\n", filepath.Base(testFile))
	
	// 7.3 执行完整解密流程
	fmt.Println("\n   7.3 执行完整解密流程:")
	
	// 步骤1: 检查文件是否存在
	if _, err := os.Stat(testFile); err != nil {
		fmt.Printf("   步骤1: 文件检查失败 - %v\n", err)
		return
	}
	fmt.Printf("   步骤1: 文件存在检查 ✓\n")
	
	// 步骤2: 检查加密状态
	isEncrypted, err := decryptor.IsPDFEncrypted(testFile)
	if err != nil {
		fmt.Printf("   步骤2: 加密状态检查失败 - %v\n", err)
		return
	}
	fmt.Printf("   步骤2: 加密状态检查 ✓ (加密: %t)\n", isEncrypted)
	
	// 步骤3: 检查密码缓存
	cachedPassword, hasCached := passwordManager.GetPassword(testFile)
	if hasCached {
		fmt.Printf("   步骤3: 找到缓存密码 ✓ (%s)\n", cachedPassword)
	} else {
		fmt.Printf("   步骤3: 无缓存密码，将使用常用密码列表\n")
	}
	
	// 步骤4: 执行解密
	fmt.Printf("   步骤4: 执行自动解密...\n")
	
	// 创建进度输出
	progressOutput := &strings.Builder{}
	result, err := decryptor.DecryptWithProgress(testFile, progressOutput)
	
	// 步骤5: 处理结果
	fmt.Printf("   步骤5: 处理解密结果\n")
	if err != nil {
		fmt.Printf("   - 解密失败: %v\n", err)
	} else {
		if result.Success {
			if result.IsOriginalFile {
				fmt.Printf("   - 文件未加密，无需解密 ✓\n")
			} else {
				fmt.Printf("   - 解密成功 ✓\n")
				fmt.Printf("   - 使用密码: %s\n", result.UsedPassword)
				fmt.Printf("   - 尝试次数: %d\n", result.AttemptCount)
				fmt.Printf("   - 解密文件: %s\n", filepath.Base(result.DecryptedPath))
				
				// 缓存成功的密码
				passwordManager.SetPassword(testFile, result.UsedPassword)
				fmt.Printf("   - 密码已缓存 ✓\n")
			}
		} else {
			fmt.Printf("   - 解密失败，尝试了 %d 个密码\n", result.AttemptCount)
		}
		fmt.Printf("   - 处理时间: %v\n", result.ProcessingTime)
	}
	
	// 显示进度输出
	if progressOutput.Len() > 0 {
		fmt.Printf("   - 进度信息:\n")
		lines := strings.Split(strings.TrimSpace(progressOutput.String()), "\n")
		for _, line := range lines {
			fmt.Printf("     %s\n", line)
		}
	}
	
	// 7.4 清理资源
	fmt.Println("\n   7.4 清理资源:")
	decryptor.CleanupTempFiles()
	fmt.Printf("   - 临时文件清理完成 ✓\n")
	
	fmt.Println("\n   完整解密流程演示完成 🎉")
	fmt.Println("   所有组件协同工作正常")
	
	fmt.Println()
}

// 辅助函数

func createTestPDFFiles(dir string) map[string]string {
	files := make(map[string]string)
	
	// 未加密PDF
	unencryptedContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
%%EOF`
	unencryptedPath := filepath.Join(dir, "unencrypted.pdf")
	os.WriteFile(unencryptedPath, []byte(unencryptedContent), 0644)
	files["未加密PDF"] = unencryptedPath
	
	// 模拟加密PDF（包含加密标记）
	encryptedContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Encrypt 2 0 R
>>
endobj
2 0 obj
<<
/Filter /Standard
>>
endobj
%%EOF`
	encryptedPath := filepath.Join(dir, "encrypted.pdf")
	os.WriteFile(encryptedPath, []byte(encryptedContent), 0644)
	files["模拟加密PDF"] = encryptedPath
	
	return files
}

func analyzePasswordStrength(password string) int {
	if password == "" {
		return 0
	}

	score := 0
	length := len(password)

	// 长度评分
	if length >= 8 {
		score += 25
	} else if length >= 6 {
		score += 15
	} else if length >= 4 {
		score += 10
	}

	// 字符类型评分
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	if hasLower {
		score += 15
	}
	if hasUpper {
		score += 15
	}
	if hasDigit {
		score += 15
	}
	if hasSpecial {
		score += 20
	}

	// 长度奖励
	if length > 12 {
		score += 10
	}

	// 常见密码惩罚
	commonPasswords := []string{"password", "123456", "admin", "secret", "test"}
	for _, common := range commonPasswords {
		if password == common {
			score -= 30
			break
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func getPasswordLevel(score int) string {
	switch {
	case score >= 80:
		return "强"
	case score >= 60:
		return "中等"
	case score >= 40:
		return "弱"
	default:
		return "很弱"
	}
}

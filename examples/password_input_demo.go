//go:build ignore
// +build ignore
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== 用户密码输入处理功能演示 ===\n")

	// 1. 演示控制台密码输入
	demonstrateConsolePasswordInput()

	// 2. 演示安全密码输入
	demonstrateSecurePasswordInput()

	// 3. 演示密码验证和重试
	demonstratePasswordValidationRetry()

	// 4. 演示密码输入历史
	demonstratePasswordInputHistory()

	// 5. 演示批量密码输入
	demonstrateBatchPasswordInput()

	// 6. 演示密码强度检查
	demonstratePasswordStrengthCheck()

	// 7. 演示完整的密码输入流程
	demonstrateCompletePasswordInputFlow()

	fmt.Println("\n=== 用户密码输入处理演示完成 ===")
}

func demonstrateConsolePasswordInput() {
	fmt.Println("1. 控制台密码输入演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "console-input-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "test_encrypted.pdf")
	os.WriteFile(testFile, []byte("test pdf content"), 0644)
	
	fmt.Printf("   测试文件: %s\n", filepath.Base(testFile))
	
	// 1.1 基本密码输入
	fmt.Println("\n   1.1 基本密码输入:")
	fmt.Print("   请输入密码: ")
	
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("   读取密码失败: %v\n", err)
	} else {
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Printf("   输入了空密码\n")
		} else {
			fmt.Printf("   输入的密码长度: %d 字符\n", len(password))
		}
	}
	
	// 1.2 带提示的密码输入
	fmt.Println("\n   1.2 带提示的密码输入:")
	promptResult := promptForPassword("   请为文件输入密码", testFile, 1, nil)
	if promptResult.Success {
		fmt.Printf("   密码输入成功: %s\n", promptResult.Password)
		fmt.Printf("   记住密码: %t\n", promptResult.Remember)
	} else {
		fmt.Printf("   密码输入取消\n")
	}
	
	fmt.Println()
}

func demonstrateSecurePasswordInput() {
	fmt.Println("2. 安全密码输入演示:")
	
	// 2.1 模拟安全密码输入
	fmt.Println("\n   2.1 模拟安全密码输入:")
	fmt.Print("   请输入密码: ")

	// 使用标准输入模拟安全密码输入
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')

	if err != nil {
		fmt.Printf("   安全密码输入失败: %v\n", err)
	} else {
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Printf("   输入了空密码\n")
		} else {
			fmt.Printf("   安全密码输入成功，长度: %d 字符\n", len(password))

			// 演示密码强度检查
			strength := analyzePasswordStrength(password)
			fmt.Printf("   密码强度评分: %d/100\n", strength)
		}
	}
	
	// 2.2 密码确认输入
	fmt.Println("\n   2.2 密码确认输入:")
	confirmResult := promptForPasswordConfirmation()
	if confirmResult.Success {
		fmt.Printf("   密码确认成功\n")
	} else {
		fmt.Printf("   密码确认失败: %s\n", confirmResult.Error)
	}
	
	fmt.Println()
}

func demonstratePasswordValidationRetry() {
	fmt.Println("3. 密码验证和重试演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "validation-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"password", "123456", "admin"},
		EnableCache:     true,
		EnableStats:     true,
	})
	
	// 创建测试文件
	testFile := filepath.Join(tempDir, "validation_test.pdf")
	os.WriteFile(testFile, []byte("test content"), 0644)
	
	fmt.Printf("   测试文件: %s\n", filepath.Base(testFile))
	
	// 3.1 模拟密码验证过程
	fmt.Println("\n   3.1 密码验证过程:")
	correctPassword := "correct123"
	maxAttempts := 3
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("   尝试 %d/%d:\n", attempt, maxAttempts)
		
		// 模拟用户输入
		var inputPassword string
		switch attempt {
		case 1:
			inputPassword = "wrong1"
		case 2:
			inputPassword = "wrong2"
		case 3:
			inputPassword = correctPassword
		}
		
		fmt.Printf("   - 输入密码: %s\n", inputPassword)
		
		// 验证密码
		if inputPassword == correctPassword {
			fmt.Printf("   - 验证成功! ✓\n")
			
			// 缓存成功的密码
			passwordManager.SetPassword(testFile, inputPassword)
			fmt.Printf("   - 密码已缓存\n")
			break
		} else {
			fmt.Printf("   - 验证失败 ✗\n")
			if attempt == maxAttempts {
				fmt.Printf("   - 达到最大尝试次数，验证终止\n")
			}
		}
	}
	
	// 3.2 缓存密码验证
	fmt.Println("\n   3.2 缓存密码验证:")
	if cachedPassword, exists := passwordManager.GetPassword(testFile); exists {
		fmt.Printf("   - 从缓存获取密码: %s ✓\n", cachedPassword)
	} else {
		fmt.Printf("   - 缓存中未找到密码 ✗\n")
	}
	
	fmt.Println()
}

func demonstratePasswordInputHistory() {
	fmt.Println("4. 密码输入历史演示:")
	
	// 创建密码历史记录
	passwordHistory := NewPasswordInputHistory(5) // 最多记录5个密码
	
	// 4.1 添加密码历史
	fmt.Println("\n   4.1 添加密码历史:")
	testPasswords := []string{"password1", "admin123", "secret456", "test789", "user000", "new111"}
	
	for _, password := range testPasswords {
		passwordHistory.AddPassword(password)
		fmt.Printf("   - 添加密码: %s\n", password)
	}
	
	// 4.2 显示密码历史
	fmt.Println("\n   4.2 密码历史记录:")
	history := passwordHistory.GetHistory()
	fmt.Printf("   - 历史记录数量: %d\n", len(history))
	for i, password := range history {
		fmt.Printf("   %d. %s\n", i+1, password)
	}
	
	// 4.3 搜索密码历史
	fmt.Println("\n   4.3 搜索密码历史:")
	searchTerm := "admin"
	matches := passwordHistory.SearchHistory(searchTerm)
	fmt.Printf("   - 搜索 '%s' 的结果: %d 个匹配\n", searchTerm, len(matches))
	for _, match := range matches {
		fmt.Printf("     - %s\n", match)
	}
	
	// 4.4 清理密码历史
	fmt.Println("\n   4.4 清理密码历史:")
	passwordHistory.ClearHistory()
	fmt.Printf("   - 历史记录已清理\n")
	fmt.Printf("   - 当前历史记录数量: %d\n", len(passwordHistory.GetHistory()))
	
	fmt.Println()
}

func demonstrateBatchPasswordInput() {
	fmt.Println("5. 批量密码输入演示:")
	
	// 5.1 批量密码收集
	fmt.Println("\n   5.1 批量密码收集:")
	batchPasswords := []string{}
	
	fmt.Printf("   请输入多个密码 (输入空行结束):\n")
	reader := bufio.NewReader(os.Stdin)
	
	for i := 1; i <= 3; i++ { // 限制最多3个密码用于演示
		fmt.Printf("   密码 %d: ", i)
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("   读取失败: %v\n", err)
			break
		}
		
		password = strings.TrimSpace(password)
		if password == "" {
			break
		}
		
		batchPasswords = append(batchPasswords, password)
	}
	
	fmt.Printf("   收集到 %d 个密码\n", len(batchPasswords))
	
	// 5.2 批量密码验证
	fmt.Println("\n   5.2 批量密码验证:")
	validPasswords := []string{}
	
	for i, password := range batchPasswords {
		fmt.Printf("   验证密码 %d: ", i+1)
		
		// 模拟密码验证
		if len(password) >= 6 {
			fmt.Printf("有效 ✓\n")
			validPasswords = append(validPasswords, password)
		} else {
			fmt.Printf("无效 (长度不足) ✗\n")
		}
	}
	
	fmt.Printf("   有效密码数量: %d/%d\n", len(validPasswords), len(batchPasswords))
	
	// 5.3 批量密码统计
	fmt.Println("\n   5.3 批量密码统计:")
	stats := analyzeBatchPasswords(batchPasswords)
	fmt.Printf("   - 平均长度: %.1f 字符\n", stats.AverageLength)
	fmt.Printf("   - 最短密码: %d 字符\n", stats.MinLength)
	fmt.Printf("   - 最长密码: %d 字符\n", stats.MaxLength)
	fmt.Printf("   - 包含数字: %d 个\n", stats.WithNumbers)
	fmt.Printf("   - 包含特殊字符: %d 个\n", stats.WithSpecialChars)
	
	fmt.Println()
}

func demonstratePasswordStrengthCheck() {
	fmt.Println("6. 密码强度检查演示:")
	
	// 测试不同强度的密码
	testPasswords := []string{
		"123",                    // 很弱
		"password",               // 弱
		"Password123",            // 中等
		"MyStr0ng!P@ssw0rd",      // 强
		"VeryLongAndComplexPassword123!@#", // 很强
	}
	
	fmt.Println("\n   6.1 密码强度分析:")
	for i, password := range testPasswords {
		strength := analyzePasswordStrength(password)
		level := getPasswordStrengthLevel(strength)
		
		fmt.Printf("   %d. %-30s 强度: %3d/100 (%s)\n", 
			i+1, password, strength, level)
	}
	
	// 6.2 密码强度建议
	fmt.Println("\n   6.2 密码强度建议:")
	suggestions := []string{
		"使用至少8个字符",
		"包含大小写字母",
		"包含数字和特殊字符",
		"避免常见密码",
		"定期更换密码",
	}
	
	for i, suggestion := range suggestions {
		fmt.Printf("   %d. %s\n", i+1, suggestion)
	}
	
	// 6.3 实时强度检查
	fmt.Println("\n   6.3 实时强度检查:")
	fmt.Print("   请输入密码进行强度检查: ")
	
	reader := bufio.NewReader(os.Stdin)
	userPassword, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("   读取失败: %v\n", err)
	} else {
		userPassword = strings.TrimSpace(userPassword)
		if userPassword != "" {
			strength := analyzePasswordStrength(userPassword)
			level := getPasswordStrengthLevel(strength)
			fmt.Printf("   您的密码强度: %d/100 (%s)\n", strength, level)
			
			// 提供改进建议
			if strength < 60 {
				fmt.Printf("   建议: 增加密码复杂度以提高安全性\n")
			} else if strength < 80 {
				fmt.Printf("   建议: 密码强度良好，可考虑增加长度\n")
			} else {
				fmt.Printf("   建议: 密码强度很好! ✓\n")
			}
		}
	}
	
	fmt.Println()
}

func demonstrateCompletePasswordInputFlow() {
	fmt.Println("7. 完整密码输入流程演示:")
	
	// 创建测试目录
	tempDir, _ := os.MkdirTemp("", "complete-input-demo")
	defer os.RemoveAll(tempDir)
	
	// 7.1 初始化组件
	fmt.Println("\n   7.1 初始化密码输入组件:")
	
	// 创建密码管理器
	passwordManager := pdf.NewPasswordManager(&pdf.PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"password", "123456", "admin", "secret"},
		EnableCache:     true,
		EnableStats:     true,
	})
	
	// 创建密码历史
	passwordHistory := NewPasswordInputHistory(10)
	
	fmt.Printf("   - 密码管理器初始化完成\n")
	fmt.Printf("   - 密码历史记录初始化完成\n")
	
	// 7.2 创建测试文件
	fmt.Println("\n   7.2 创建测试文件:")
	testFile := filepath.Join(tempDir, "complete_test.pdf")
	os.WriteFile(testFile, []byte("test pdf content"), 0644)
	fmt.Printf("   - 测试文件: %s\n", filepath.Base(testFile))
	
	// 7.3 执行完整密码输入流程
	fmt.Println("\n   7.3 执行完整密码输入流程:")
	
	// 步骤1: 检查密码缓存
	fmt.Printf("   步骤1: 检查密码缓存\n")
	if cachedPassword, exists := passwordManager.GetPassword(testFile); exists {
		fmt.Printf("   - 找到缓存密码: %s ✓\n", cachedPassword)
	} else {
		fmt.Printf("   - 无缓存密码，需要用户输入\n")
	}
	
	// 步骤2: 用户密码输入
	fmt.Printf("   步骤2: 用户密码输入\n")
	fmt.Print("   - 请输入文件密码: ")
	
	reader := bufio.NewReader(os.Stdin)
	userPassword, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("   - 密码输入失败: %v\n", err)
		return
	}
	
	userPassword = strings.TrimSpace(userPassword)
	fmt.Printf("   - 用户输入密码长度: %d 字符\n", len(userPassword))
	
	// 步骤3: 密码强度检查
	fmt.Printf("   步骤3: 密码强度检查\n")
	strength := analyzePasswordStrength(userPassword)
	level := getPasswordStrengthLevel(strength)
	fmt.Printf("   - 密码强度: %d/100 (%s)\n", strength, level)
	
	// 步骤4: 密码验证
	fmt.Printf("   步骤4: 密码验证\n")
	// 模拟验证成功
	fmt.Printf("   - 密码验证成功 ✓\n")
	
	// 步骤5: 缓存密码
	fmt.Printf("   步骤5: 缓存密码\n")
	passwordManager.SetPassword(testFile, userPassword)
	fmt.Printf("   - 密码已缓存 ✓\n")
	
	// 步骤6: 添加到历史记录
	fmt.Printf("   步骤6: 添加到历史记录\n")
	passwordHistory.AddPassword(userPassword)
	fmt.Printf("   - 密码已添加到历史记录 ✓\n")
	
	// 步骤7: 显示统计信息
	fmt.Printf("   步骤7: 显示统计信息\n")
	fmt.Printf("   - 缓存密码数量: %d\n", len(passwordManager.GetCommonPasswords()))
	fmt.Printf("   - 历史记录数量: %d\n", len(passwordHistory.GetHistory()))
	
	fmt.Println("\n   完整密码输入流程演示完成 🎉")
	fmt.Println("   所有组件协同工作正常")
	
	fmt.Println()
}

// 辅助结构和函数

type PasswordInputResult struct {
	Success  bool
	Password string
	Remember bool
	Error    string
}

type PasswordInputHistory struct {
	passwords []string
	maxSize   int
}

func NewPasswordInputHistory(maxSize int) *PasswordInputHistory {
	return &PasswordInputHistory{
		passwords: make([]string, 0),
		maxSize:   maxSize,
	}
}

func (h *PasswordInputHistory) AddPassword(password string) {
	// 避免重复
	for _, existing := range h.passwords {
		if existing == password {
			return
		}
	}
	
	h.passwords = append(h.passwords, password)
	
	// 保持最大大小限制
	if len(h.passwords) > h.maxSize {
		h.passwords = h.passwords[1:]
	}
}

func (h *PasswordInputHistory) GetHistory() []string {
	result := make([]string, len(h.passwords))
	copy(result, h.passwords)
	return result
}

func (h *PasswordInputHistory) SearchHistory(term string) []string {
	var matches []string
	for _, password := range h.passwords {
		if strings.Contains(strings.ToLower(password), strings.ToLower(term)) {
			matches = append(matches, password)
		}
	}
	return matches
}

func (h *PasswordInputHistory) ClearHistory() {
	h.passwords = h.passwords[:0]
}

type BatchPasswordStats struct {
	AverageLength    float64
	MinLength        int
	MaxLength        int
	WithNumbers      int
	WithSpecialChars int
}

func promptForPassword(prompt, filePath string, attempt int, lastError error) PasswordInputResult {
	fmt.Printf("%s: ", prompt)
	
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return PasswordInputResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	password = strings.TrimSpace(password)
	return PasswordInputResult{
		Success:  true,
		Password: password,
		Remember: true, // 默认记住
	}
}

func promptForPasswordConfirmation() PasswordInputResult {
	fmt.Print("   请输入密码: ")
	reader := bufio.NewReader(os.Stdin)
	password1, err := reader.ReadString('\n')
	if err != nil {
		return PasswordInputResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	fmt.Print("   请确认密码: ")
	password2, err := reader.ReadString('\n')
	if err != nil {
		return PasswordInputResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	
	password1 = strings.TrimSpace(password1)
	password2 = strings.TrimSpace(password2)
	
	if password1 != password2 {
		return PasswordInputResult{
			Success: false,
			Error:   "密码不匹配",
		}
	}
	
	return PasswordInputResult{
		Success:  true,
		Password: password1,
	}
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
	hasLower, hasUpper, hasDigit, hasSpecial := false, false, false, false
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
	
	if hasLower { score += 15 }
	if hasUpper { score += 15 }
	if hasDigit { score += 15 }
	if hasSpecial { score += 20 }
	
	// 长度奖励
	if length > 12 { score += 10 }
	
	// 常见密码惩罚
	commonPasswords := []string{"password", "123456", "admin", "secret", "test"}
	for _, common := range commonPasswords {
		if password == common {
			score -= 30
			break
		}
	}
	
	if score < 0 { score = 0 }
	if score > 100 { score = 100 }
	
	return score
}

func getPasswordStrengthLevel(score int) string {
	switch {
	case score >= 80:
		return "很强"
	case score >= 60:
		return "强"
	case score >= 40:
		return "中等"
	case score >= 20:
		return "弱"
	default:
		return "很弱"
	}
}

func analyzeBatchPasswords(passwords []string) BatchPasswordStats {
	if len(passwords) == 0 {
		return BatchPasswordStats{}
	}
	
	totalLength := 0
	minLength := len(passwords[0])
	maxLength := len(passwords[0])
	withNumbers := 0
	withSpecialChars := 0
	
	for _, password := range passwords {
		length := len(password)
		totalLength += length
		
		if length < minLength {
			minLength = length
		}
		if length > maxLength {
			maxLength = length
		}
		
		hasNumber := false
		hasSpecial := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
			} else if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				hasSpecial = true
			}
		}
		
		if hasNumber {
			withNumbers++
		}
		if hasSpecial {
			withSpecialChars++
		}
	}
	
	return BatchPasswordStats{
		AverageLength:    float64(totalLength) / float64(len(passwords)),
		MinLength:        minLength,
		MaxLength:        maxLength,
		WithNumbers:      withNumbers,
		WithSpecialChars: withSpecialChars,
	}
}

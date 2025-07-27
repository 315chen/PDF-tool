package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPasswordManager_Basic(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	// 创建密码管理器
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"test123", "password", "123456"},
		EnableCache:     true,
		EnableStats:     true,
	})

	// 测试基本功能
	if pm == nil {
		t.Fatal("PasswordManager should not be nil")
	}

	// 测试常用密码列表
	commonPasswords := pm.GetCommonPasswords()
	if len(commonPasswords) != 3 {
		t.Errorf("Expected 3 common passwords, got %d", len(commonPasswords))
	}
}

func TestPasswordManager_Cache(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
		EnableCache:    true,
	})

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 测试设置和获取密码
	testPassword := "testpassword123"
	pm.SetPassword(testFile, testPassword)

	// 获取密码
	cachedPassword, exists := pm.GetPassword(testFile)
	if !exists {
		t.Error("Password should exist in cache")
	}
	if cachedPassword != testPassword {
		t.Errorf("Expected password %s, got %s", testPassword, cachedPassword)
	}

	// 测试移除密码
	pm.RemovePassword(testFile)
	_, exists = pm.GetPassword(testFile)
	if exists {
		t.Error("Password should not exist after removal")
	}

	// 测试清空缓存
	pm.SetPassword(testFile, testPassword)
	pm.ClearCache()
	_, exists = pm.GetPassword(testFile)
	if exists {
		t.Error("Password should not exist after clearing cache")
	}
}

func TestPasswordManager_CommonPasswords(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
	})

	// 测试添加常用密码
	pm.AddCommonPassword("newpassword")
	pm.AddCommonPassword("anotherpassword")

	commonPasswords := pm.GetCommonPasswords()
	found := false
	for _, pwd := range commonPasswords {
		if pwd == "newpassword" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Added password should be in common passwords list")
	}

	// 测试重复添加
	originalCount := len(commonPasswords)
	pm.AddCommonPassword("newpassword") // 重复添加
	newCount := len(pm.GetCommonPasswords())
	if newCount != originalCount {
		t.Error("Duplicate password should not be added")
	}

	// 测试移除常用密码
	pm.RemoveCommonPassword("newpassword")
	commonPasswords = pm.GetCommonPasswords()
	for _, pwd := range commonPasswords {
		if pwd == "newpassword" {
			t.Error("Removed password should not be in list")
		}
	}
}

func TestPasswordManager_PasswordStrength(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
	})

	// 测试弱密码
	weakPassword := "123"
	strength := pm.ValidatePasswordStrength(weakPassword)
	if strength.Level != "weak" {
		t.Errorf("Expected weak password, got %s", strength.Level)
	}
	if len(strength.Suggestions) == 0 {
		t.Error("Weak password should have suggestions")
	}

	// 测试中等强度密码
	mediumPassword := "password123"
	strength = pm.ValidatePasswordStrength(mediumPassword)
	t.Logf("Password: %s, Score: %d, Level: %s", mediumPassword, strength.Score, strength.Level)
	if strength.Level != "medium" {
		t.Errorf("Expected medium password, got %s (score: %d)", strength.Level, strength.Score)
	}

	// 测试强密码
	strongPassword := "MyStr0ng!P@ssw0rd"
	strength = pm.ValidatePasswordStrength(strongPassword)
	if strength.Level != "strong" {
		t.Errorf("Expected strong password, got %s", strength.Level)
	}
	if strength.Score < 80 {
		t.Errorf("Strong password should have score >= 80, got %d", strength.Score)
	}

	// 测试常见密码
	commonPassword := "123456"
	strength = pm.ValidatePasswordStrength(commonPassword)
	if strength.Level != "weak" {
		t.Errorf("Common password should be weak, got %s", strength.Level)
	}
}

func TestPasswordManager_OptimizedPasswordList(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory:  tempDir,
		CommonPasswords: []string{"common1", "common2", "common3"},
		EnableStats:     true,
	})

	// 模拟使用统计
	testFile1 := filepath.Join(tempDir, "test1.pdf")
	testFile2 := filepath.Join(tempDir, "test2.pdf")
	
	pm.SetPassword(testFile1, "frequent1") // 这会增加统计
	pm.SetPassword(testFile2, "frequent1") // 再次使用，增加频率

	// 获取优化列表
	optimizedList := pm.GetOptimizedPasswordList()
	
	// 检查高频密码是否在前面
	foundFrequent := false
	for i, password := range optimizedList {
		if password == "frequent1" {
			foundFrequent = true
			if i > 2 { // 应该在前面
				t.Error("Frequently used password should be at the beginning")
			}
			break
		}
	}
	if !foundFrequent {
		t.Error("Frequently used password should be in optimized list")
	}
}

func TestPasswordManager_BatchTryPasswords(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
		EnableCache:    true,
	})

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 模拟解密函数
	mockDecryptFunc := func(filePath, password string) (string, error) {
		if password == "correct" {
			return filePath + ".decrypted", nil
		}
		return "", &PDFError{
			Type:    ErrorEncrypted,
			Message: "wrong password",
		}
	}

	// 测试批量尝试
	passwords := []string{"wrong1", "wrong2", "correct", "wrong3"}
	decryptedPath, usedPassword, err := pm.BatchTryPasswords(testFile, passwords, mockDecryptFunc)
	
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if usedPassword != "correct" {
		t.Errorf("Expected password 'correct', got '%s'", usedPassword)
	}
	if decryptedPath != testFile+".decrypted" {
		t.Errorf("Expected decrypted path '%s', got '%s'", testFile+".decrypted", decryptedPath)
	}

	// 验证密码已缓存
	cachedPassword, exists := pm.GetPassword(testFile)
	if !exists {
		t.Error("Password should be cached after successful decryption")
	}
	if cachedPassword != "correct" {
		t.Errorf("Expected cached password 'correct', got '%s'", cachedPassword)
	}
}

func TestPasswordManager_Stats(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
		EnableStats:    true,
	})

	// 模拟使用统计
	testFile1 := filepath.Join(tempDir, "test1.pdf")
	testFile2 := filepath.Join(tempDir, "test2.pdf")
	
	pm.SetPassword(testFile1, "password1")
	pm.SetPassword(testFile2, "password1") // 重复使用
	pm.SetPassword(testFile1, "password2")

	// 获取统计信息
	stats := pm.GetPasswordStats()
	
	if stats.TotalAttempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", stats.TotalAttempts)
	}
	
	if stats.MostUsedPasswords["password1"] < 2 {
		t.Error("password1 should be used at least 2 times")
	}
}

func TestPasswordManager_FileHash(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
	})

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 设置密码
	pm.SetPassword(testFile, "testpassword")

	// 修改文件内容
	time.Sleep(time.Millisecond * 10) // 确保时间戳不同
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 密码应该仍然存在（因为文件路径相同）
	cachedPassword, exists := pm.GetPassword(testFile)
	if !exists {
		t.Error("Password should still exist for same file path")
	}
	if cachedPassword != "testpassword" {
		t.Errorf("Expected cached password 'testpassword', got '%s'", cachedPassword)
	}
}

func TestPasswordManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
		EnableCache:    true,
		EnableStats:    true,
	})

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 并发测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			password := fmt.Sprintf("password%d", id)
			pm.SetPassword(testFile, password)
			pm.GetPassword(testFile)
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证最终状态
	_, exists := pm.GetPassword(testFile)
	if !exists {
		t.Error("Password should exist after concurrent access")
	}
}

func TestPasswordManager_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	pm := NewPasswordManager(&PasswordManagerOptions{
		CacheDirectory: tempDir,
	})

	// 测试空密码
	pm.SetPassword("test.pdf", "")
	cachedPassword, exists := pm.GetPassword("test.pdf")
	if !exists {
		t.Error("Empty password should be cached")
	}
	if cachedPassword != "" {
		t.Error("Cached password should be empty")
	}

	// 测试不存在的文件
	_, exists = pm.GetPassword("nonexistent.pdf")
	if exists {
		t.Error("Password should not exist for nonexistent file")
	}

	// 测试空常用密码列表
	pm.SetCommonPasswords([]string{})
	commonPasswords := pm.GetCommonPasswords()
	if len(commonPasswords) != 0 {
		t.Error("Common passwords should be empty")
	}

	// 测试密码强度验证边界情况
	strength := pm.ValidatePasswordStrength("")
	if strength.Level != "weak" {
		t.Error("Empty password should be weak")
	}

	// 测试非常长的密码
	longPassword := strings.Repeat("a", 1000)
	strength = pm.ValidatePasswordStrength(longPassword)
	t.Logf("Long password: %s..., Score: %d, Level: %s", longPassword[:10], strength.Score, strength.Level)
	if strength.Level != "medium" {
		t.Errorf("Very long password should be medium, got %s (score: %d)", strength.Level, strength.Score)
	}
} 
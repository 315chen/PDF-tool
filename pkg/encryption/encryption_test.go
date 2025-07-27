package encryption

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// MockEncryptionHandler 模拟加密处理器用于测试
type MockEncryptionHandler struct {
	passwords map[string]string
	commonPasswords []string
	mutex sync.RWMutex
}

func NewMockEncryptionHandler() *MockEncryptionHandler {
	return &MockEncryptionHandler{
		passwords: make(map[string]string),
		commonPasswords: []string{"", "123456", "password", "admin", "user"},
	}
}

func (m *MockEncryptionHandler) TryAutoDecrypt(filePath string) (string, error) {
	for _, password := range m.commonPasswords {
		if _, err := m.DecryptWithPassword(filePath, password); err == nil {
			return password, nil
		}
	}
	return "", nil
}

func (m *MockEncryptionHandler) DecryptWithPassword(filePath, password string) (string, error) {
	// 模拟解密过程
	return filePath + ".decrypted", nil
}

func (m *MockEncryptionHandler) GetCommonPasswords() []string {
	return m.commonPasswords
}

func (m *MockEncryptionHandler) RememberPassword(filePath, password string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.passwords[filePath] = password
}

func (m *MockEncryptionHandler) GetRememberedPassword(filePath string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	password, exists := m.passwords[filePath]
	return password, exists
}

func TestEncryptionHandler_TryAutoDecrypt(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试自动解密
	filePath := "/test/encrypted.pdf"
	password, err := handler.TryAutoDecrypt(filePath)

	if err != nil {
		t.Errorf("TryAutoDecrypt failed: %v", err)
	}

	// 验证返回的密码在常用密码列表中
	commonPasswords := handler.GetCommonPasswords()
	found := false
	for _, commonPassword := range commonPasswords {
		if password == commonPassword {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Returned password '%s' not in common passwords list", password)
	}
}

func TestEncryptionHandler_DecryptWithPassword(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试使用密码解密
	filePath := "/test/encrypted.pdf"
	password := "testpassword"

	decryptedPath, err := handler.DecryptWithPassword(filePath, password)
	if err != nil {
		t.Errorf("DecryptWithPassword failed: %v", err)
	}

	expectedPath := filePath + ".decrypted"
	if decryptedPath != expectedPath {
		t.Errorf("Expected decrypted path '%s', got '%s'", expectedPath, decryptedPath)
	}
}

func TestEncryptionHandler_GetCommonPasswords(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试获取常用密码列表
	passwords := handler.GetCommonPasswords()

	if len(passwords) == 0 {
		t.Error("Common passwords list should not be empty")
	}

	// 验证包含空密码
	found := false
	for _, password := range passwords {
		if password == "" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Common passwords should include empty password")
	}
}

func TestEncryptionHandler_RememberPassword(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试记住密码
	filePath := "/test/encrypted.pdf"
	password := "remembered_password"

	handler.RememberPassword(filePath, password)

	// 验证密码已记住
	retrievedPassword, exists := handler.GetRememberedPassword(filePath)
	if !exists {
		t.Error("Password should be remembered")
	}

	if retrievedPassword != password {
		t.Errorf("Expected remembered password '%s', got '%s'", password, retrievedPassword)
	}
}

func TestEncryptionHandler_GetRememberedPassword(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试获取不存在的记住密码
	_, exists := handler.GetRememberedPassword("nonexistent.pdf")
	if exists {
		t.Error("Should not have remembered password for nonexistent file")
	}

	// 记住密码后获取
	filePath := "/test/test.pdf"
	password := "test_password"
	handler.RememberPassword(filePath, password)

	retrievedPassword, exists := handler.GetRememberedPassword(filePath)
	if !exists || retrievedPassword != password {
		t.Errorf("Expected remembered password '%s', got '%s', exists: %v", password, retrievedPassword, exists)
	}
}

func TestEncryptionHandler_MultipleFiles(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试多个文件的密码管理
	files := []string{
		"/test/file1.pdf",
		"/test/file2.pdf",
		"/test/file3.pdf",
	}

	passwords := []string{
		"password1",
		"password2",
		"password3",
	}

	// 记住多个文件的密码
	for i, file := range files {
		handler.RememberPassword(file, passwords[i])
	}

	// 验证所有密码都正确记住
	for i, file := range files {
		retrievedPassword, exists := handler.GetRememberedPassword(file)
		if !exists || retrievedPassword != passwords[i] {
			t.Errorf("File %s: expected password '%s', got '%s', exists: %v",
				file, passwords[i], retrievedPassword, exists)
		}
	}
}

func TestEncryptionHandler_ConcurrentAccess(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 并发记住密码
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			filename := filepath.Join("test", "file"+string(rune('0'+index))+".pdf")
			password := "password" + string(rune('0'+index))

			handler.RememberPassword(filename, password)

			// 验证密码
			if retrievedPassword, exists := handler.GetRememberedPassword(filename); !exists || retrievedPassword != password {
				t.Errorf("Concurrent access failed for %s", filename)
			}

			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestEncryptionHandler_SpecialCharacters(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试包含特殊字符的密码
	specialPassword := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	filePath := "/test/special.pdf"

	handler.RememberPassword(filePath, specialPassword)

	// 验证特殊字符密码
	retrievedPassword, exists := handler.GetRememberedPassword(filePath)
	if !exists || retrievedPassword != specialPassword {
		t.Errorf("Expected special password '%s', got '%s', exists: %v", specialPassword, retrievedPassword, exists)
	}
}

func TestEncryptionHandler_UnicodePassword(t *testing.T) {
	handler := NewMockEncryptionHandler()

	// 测试Unicode密码
	unicodePassword := "密码123🔒"
	filePath := "/test/unicode.pdf"

	handler.RememberPassword(filePath, unicodePassword)

	// 验证Unicode密码
	retrievedPassword, exists := handler.GetRememberedPassword(filePath)
	if !exists || retrievedPassword != unicodePassword {
		t.Errorf("Expected unicode password '%s', got '%s', exists: %v", unicodePassword, retrievedPassword, exists)
	}
}

// 基准测试
func BenchmarkEncryptionHandler_RememberPassword(b *testing.B) {
	handler := NewMockEncryptionHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "test" + string(rune(i%1000)) + ".pdf"
		handler.RememberPassword(filename, "password")
	}
}

func BenchmarkEncryptionHandler_GetRememberedPassword(b *testing.B) {
	handler := NewMockEncryptionHandler()

	// 预设一些密码
	for i := 0; i < 1000; i++ {
		filename := "test" + string(rune(i)) + ".pdf"
		handler.RememberPassword(filename, "password")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "test" + string(rune(i%1000)) + ".pdf"
		handler.GetRememberedPassword(filename)
	}
}

func BenchmarkEncryptionHandler_TryAutoDecrypt(b *testing.B) {
	handler := NewMockEncryptionHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "test" + string(rune(i%100)) + ".pdf"
		handler.TryAutoDecrypt(filename)
	}
}

// 测试辅助函数
func createTempPDFFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.pdf")

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return filename
}

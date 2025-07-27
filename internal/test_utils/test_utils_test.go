package test_utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTempDir(t *testing.T) {
	// 测试创建临时目录
	tempDir := CreateTempDir(t, "test-prefix")

	// 验证目录存在
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temporary directory should exist: %s", tempDir)
	}

	// 验证目录路径包含前缀
	if !strings.Contains(tempDir, "test-prefix") {
		t.Errorf("Temporary directory should contain prefix: %s", tempDir)
	}
}

func TestCreateTestPDFFile(t *testing.T) {
	tempDir := CreateTempDir(t, "test-pdf")

	// 测试创建测试PDF文件
	filename := "test.pdf"

	filePath := CreateTestPDFFile(t, tempDir, filename)

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Test PDF file should exist: %s", filePath)
	}

	// 验证文件路径
	expectedPath := filepath.Join(tempDir, filename)
	if filePath != expectedPath {
		t.Errorf("Expected file path %s, got %s", expectedPath, filePath)
	}

	// 验证文件内容包含PDF标识
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "%PDF-") {
		t.Error("Test PDF file should start with %PDF-")
	}
}

func TestCreateTestFile(t *testing.T) {
	tempDir := CreateTempDir(t, "test-file")

	// 测试创建测试文件
	filename := "test.txt"
	content := []byte("Test file content")

	filePath := CreateTestFile(t, tempDir, filename, content)

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Test file should exist: %s", filePath)
	}

	// 验证文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("Expected content '%s', got '%s'", string(content), string(data))
	}
}

// 基准测试
func BenchmarkCreateTempDir(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateTempDir(b, "benchmark")
	}
}

func BenchmarkCreateTestPDFFile(b *testing.B) {
	tempDir := CreateTempDir(b, "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "bench_test.pdf"
		CreateTestPDFFile(b, tempDir, filename)
	}
}

func BenchmarkCreateTestFile(b *testing.B) {
	tempDir := CreateTempDir(b, "benchmark")
	content := []byte("Benchmark test content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "bench_test.txt"
		CreateTestFile(b, tempDir, filename, content)
	}
}

func TestFileOperations(t *testing.T) {
	tempDir := CreateTempDir(t, "file-ops")

	// 创建测试文件
	testFile := CreateTestPDFFile(t, tempDir, "test.pdf")

	// 验证文件存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Test file should exist")
	}

	// 测试文件不存在的情况
	nonExistentFile := filepath.Join(tempDir, "nonexistent.pdf")

	// 验证文件确实不存在
	if _, err := os.Stat(nonExistentFile); !os.IsNotExist(err) {
		t.Error("Non-existent file should not exist")
	}
}

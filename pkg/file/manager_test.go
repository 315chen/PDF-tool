package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileManagerImpl_ValidateFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	tests := []struct {
		name        string
		setupFile   func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "空文件路径",
			setupFile: func() string {
				return ""
			},
			expectError: true,
			errorMsg:    "文件路径不能为空",
		},
		{
			name: "文件不存在",
			setupFile: func() string {
				return "/nonexistent/file.pdf"
			},
			expectError: true,
			errorMsg:    "文件不存在",
		},
		{
			name: "路径指向目录",
			setupFile: func() string {
				dir := filepath.Join(tempDir, "testdir")
				os.Mkdir(dir, 0755)
				return dir
			},
			expectError: true,
			errorMsg:    "路径指向目录而不是文件",
		},
		{
			name: "非PDF文件",
			setupFile: func() string {
				file := filepath.Join(tempDir, "test.txt")
				os.WriteFile(file, []byte("test content"), 0644)
				return file
			},
			expectError: true,
			errorMsg:    "不支持的文件格式",
		},
		{
			name: "空PDF文件",
			setupFile: func() string {
				file := filepath.Join(tempDir, "empty.pdf")
				os.WriteFile(file, []byte(""), 0644)
				return file
			},
			expectError: true,
			errorMsg:    "文件为空",
		},
		{
			name: "有效PDF文件",
			setupFile: func() string {
				file := filepath.Join(tempDir, "valid.pdf")
				// 创建一个简单的PDF文件内容
				content := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n%%EOF"
				os.WriteFile(file, []byte(content), 0644)
				return file
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile()
			err := fm.ValidateFile(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("错误消息不匹配，期望包含: %s, 实际: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("不期望出现错误，但出现了: %v", err)
				}
			}
		})
	}
}

func TestFileManagerImpl_CreateTempFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 测试创建临时文件
	tempFile, err := fm.CreateTempFile()
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	// 检查文件是否存在
	if !FileExists(tempFile) {
		t.Errorf("临时文件不存在: %s", tempFile)
	}
}

func TestFileManagerImpl_CreateTempFileWithPrefix(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 测试创建带前缀的临时文件
	prefix := "test_prefix_"
	suffix := ".pdf"
	tempFile, file, err := fm.CreateTempFileWithPrefix(prefix, suffix)
	if err != nil {
		t.Fatalf("创建带前缀的临时文件失败: %v", err)
	}
	defer file.Close()

	// 检查文件是否存在
	if !FileExists(tempFile) {
		t.Errorf("临时文件不存在: %s", tempFile)
	}

	// 检查文件名是否包含前缀和后缀
	fileName := filepath.Base(tempFile)
	if !strings.HasPrefix(fileName, prefix) {
		t.Errorf("文件名不包含前缀，期望前缀: %s, 实际文件名: %s", prefix, fileName)
	}
	if !strings.HasSuffix(fileName, suffix) {
		t.Errorf("文件名不包含后缀，期望后缀: %s, 实际文件名: %s", suffix, fileName)
	}
}

func TestFileManagerImpl_CreateTempFileWithContent(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 测试创建带内容的临时文件
	content := []byte("测试内容")
	tempFile, err := fm.CreateTempFileWithContent("content_", ".txt", content)
	if err != nil {
		t.Fatalf("创建带内容的临时文件失败: %v", err)
	}

	// 检查文件是否存在
	if !FileExists(tempFile) {
		t.Errorf("临时文件不存在: %s", tempFile)
	}

	// 读取文件内容并验证
	readContent, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(content), string(readContent))
	}
}

func TestFileManagerImpl_CopyToTempFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 创建源文件
	sourceContent := []byte("%PDF-1.4\n源文件内容\n%%EOF")
	sourceFile := filepath.Join(tempDir, "source.pdf")
	if err := os.WriteFile(sourceFile, sourceContent, 0644); err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}

	// 复制到临时文件
	tempFile, err := fm.CopyToTempFile(sourceFile, "copy_")
	if err != nil {
		t.Fatalf("复制到临时文件失败: %v", err)
	}

	// 检查文件是否存在
	if !FileExists(tempFile) {
		t.Errorf("临时文件不存在: %s", tempFile)
	}

	// 读取文件内容并验证
	readContent, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}

	if string(readContent) != string(sourceContent) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(sourceContent), string(readContent))
	}
}

func TestFileManagerImpl_CleanupTempFiles(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 创建几个临时文件
	var tempFiles []string
	for i := 0; i < 3; i++ {
		tempFile, err := fm.CreateTempFile()
		if err != nil {
			t.Fatalf("创建临时文件失败: %v", err)
		}
		tempFiles = append(tempFiles, tempFile)
	}

	// 验证文件存在
	for _, file := range tempFiles {
		if !FileExists(file) {
			t.Errorf("临时文件不存在: %s", file)
		}
	}

	// 清理临时文件
	err := fm.CleanupTempFiles()
	if err != nil {
		t.Errorf("清理临时文件失败: %v", err)
	}

	// 验证文件已被删除
	for _, file := range tempFiles {
		if FileExists(file) {
			t.Errorf("临时文件未被删除: %s", file)
		}
	}
}

func TestFileManagerImpl_RemoveTempFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 创建临时文件
	tempFile, err := fm.CreateTempFile()
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	// 检查文件是否存在
	if !FileExists(tempFile) {
		t.Errorf("临时文件不存在: %s", tempFile)
	}

	// 删除临时文件
	err = fm.RemoveTempFile(tempFile)
	if err != nil {
		t.Errorf("删除临时文件失败: %v", err)
	}

	// 验证文件已被删除
	if FileExists(tempFile) {
		t.Errorf("临时文件未被删除: %s", tempFile)
	}
}

func TestFileManagerImpl_GetFileInfo(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 创建测试文件
	testContent := "%PDF-1.4\ntest content\n%%EOF"
	testFile := filepath.Join(tempDir, "test.pdf")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := fm.GetFileInfo(testFile)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	// 验证文件信息
	if fileInfo.Name != "test.pdf" {
		t.Errorf("文件名不匹配，期望: test.pdf, 实际: %s", fileInfo.Name)
	}

	if fileInfo.Size != int64(len(testContent)) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", len(testContent), fileInfo.Size)
	}

	if fileInfo.Path != testFile {
		t.Errorf("文件路径不匹配，期望: %s, 实际: %s", testFile, fileInfo.Path)
	}

	if !fileInfo.IsValid {
		t.Errorf("文件应该是有效的")
	}
}

func TestFileManagerImpl_EnsureDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 测试创建新目录
	newDir := filepath.Join(tempDir, "newdir", "subdir")
	err := fm.EnsureDirectoryExists(newDir)
	if err != nil {
		t.Errorf("创建目录失败: %v", err)
	}

	// 验证目录存在
	if !DirExists(newDir) {
		t.Errorf("目录不存在: %s", newDir)
	}

	// 测试已存在的目录
	err = fm.EnsureDirectoryExists(newDir)
	if err != nil {
		t.Errorf("处理已存在目录失败: %v", err)
	}
}

func TestFileManagerImpl_GetTempDir(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 获取临时目录
	dir := fm.GetTempDir()

	// 验证目录存在
	if !DirExists(dir) {
		t.Errorf("临时目录不存在: %s", dir)
	}
}

func TestFileManagerImpl_SetTempFileMaxAge(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 设置临时文件最大保留时间
	maxAge := 1 * time.Hour
	fm.SetTempFileMaxAge(maxAge)

	// 由于这是内部状态，我们无法直接验证，但至少确保方法不会崩溃
}

func TestFileManagerImpl_CopyFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 创建源文件
	sourceContent := []byte("%PDF-1.4\n源文件内容\n%%EOF")
	sourceFile := filepath.Join(tempDir, "source.pdf")
	if err := os.WriteFile(sourceFile, sourceContent, 0644); err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}

	// 复制文件
	destFile := filepath.Join(tempDir, "dest.pdf")
	err := fm.CopyFile(sourceFile, destFile)
	if err != nil {
		t.Fatalf("复制文件失败: %v", err)
	}

	// 验证目标文件存在
	if !FileExists(destFile) {
		t.Errorf("目标文件不存在: %s", destFile)
	}

	// 读取目标文件内容并验证
	readContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("读取目标文件失败: %v", err)
	}

	if string(readContent) != string(sourceContent) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(sourceContent), string(readContent))
	}
}

func TestFileManagerImpl_WriteAndReadFile(t *testing.T) {
	tempDir := t.TempDir()
	fm := NewFileManager(tempDir)

	// 写入文件
	content := []byte("测试内容")
	filePath := filepath.Join(tempDir, "test.txt")
	err := fm.WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 验证文件存在
	if !FileExists(filePath) {
		t.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件
	readContent, err := fm.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	// 验证内容
	if string(readContent) != string(content) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(content), string(readContent))
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
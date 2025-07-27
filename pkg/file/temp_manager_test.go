package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTempFileManager(t *testing.T) {
	// 创建临时目录作为基础目录
	baseDir := t.TempDir()

	// 创建临时文件管理器
	manager, err := NewTempFileManager(baseDir)
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 验证会话目录已创建
	sessionDir := manager.GetSessionDir()
	if !DirExists(sessionDir) {
		t.Errorf("会话目录未创建: %s", sessionDir)
	}

	// 验证会话目录在基础目录内
	if !filepath.HasPrefix(sessionDir, baseDir) {
		t.Errorf("会话目录不在基础目录内: %s", sessionDir)
	}
}

func TestTempFileManager_CreateTempFile(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 创建临时文件
	filePath, file, err := manager.CreateTempFile("test_", ".pdf")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer file.Close()

	// 验证文件已创建
	if !FileExists(filePath) {
		t.Errorf("临时文件未创建: %s", filePath)
	}

	// 验证文件在会话目录内
	sessionDir := manager.GetSessionDir()
	if !filepath.HasPrefix(filePath, sessionDir) {
		t.Errorf("临时文件不在会话目录内: %s", filePath)
	}

	// 验证文件计数
	if count := manager.GetFileCount(); count != 1 {
		t.Errorf("期望文件计数为1，实际为: %d", count)
	}
}

func TestTempFileManager_CreateTempFileWithContent(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 创建带内容的临时文件
	content := []byte("测试内容")
	filePath, err := manager.CreateTempFileWithContent("content_", ".txt", content)
	if err != nil {
		t.Fatalf("创建带内容的临时文件失败: %v", err)
	}

	// 验证文件已创建
	if !FileExists(filePath) {
		t.Errorf("临时文件未创建: %s", filePath)
	}

	// 读取文件内容并验证
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(content), string(readContent))
	}
}

func TestTempFileManager_CopyToTempFile(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 创建源文件
	sourceContent := []byte("源文件内容")
	sourceFile, err := os.CreateTemp("", "source_*.pdf")
	if err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}
	sourcePath := sourceFile.Name()
	defer os.Remove(sourcePath)

	// 写入源文件内容
	if _, err := sourceFile.Write(sourceContent); err != nil {
		t.Fatalf("写入源文件失败: %v", err)
	}
	sourceFile.Close()

	// 复制到临时文件
	tempPath, err := manager.CopyToTempFile(sourcePath, "copy_")
	if err != nil {
		t.Fatalf("复制到临时文件失败: %v", err)
	}

	// 验证临时文件已创建
	if !FileExists(tempPath) {
		t.Errorf("临时文件未创建: %s", tempPath)
	}

	// 读取临时文件内容并验证
	readContent, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}

	if string(readContent) != string(sourceContent) {
		t.Errorf("文件内容不匹配，期望: %s, 实际: %s", string(sourceContent), string(readContent))
	}
}

func TestTempFileManager_RemoveFile(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 创建临时文件
	filePath, file, err := manager.CreateTempFile("remove_", ".tmp")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	file.Close()

	// 删除文件
	if err := manager.RemoveFile(filePath); err != nil {
		t.Fatalf("删除临时文件失败: %v", err)
	}

	// 验证文件已删除
	if FileExists(filePath) {
		t.Errorf("临时文件未被删除: %s", filePath)
	}

	// 验证文件计数
	if count := manager.GetFileCount(); count != 0 {
		t.Errorf("期望文件计数为0，实际为: %d", count)
	}

	// 尝试删除不存在的文件
	nonExistentPath := filepath.Join(manager.GetSessionDir(), "nonexistent.tmp")
	err = manager.RemoveFile(nonExistentPath)
	if err == nil {
		t.Error("期望删除不存在的文件时返回错误，但没有")
	}
}

func TestTempFileManager_Cleanup(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}

	// 创建多个临时文件
	for i := 0; i < 3; i++ {
		_, file, err := manager.CreateTempFile("cleanup_", ".tmp")
		if err != nil {
			t.Fatalf("创建临时文件失败: %v", err)
		}
		file.Close()
	}

	// 获取会话目录
	sessionDir := manager.GetSessionDir()

	// 清理所有文件
	manager.Cleanup()

	// 验证文件计数
	if count := manager.GetFileCount(); count != 0 {
		t.Errorf("期望文件计数为0，实际为: %d", count)
	}

	// 验证会话目录已删除
	if DirExists(sessionDir) {
		t.Errorf("会话目录未被删除: %s", sessionDir)
	}
}

func TestTempFileManager_CleanupExpired(t *testing.T) {
	manager, err := NewTempFileManager("")
	if err != nil {
		t.Fatalf("创建临时文件管理器失败: %v", err)
	}
	defer manager.Close()

	// 设置较短的过期时间
	manager.SetMaxAge(10 * time.Millisecond)

	// 创建临时文件
	filePath, file, err := manager.CreateTempFile("expire_", ".tmp")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	file.Close()

	// 等待文件过期
	time.Sleep(20 * time.Millisecond)

	// 清理过期文件
	manager.CleanupExpired()

	// 验证文件已被删除
	if FileExists(filePath) {
		t.Errorf("过期的临时文件未被删除: %s", filePath)
	}

	// 验证文件计数
	if count := manager.GetFileCount(); count != 0 {
		t.Errorf("期望文件计数为0，实际为: %d", count)
	}
}

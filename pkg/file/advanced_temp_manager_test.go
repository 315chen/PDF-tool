package file

import (
	"os"
	"testing"
	"time"
)

func TestAdvancedTempManager_CreateTempFileWithTags(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024) // 10MB limit
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	tags := []string{"test", "pdf", "temp"}
	filePath, file, err := atm.CreateTempFileWithTags("test_", ".pdf", tags)
	if err != nil {
		t.Fatalf("Failed to create temp file with tags: %v", err)
	}
	file.Close()

	// 验证文件信息
	info, err := atm.GetFileInfo(filePath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if len(info.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(info.Tags))
	}

	// 验证标签索引
	for _, tag := range tags {
		files := atm.GetFilesByTag(tag)
		if len(files) != 1 || files[0] != filePath {
			t.Errorf("Tag index not working for tag %s", tag)
		}
	}
}

func TestAdvancedTempManager_CreateTempFileWithContentAndTags(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	content := []byte("test content for PDF file")
	tags := []string{"content", "test"}
	
	filePath, err := atm.CreateTempFileWithContentAndTags("content_", ".pdf", content, tags)
	if err != nil {
		t.Fatalf("Failed to create temp file with content and tags: %v", err)
	}

	// 验证文件内容
	readContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("File content mismatch")
	}

	// 验证文件信息
	info, err := atm.GetFileInfo(filePath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), info.Size)
	}

	if info.Hash == "" {
		t.Error("Expected hash to be calculated")
	}
}

func TestAdvancedTempManager_SizeLimit(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 100) // 100 bytes limit
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	// 创建一个小文件，应该成功
	smallContent := []byte("small")
	_, err = atm.CreateTempFileWithContentAndTags("small_", ".txt", smallContent, []string{"small"})
	if err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// 尝试创建一个大文件，应该失败
	largeContent := make([]byte, 200)
	_, err = atm.CreateTempFileWithContentAndTags("large_", ".txt", largeContent, []string{"large"})
	if err == nil {
		t.Error("Expected error when exceeding size limit")
	}
}

func TestAdvancedTempManager_GetFilesByTag(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	// 创建多个带相同标签的文件
	tag := "common"
	var filePaths []string

	for i := 0; i < 3; i++ {
		content := []byte("test content")
		filePath, err := atm.CreateTempFileWithContentAndTags("test_", ".txt", content, []string{tag})
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
		filePaths = append(filePaths, filePath)
	}

	// 验证标签查询
	foundFiles := atm.GetFilesByTag(tag)
	if len(foundFiles) != 3 {
		t.Errorf("Expected 3 files with tag %s, got %d", tag, len(foundFiles))
	}

	// 验证所有文件都被找到
	for _, expectedPath := range filePaths {
		found := false
		for _, foundPath := range foundFiles {
			if expectedPath == foundPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File %s not found in tag query results", expectedPath)
		}
	}
}

func TestAdvancedTempManager_RemoveFileAdvanced(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	content := []byte("test content")
	tags := []string{"test", "remove"}
	
	filePath, err := atm.CreateTempFileWithContentAndTags("remove_", ".txt", content, tags)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 验证文件存在
	if _, err := atm.GetFileInfo(filePath); err != nil {
		t.Fatalf("File info should exist: %v", err)
	}

	// 删除文件
	err = atm.RemoveFileAdvanced(filePath)
	if err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// 验证文件信息已删除
	if _, err := atm.GetFileInfo(filePath); err == nil {
		t.Error("File info should not exist after removal")
	}

	// 验证标签索引已更新
	for _, tag := range tags {
		files := atm.GetFilesByTag(tag)
		if len(files) != 0 {
			t.Errorf("Tag %s should have no files after removal", tag)
		}
	}
}

func TestAdvancedTempManager_CleanupByTag(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	// 创建多个文件，部分带有相同标签
	targetTag := "cleanup"
	content := []byte("test content")

	// 创建带目标标签的文件
	for i := 0; i < 2; i++ {
		_, err := atm.CreateTempFileWithContentAndTags("cleanup_", ".txt", content, []string{targetTag})
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
	}

	// 创建不带目标标签的文件
	_, err = atm.CreateTempFileWithContentAndTags("keep_", ".txt", content, []string{"keep"})
	if err != nil {
		t.Fatalf("Failed to create keep file: %v", err)
	}

	// 验证初始状态
	if len(atm.GetFilesByTag(targetTag)) != 2 {
		t.Error("Should have 2 files with target tag")
	}
	if len(atm.GetFilesByTag("keep")) != 1 {
		t.Error("Should have 1 file with keep tag")
	}

	// 按标签清理
	err = atm.CleanupByTag(targetTag)
	if err != nil {
		t.Fatalf("Failed to cleanup by tag: %v", err)
	}

	// 验证清理结果
	if len(atm.GetFilesByTag(targetTag)) != 0 {
		t.Error("Should have no files with target tag after cleanup")
	}
	if len(atm.GetFilesByTag("keep")) != 1 {
		t.Error("Should still have 1 file with keep tag")
	}
}

func TestAdvancedTempManager_CleanupOldFiles(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	content := []byte("test content")

	// 创建文件
	filePath, err := atm.CreateTempFileWithContentAndTags("old_", ".txt", content, []string{"old"})
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 手动设置旧的访问时间
	atm.infoMutex.Lock()
	if info, exists := atm.fileInfos[filePath]; exists {
		info.LastAccessed = time.Now().Add(-2 * time.Hour)
	}
	atm.infoMutex.Unlock()

	// 清理1小时前的文件
	cleanedCount := atm.CleanupOldFiles(1 * time.Hour)
	if cleanedCount != 1 {
		t.Errorf("Expected to clean 1 file, cleaned %d", cleanedCount)
	}

	// 验证文件已被删除
	if _, err := atm.GetFileInfo(filePath); err == nil {
		t.Error("File should have been cleaned up")
	}
}

func TestAdvancedTempManager_GetStatistics(t *testing.T) {
	atm, err := NewAdvancedTempManager("", 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to create AdvancedTempManager: %v", err)
	}
	defer atm.Close()

	// 创建一些文件
	content := []byte("test content")
	tags := []string{"stats", "test"}

	for i := 0; i < 3; i++ {
		_, err := atm.CreateTempFileWithContentAndTags("stats_", ".txt", content, tags)
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
	}

	// 获取统计信息
	stats := atm.GetStatistics()

	if stats["total_files"].(int) != 3 {
		t.Errorf("Expected 3 total files, got %v", stats["total_files"])
	}

	if stats["total_size"].(int64) != int64(len(content)*3) {
		t.Errorf("Expected total size %d, got %v", len(content)*3, stats["total_size"])
	}

	if stats["total_tags"].(int) != 2 {
		t.Errorf("Expected 2 total tags, got %v", stats["total_tags"])
	}

	tagStats := stats["tag_stats"].(map[string]int)
	if tagStats["stats"] != 3 {
		t.Errorf("Expected 3 files with 'stats' tag, got %d", tagStats["stats"])
	}
	if tagStats["test"] != 3 {
		t.Errorf("Expected 3 files with 'test' tag, got %d", tagStats["test"])
	}
}

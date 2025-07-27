package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewOutputManager(t *testing.T) {
	// 测试默认选项
	manager := NewOutputManager(nil)
	if manager == nil {
		t.Fatal("NewOutputManager 返回 nil")
	}

	if manager.baseDir != "." {
		t.Errorf("期望默认基础目录为 '.'，实际为 %s", manager.baseDir)
	}

	if manager.defaultFileName != "merged_output.pdf" {
		t.Errorf("期望默认文件名为 'merged_output.pdf'，实际为 %s", manager.defaultFileName)
	}

	if !manager.autoIncrement {
		t.Error("期望默认启用自动递增")
	}

	// 测试自定义选项
	options := &OutputOptions{
		BaseDirectory:   "/tmp/test",
		DefaultFileName: "custom.pdf",
		AutoIncrement:   false,
		TimestampSuffix: true,
		BackupEnabled:   false,
	}

	manager = NewOutputManager(options)
	if manager.baseDir != "/tmp/test" {
		t.Errorf("期望基础目录为 '/tmp/test'，实际为 %s", manager.baseDir)
	}

	if manager.defaultFileName != "custom.pdf" {
		t.Errorf("期望文件名为 'custom.pdf'，实际为 %s", manager.defaultFileName)
	}

	if manager.autoIncrement {
		t.Error("期望禁用自动递增")
	}

	if !manager.timestampSuffix {
		t.Error("期望启用时间戳后缀")
	}
}

func TestOutputManager_ResolveOutputPath(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "output_manager_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	manager := NewOutputManager(&OutputOptions{
		BaseDirectory:   testDir,
		DefaultFileName: "default.pdf",
		AutoIncrement:   true,
		TimestampSuffix: false,
		BackupEnabled:   true,
	})

	// 测试空路径
	info, err := manager.ResolveOutputPath("")
	if err != nil {
		t.Fatalf("解析空路径失败: %v", err)
	}

	expectedDefault := filepath.Join(testDir, "default.pdf")
	if info.FinalPath != expectedDefault {
		t.Errorf("期望默认路径为 %s，实际为 %s", expectedDefault, info.FinalPath)
	}

	// 测试相对路径
	info, err = manager.ResolveOutputPath("relative.pdf")
	if err != nil {
		t.Fatalf("解析相对路径失败: %v", err)
	}

	expectedRelative := filepath.Join(testDir, "relative.pdf")
	if info.FinalPath != expectedRelative {
		t.Errorf("期望相对路径为 %s，实际为 %s", expectedRelative, info.FinalPath)
	}

	// 测试绝对路径
	absolutePath := filepath.Join(testDir, "absolute.pdf")
	info, err = manager.ResolveOutputPath(absolutePath)
	if err != nil {
		t.Fatalf("解析绝对路径失败: %v", err)
	}

	if info.FinalPath != absolutePath {
		t.Errorf("期望绝对路径为 %s，实际为 %s", absolutePath, info.FinalPath)
	}
}

func TestOutputManager_AutoIncrement(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "auto_increment_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	manager := NewOutputManager(&OutputOptions{
		BaseDirectory: testDir,
		AutoIncrement: true,
	})

	basePath := filepath.Join(testDir, "test.pdf")

	// 创建现有文件
	err = os.WriteFile(basePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试自动递增
	finalPath, incremented := manager.resolveAutoIncrement(basePath)
	if !incremented {
		t.Error("应该触发自动递增")
	}

	expectedPath := filepath.Join(testDir, "test_1.pdf")
	if finalPath != expectedPath {
		t.Errorf("期望递增路径为 %s，实际为 %s", expectedPath, finalPath)
	}

	// 创建递增文件，测试进一步递增
	err = os.WriteFile(expectedPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("创建递增测试文件失败: %v", err)
	}

	finalPath, incremented = manager.resolveAutoIncrement(basePath)
	if !incremented {
		t.Error("应该触发进一步递增")
	}

	expectedPath2 := filepath.Join(testDir, "test_2.pdf")
	if finalPath != expectedPath2 {
		t.Errorf("期望进一步递增路径为 %s，实际为 %s", expectedPath2, finalPath)
	}
}

func TestOutputManager_TimestampSuffix(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "timestamp_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	manager := NewOutputManager(&OutputOptions{
		BaseDirectory:   testDir,
		TimestampSuffix: true,
	})

	originalPath := filepath.Join(testDir, "test.pdf")
	timestampPath := manager.addTimestampSuffix(originalPath)

	// 检查时间戳格式
	if !strings.Contains(timestampPath, "_20") {
		t.Errorf("时间戳路径应该包含年份: %s", timestampPath)
	}

	if !strings.HasSuffix(timestampPath, ".pdf") {
		t.Errorf("时间戳路径应该保持PDF扩展名: %s", timestampPath)
	}

	if strings.Contains(timestampPath, "test.pdf") {
		t.Errorf("时间戳路径不应该包含原始文件名: %s", timestampPath)
	}
}

func TestOutputManager_BackupOperations(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "backup_test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(testDir)

	manager := NewOutputManager(&OutputOptions{
		BaseDirectory: testDir,
		BackupEnabled: true,
	})

	// 创建原始文件
	originalPath := filepath.Join(testDir, "original.pdf")
	originalContent := "original content"
	err = os.WriteFile(originalPath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("创建原始文件失败: %v", err)
	}

	// 生成备份路径
	backupPath := manager.generateBackupPath(originalPath)
	if !strings.Contains(backupPath, "backup_") {
		t.Errorf("备份路径应该包含backup标识: %s", backupPath)
	}

	// 创建备份
	err = manager.CreateBackup(originalPath, backupPath)
	if err != nil {
		t.Fatalf("创建备份失败: %v", err)
	}

	// 验证备份内容
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("读取备份文件失败: %v", err)
	}

	if string(backupContent) != originalContent {
		t.Error("备份内容不匹配")
	}

	// 修改原始文件
	modifiedContent := "modified content"
	err = os.WriteFile(originalPath, []byte(modifiedContent), 0644)
	if err != nil {
		t.Fatalf("修改原始文件失败: %v", err)
	}

	// 恢复备份
	err = manager.RestoreBackup(backupPath, originalPath)
	if err != nil {
		t.Fatalf("恢复备份失败: %v", err)
	}

	// 验证恢复内容
	restoredContent, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("读取恢复文件失败: %v", err)
	}

	if string(restoredContent) != originalContent {
		t.Error("恢复内容不匹配")
	}

	// 清理备份
	err = manager.CleanupBackup(backupPath)
	if err != nil {
		t.Fatalf("清理备份失败: %v", err)
	}

	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("备份文件应该被删除")
	}
}

func TestOutputManager_GetSuggestedPath(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "suggested_path_test")
	manager := NewOutputManager(&OutputOptions{
		BaseDirectory:   testDir,
		DefaultFileName: "merged_output.pdf",
	})

	// 测试空输入文件列表
	suggestedPath := manager.GetSuggestedPath([]string{})
	expectedDefault := filepath.Join(testDir, "merged_output.pdf")
	if suggestedPath != expectedDefault {
		t.Errorf("期望默认建议路径为 %s，实际为 %s", expectedDefault, suggestedPath)
	}

	// 测试有输入文件
	inputFiles := []string{"/path/to/document.pdf", "/path/to/other.pdf"}
	suggestedPath = manager.GetSuggestedPath(inputFiles)
	expectedSuggested := filepath.Join(testDir, "document_merged.pdf")
	if suggestedPath != expectedSuggested {
		t.Errorf("期望建议路径为 %s，实际为 %s", expectedSuggested, suggestedPath)
	}
}

func TestOutputManager_ValidateOutputPath(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "validate_test")
	manager := NewOutputManager(&OutputOptions{
		BaseDirectory: testDir,
	})

	testCases := []struct {
		name        string
		path        string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "有效PDF路径",
			path:        filepath.Join(testDir, "valid.pdf"),
			expectError: false,
		},
		{
			name:        "非PDF扩展名",
			path:        filepath.Join(testDir, "invalid.txt"),
			expectError: true,
			errorType:   ErrorInvalidFile,
		},
		{
			name:        "PDF扩展名大写",
			path:        filepath.Join(testDir, "valid.PDF"),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.ValidateOutputPath(tc.path)
			
			if tc.expectError {
				if err == nil {
					t.Error("期望有错误但没有返回错误")
				} else if pdfErr, ok := err.(*PDFError); ok {
					if pdfErr.Type != tc.errorType {
						t.Errorf("期望错误类型 %v，实际为 %v", tc.errorType, pdfErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("不期望有错误但返回了错误: %v", err)
				}
			}
		})
	}
}

func TestOutputManager_DirectoryOperations(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "directory_test")
	manager := NewOutputManager(&OutputOptions{
		BaseDirectory: testDir,
	})

	// 测试获取输出目录
	if manager.GetOutputDirectory() != testDir {
		t.Errorf("期望输出目录为 %s，实际为 %s", testDir, manager.GetOutputDirectory())
	}

	// 测试设置输出目录
	newDir := filepath.Join(os.TempDir(), "new_directory_test")
	err := manager.SetOutputDirectory(newDir)
	if err != nil {
		t.Fatalf("设置输出目录失败: %v", err)
	}
	defer os.RemoveAll(newDir)

	if manager.GetOutputDirectory() != newDir {
		t.Errorf("期望新输出目录为 %s，实际为 %s", newDir, manager.GetOutputDirectory())
	}

	// 验证目录已创建
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("目录应该被创建")
	}
}

func TestOutputManager_FileNameOperations(t *testing.T) {
	manager := NewOutputManager(nil)

	// 测试获取默认文件名
	if manager.GetDefaultFileName() != "merged_output.pdf" {
		t.Errorf("期望默认文件名为 'merged_output.pdf'，实际为 %s", manager.GetDefaultFileName())
	}

	// 测试设置有效文件名
	err := manager.SetDefaultFileName("new_default.pdf")
	if err != nil {
		t.Fatalf("设置默认文件名失败: %v", err)
	}

	if manager.GetDefaultFileName() != "new_default.pdf" {
		t.Errorf("期望新默认文件名为 'new_default.pdf'，实际为 %s", manager.GetDefaultFileName())
	}

	// 测试设置无效文件名
	err = manager.SetDefaultFileName("invalid.txt")
	if err == nil {
		t.Error("设置无效文件名应该失败")
	}
}
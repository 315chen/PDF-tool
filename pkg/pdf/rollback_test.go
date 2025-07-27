package pdf

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRollbackManager_Basic(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "这是测试内容"
	err = ioutil.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试备份文件
	backupPath, err := rollbackMgr.BackupFile(testFile)
	if err != nil {
		t.Fatalf("备份文件失败: %v", err)
	}

	// 验证备份文件存在
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("备份文件不存在: %v", err)
	}

	// 验证备份文件内容
	backupContent, err := ioutil.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("读取备份文件失败: %v", err)
	}
	if string(backupContent) != testContent {
		t.Errorf("备份文件内容不匹配，期望: %s, 实际: %s", testContent, string(backupContent))
	}
}

func TestRollbackManager_RestoreFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 创建原始文件
	originalFile := filepath.Join(tempDir, "original.txt")
	originalContent := "原始内容"
	err = ioutil.WriteFile(originalFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("创建原始文件失败: %v", err)
	}

	// 备份文件
	backupPath, err := rollbackMgr.BackupFile(originalFile)
	if err != nil {
		t.Fatalf("备份文件失败: %v", err)
	}

	// 修改原始文件
	modifiedContent := "修改后的内容"
	err = ioutil.WriteFile(originalFile, []byte(modifiedContent), 0644)
	if err != nil {
		t.Fatalf("修改原始文件失败: %v", err)
	}

	// 恢复文件
	err = rollbackMgr.RestoreFile(backupPath, originalFile)
	if err != nil {
		t.Fatalf("恢复文件失败: %v", err)
	}

	// 验证文件已恢复
	restoredContent, err := ioutil.ReadFile(originalFile)
	if err != nil {
		t.Fatalf("读取恢复后的文件失败: %v", err)
	}
	if string(restoredContent) != originalContent {
		t.Errorf("文件恢复失败，期望: %s, 实际: %s", originalContent, string(restoredContent))
	}
}

func TestRollbackManager_RollbackIfFailed(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "原始内容"
	err = ioutil.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试成功操作（不应该回滚）
	err = rollbackMgr.RollbackIfFailed(testFile, func() error {
		// 模拟成功操作
		return nil
	})
	if err != nil {
		t.Fatalf("成功操作不应该返回错误: %v", err)
	}

	// 验证文件内容未改变
	currentContent, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if string(currentContent) != originalContent {
		t.Errorf("成功操作后文件内容不应改变，期望: %s, 实际: %s", originalContent, string(currentContent))
	}

	// 测试失败操作（应该回滚）
	err = rollbackMgr.RollbackIfFailed(testFile, func() error {
		// 在操作中修改文件内容
		modifiedContent := "修改后的内容"
		err := ioutil.WriteFile(testFile, []byte(modifiedContent), 0644)
		if err != nil {
			return err
		}
		// 模拟失败操作
		return fmt.Errorf("模拟操作失败")
	})
	if err == nil {
		t.Fatal("失败操作应该返回错误")
	}

	// 验证文件已回滚
	restoredContent, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("读取回滚后的文件失败: %v", err)
	}
	if string(restoredContent) != originalContent {
		t.Errorf("文件回滚失败，期望: %s, 实际: %s", originalContent, string(restoredContent))
	}
}

func TestRollbackManager_ErrorHandling(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 测试备份不存在的文件
	_, err = rollbackMgr.BackupFile("不存在的文件.txt")
	if err == nil {
		t.Fatal("备份不存在的文件应该返回错误")
	}

	// 测试恢复不存在的备份文件
	err = rollbackMgr.RestoreFile("不存在的备份文件.txt", "目标文件.txt")
	if err == nil {
		t.Fatal("恢复不存在的备份文件应该返回错误")
	}

	// 测试回滚前备份失败
	err = rollbackMgr.RollbackIfFailed("不存在的文件.txt", func() error {
		return fmt.Errorf("模拟操作失败")
	})
	if err == nil {
		t.Fatal("回滚前备份失败应该返回错误")
	}
}

func TestRollbackManager_ConcurrentAccess(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "concurrent_test.txt")
	testContent := "并发测试内容"
	err = ioutil.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 并发测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 并发备份
			backupPath, err := rollbackMgr.BackupFile(testFile)
			if err != nil {
				t.Errorf("并发备份 %d 失败: %v", id, err)
				return
			}

			// 验证备份文件
			if _, err := os.Stat(backupPath); err != nil {
				t.Errorf("并发备份 %d 的备份文件不存在: %v", id, err)
				return
			}

			// 并发恢复
			err = rollbackMgr.RestoreFile(backupPath, testFile)
			if err != nil {
				t.Errorf("并发恢复 %d 失败: %v", id, err)
				return
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRollbackManager_IntegrationWithWriter(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试PDF文件
	testPDF := filepath.Join(tempDir, "test.pdf")
	testContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids []\n/Count 0\n>>\nendobj\nxref\n0 3\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \ntrailer\n<<\n/Size 3\n/Root 1 0 R\n>>\nstartxref\n108\n%%EOF")
	err = ioutil.WriteFile(testPDF, testContent, 0644)
	if err != nil {
		t.Fatalf("创建测试PDF文件失败: %v", err)
	}

	// 创建PDFWriter
	writer, err := NewPDFWriter(testPDF, &WriterOptions{
		BackupEnabled: true,
		TempDirectory: tempDir,
	})
	if err != nil {
		t.Fatalf("创建PDFWriter失败: %v", err)
	}

	// 打开写入器
	err = writer.Open()
	if err != nil {
		t.Fatalf("打开PDFWriter失败: %v", err)
	}
	defer writer.Close()

	// 添加内容
	err = writer.AddContent(testContent)
	if err != nil {
		t.Fatalf("添加内容失败: %v", err)
	}

	// 测试写入成功（不应该回滚）
	result, err := writer.Write(context.Background(), nil)
	if err != nil {
		t.Fatalf("写入成功不应该返回错误: %v", err)
	}
	if !result.Success {
		t.Fatal("写入应该成功")
	}

	// 验证文件存在
	if _, err := os.Stat(testPDF); err != nil {
		t.Fatalf("写入后的文件不存在: %v", err)
	}
}

func TestRollbackManager_IntegrationWithMerger(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试PDF文件
	pdf1 := filepath.Join(tempDir, "test1.pdf")
	pdf2 := filepath.Join(tempDir, "test2.pdf")
	outputPDF := filepath.Join(tempDir, "merged.pdf")

	// 创建简单的PDF内容（更完整的PDF结构）
	pdfContent := []byte(`%PDF-1.4
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
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
72 720 Td
(Hello World) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000111 00000 n 
0000000206 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
297
%%EOF`)

	err = ioutil.WriteFile(pdf1, pdfContent, 0644)
	if err != nil {
		t.Fatalf("创建PDF1失败: %v", err)
	}

	err = ioutil.WriteFile(pdf2, pdfContent, 0644)
	if err != nil {
		t.Fatalf("创建PDF2失败: %v", err)
	}

	// 创建合并器
	merger := NewStreamingMerger(&MergeOptions{
		TempDirectory: tempDir,
	})

	// 测试合并成功
	result, err := merger.MergeFiles([]string{pdf1, pdf2}, outputPDF, nil)
	if err != nil {
		t.Fatalf("合并成功不应该返回错误: %v", err)
	}
	if result.ProcessedFiles != 2 {
		t.Errorf("期望处理2个文件，实际处理 %d 个", result.ProcessedFiles)
	}

	// 验证输出文件存在
	if _, err := os.Stat(outputPDF); err != nil {
		t.Fatalf("合并后的文件不存在: %v", err)
	}
}

func TestRollbackManager_Performance(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "rollback_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建回滚管理器
	rollbackMgr := NewRollbackManager(tempDir)

	// 创建大文件进行性能测试
	largeFile := filepath.Join(tempDir, "large_test.txt")
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err = ioutil.WriteFile(largeFile, largeContent, 0644)
	if err != nil {
		t.Fatalf("创建大文件失败: %v", err)
	}

	// 性能测试：备份大文件
	startTime := time.Now()
	backupPath, err := rollbackMgr.BackupFile(largeFile)
	backupTime := time.Since(startTime)
	if err != nil {
		t.Fatalf("备份大文件失败: %v", err)
	}

	t.Logf("备份1MB文件耗时: %v", backupTime)

	// 性能测试：恢复大文件
	startTime = time.Now()
	err = rollbackMgr.RestoreFile(backupPath, largeFile)
	restoreTime := time.Since(startTime)
	if err != nil {
		t.Fatalf("恢复大文件失败: %v", err)
	}

	t.Logf("恢复1MB文件耗时: %v", restoreTime)

	// 验证文件完整性
	restoredContent, err := ioutil.ReadFile(largeFile)
	if err != nil {
		t.Fatalf("读取恢复后的文件失败: %v", err)
	}
	if len(restoredContent) != len(largeContent) {
		t.Errorf("文件大小不匹配，期望: %d, 实际: %d", len(largeContent), len(restoredContent))
	}
}

package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempDir 创建临时测试目录
func createTempDir(t testing.TB, prefix string) string {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	
	// 注册清理函数
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	
	return tempDir
}

// createTestFile 创建测试文件
func createTestFile(t testing.TB, dir, filename string, content []byte) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("无法创建测试文件 %s: %v", filePath, err)
	}
	return filePath
}

// createTestPDFFile 创建简单的测试PDF文件
func createTestPDFFile(t testing.TB, dir, filename string) string {
	// 创建一个简单的PDF文件内容
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
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
173
%%EOF`)
	
	return createTestFile(t, dir, filename, pdfContent)
}

// createCorruptedPDFFile 创建损坏的PDF文件
func createCorruptedPDFFile(t testing.TB, dir, filename string) string {
	// 创建一个损坏的PDF文件内容
	corruptedContent := []byte(`%PDF-1.4
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
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
999999
%%EOF`)
	
	return createTestFile(t, dir, filename, corruptedContent)
}

// testFileExists 检查文件是否存在（测试专用）
func testFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

package test_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestingInterface 测试接口，支持testing.T和testing.B
type TestingInterface interface {
	Fatalf(format string, args ...interface{})
	Cleanup(func())
}

// CreateTempDir 创建临时测试目录
func CreateTempDir(t TestingInterface, prefix string) string {
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

// CreateTestFile 创建测试文件
func CreateTestFile(t TestingInterface, dir, filename string, content []byte) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("无法创建测试文件 %s: %v", filePath, err)
	}
	return filePath
}

// CreateTestPDFFile 创建简单的测试PDF文件
func CreateTestPDFFile(t TestingInterface, dir, filename string) string {
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

	return CreateTestFile(t, dir, filename, pdfContent)
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// GetFileSize 获取文件大小
func GetFileSize(t TestingInterface, filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("无法获取文件信息 %s: %v", filePath, err)
	}
	return info.Size()
}

// CompareFiles 比较两个文件是否相同
func CompareFiles(t TestingInterface, file1, file2 string) bool {
	content1, err := os.ReadFile(file1)
	if err != nil {
		t.Fatalf("无法读取文件 %s: %v", file1, err)
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		t.Fatalf("无法读取文件 %s: %v", file2, err)
	}

	if len(content1) != len(content2) {
		return false
	}

	for i := range content1 {
		if content1[i] != content2[i] {
			return false
		}
	}

	return true
}

// CreateEncryptedPDFFile 创建加密的测试PDF文件
func CreateEncryptedPDFFile(t TestingInterface, dir, filename string) string {
	// 创建一个简单的加密PDF文件内容（模拟）
	encryptedContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Encrypt 4 0 R
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

4 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O <encrypted_owner_password>
/U <encrypted_user_password>
/P -4
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000068 00000 n 
0000000140 00000 n 
0000000197 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
297
%%EOF`)

	return CreateTestFile(t, dir, filename, encryptedContent)
}

// CreateCorruptedPDFFile 创建损坏的测试PDF文件
func CreateCorruptedPDFFile(t TestingInterface, dir, filename string) string {
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
/MediaBox [0 0 612 792
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
CORRUPTED
%%EOF`)

	return CreateTestFile(t, dir, filename, corruptedContent)
}

// CreateLargePDFFile 创建大的测试PDF文件
func CreateLargePDFFile(t TestingInterface, dir, filename string, sizeKB int) string {
	// 如果请求的大小太小，使用基本PDF
	if sizeKB < 1 {
		return CreateTestPDFFile(t, dir, filename)
	}

	// 创建填充内容 - 使用简单的重复字符
	paddingSize := sizeKB*1024 - 400 // 减去基础PDF结构的大小
	if paddingSize < 0 {
		paddingSize = 100 // 最小填充
	}

	// 创建填充数据
	padding := strings.Repeat("A", paddingSize)

	// 构建PDF内容，使用正确的偏移量
	var content strings.Builder
	content.WriteString("%PDF-1.4\n")

	// Object 1 - Catalog
	obj1Start := content.Len()
	content.WriteString("1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n\n")

	// Object 2 - Pages
	obj2Start := content.Len()
	content.WriteString("2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n\n")

	// Object 3 - Page
	obj3Start := content.Len()
	content.WriteString("3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n/Contents 4 0 R\n>>\nendobj\n\n")

	// Object 4 - Contents with padding
	obj4Start := content.Len()
	content.WriteString(fmt.Sprintf("4 0 obj\n<<\n/Length %d\n>>\nstream\nBT\n/F1 12 Tf\n100 700 Td\n(%s) Tj\nET\nendstream\nendobj\n\n",
		len(padding)+20, padding))

	// xref table
	xrefStart := content.Len()
	content.WriteString("xref\n0 5\n")
	content.WriteString("0000000000 65535 f \n")
	content.WriteString(fmt.Sprintf("%010d 00000 n \n", obj1Start))
	content.WriteString(fmt.Sprintf("%010d 00000 n \n", obj2Start))
	content.WriteString(fmt.Sprintf("%010d 00000 n \n", obj3Start))
	content.WriteString(fmt.Sprintf("%010d 00000 n \n", obj4Start))

	// trailer
	content.WriteString("trailer\n<<\n/Size 5\n/Root 1 0 R\n>>\n")
	content.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF", xrefStart))

	return CreateTestFile(t, dir, filename, []byte(content.String()))
}

// MockErrorCallback 模拟错误回调
type MockErrorCallback struct {
	Called    bool
	Error     error
	CallCount int
}

func (m *MockErrorCallback) Callback(err error) {
	m.Called = true
	m.Error = err
	m.CallCount++
}

func (m *MockErrorCallback) Reset() {
	m.Called = false
	m.Error = nil
	m.CallCount = 0
}

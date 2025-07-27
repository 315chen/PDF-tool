//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/user/pdf-merger/pkg/file"
)

func main() {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "pdf_validation_test")
	if err != nil {
		log.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件验证器
	validator := file.NewFileValidator(tempDir)

	// 创建测试文件
	testFiles := createTestFiles(tempDir)

	fmt.Println("PDF文件验证示例")
	fmt.Println("================")

	// 验证每个测试文件
	for _, testFile := range testFiles {
		fmt.Printf("\n验证文件: %s\n", filepath.Base(testFile.path))
		fmt.Printf("描述: %s\n", testFile.description)

		result, err := validator.ValidateAndGetInfo(testFile.path)
		
		if result.IsValid {
			fmt.Println("✓ 文件验证通过")
			if result.FileInfo != nil {
				fmt.Printf("  文件大小: %d 字节\n", result.FileInfo.Size)
			}
			if result.PDFInfo != nil {
				fmt.Printf("  是否加密: %v\n", result.PDFInfo.IsEncrypted)
				fmt.Printf("  文件大小: %d 字节\n", result.PDFInfo.FileSize)
			}
		} else {
			fmt.Println("✗ 文件验证失败")
			if err != nil {
				fmt.Printf("  错误: %v\n", err)
			}
		}
	}
}

type testFile struct {
	path        string
	description string
}

func createTestFiles(tempDir string) []testFile {
	var files []testFile

	// 1. 有效的PDF文件
	validPDFContent := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000010 00000 n \n0000000079 00000 n \n0000000173 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n253\n%%EOF"
	validFile := filepath.Join(tempDir, "valid.pdf")
	os.WriteFile(validFile, []byte(validPDFContent), 0644)
	files = append(files, testFile{validFile, "有效的PDF文件"})

	// 2. 无效的文件头
	invalidFile := filepath.Join(tempDir, "invalid.pdf")
	os.WriteFile(invalidFile, []byte("NOT_A_PDF_FILE"), 0644)
	files = append(files, testFile{invalidFile, "无效的PDF文件头"})

	// 3. 空文件
	emptyFile := filepath.Join(tempDir, "empty.pdf")
	os.WriteFile(emptyFile, []byte(""), 0644)
	files = append(files, testFile{emptyFile, "空PDF文件"})

	// 4. 非PDF文件
	txtFile := filepath.Join(tempDir, "text.txt")
	os.WriteFile(txtFile, []byte("This is a text file"), 0644)
	files = append(files, testFile{txtFile, "非PDF文件（.txt）"})

	// 5. 包含加密标记的PDF
	encryptedPDFContent := "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Encrypt 2 0 R\n>>\nendobj\n%%EOF"
	encryptedFile := filepath.Join(tempDir, "encrypted.pdf")
	os.WriteFile(encryptedFile, []byte(encryptedPDFContent), 0644)
	files = append(files, testFile{encryptedFile, "包含加密标记的PDF"})

	return files
}
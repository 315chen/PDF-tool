//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pdf-merger/pkg/file"
	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("=== PDF文件验证和信息获取功能演示 ===\n")

	// 1. 演示文件管理器的基本验证功能
	demonstrateFileManagerValidation()

	// 2. 演示PDF验证器的高级功能
	demonstratePDFValidation()

	// 3. 演示PDF信息获取功能
	demonstratePDFInfoExtraction()

	// 4. 演示综合文件处理流程
	demonstrateComprehensiveFileProcessing()

	fmt.Println("\n=== 文件验证和信息获取演示完成 ===")
}

func demonstrateFileManagerValidation() {
	fmt.Println("1. 文件管理器基本验证功能演示:")
	
	// 创建文件管理器
	fm := file.NewFileManager("")
	
	// 创建测试文件
	tempDir, _ := os.MkdirTemp("", "file-validation-demo")
	defer os.RemoveAll(tempDir)
	
	// 1.1 创建有效的PDF文件
	validPDFPath := filepath.Join(tempDir, "valid.pdf")
	validPDFContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000074 00000 n \n0000000120 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n179\n%%EOF")
	os.WriteFile(validPDFPath, validPDFContent, 0644)
	
	// 1.2 创建无效文件
	invalidPath := filepath.Join(tempDir, "invalid.txt")
	os.WriteFile(invalidPath, []byte("This is not a PDF"), 0644)
	
	// 1.3 创建空PDF文件
	emptyPDFPath := filepath.Join(tempDir, "empty.pdf")
	os.WriteFile(emptyPDFPath, []byte(""), 0644)
	
	// 测试文件验证
	testCases := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"有效PDF文件", validPDFPath, true},
		{"无效文件格式", invalidPath, false},
		{"空PDF文件", emptyPDFPath, false},
		{"不存在的文件", "/nonexistent/file.pdf", false},
		{"空路径", "", false},
	}
	
	for _, tc := range testCases {
		err := fm.ValidateFile(tc.filePath)
		isValid := err == nil
		status := "✓"
		if !isValid {
			status = "✗"
		}
		fmt.Printf("   %s %s: %s", status, tc.name, tc.filePath)
		if !isValid {
			fmt.Printf(" (错误: %v)", err)
		}
		fmt.Println()
	}
	
	// 获取文件信息
	fmt.Println("\n   文件信息获取:")
	if info, err := fm.GetFileInfo(validPDFPath); err == nil {
		fmt.Printf("   - 文件名: %s\n", info.Name)
		fmt.Printf("   - 文件大小: %d 字节\n", info.Size)
		fmt.Printf("   - 文件路径: %s\n", info.Path)
		fmt.Printf("   - 是否有效: %t\n", info.IsValid)
	}
	
	fmt.Println()
}

func demonstratePDFValidation() {
	fmt.Println("2. PDF验证器高级功能演示:")
	
	// 创建PDF验证器
	validator := pdf.NewPDFValidator()
	
	// 创建测试文件
	tempDir, _ := os.MkdirTemp("", "pdf-validation-demo")
	defer os.RemoveAll(tempDir)
	
	// 2.1 创建各种测试PDF文件
	testFiles := map[string][]byte{
		"valid.pdf": []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000074 00000 n \n0000000120 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n179\n%%EOF"),
		"invalid_header.pdf": []byte("NOT-PDF-1.4\nSome content here"),
		"too_small.pdf": []byte("%PD"),
		"no_eof.pdf": []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj"),
		"encrypted.pdf": []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Encrypt 5 0 R\n>>\nendobj\n%%EOF"),
	}
	
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		os.WriteFile(filePath, content, 0644)
	}
	
	// 测试PDF验证
	fmt.Println("   PDF格式验证:")
	for filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := validator.ValidatePDFFile(filePath)
		status := "✓"
		if err != nil {
			status = "✗"
		}
		fmt.Printf("   %s %s", status, filename)
		if err != nil {
			fmt.Printf(" (错误: %v)", err)
		}
		fmt.Println()
	}
	
	// 测试加密检测
	fmt.Println("\n   加密状态检测:")
	for filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if info, err := validator.GetBasicPDFInfo(filePath); err == nil {
			fmt.Printf("   %s: 加密状态 = %t\n", filename, info.IsEncrypted)
		}
	}
	
	fmt.Println()
}

func demonstratePDFInfoExtraction() {
	fmt.Println("3. PDF信息获取功能演示:")
	
	// 创建PDF服务
	service := pdf.NewPDFService()
	
	// 创建测试文件
	tempDir, _ := os.MkdirTemp("", "pdf-info-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建一个更完整的PDF文件
	completePDFPath := filepath.Join(tempDir, "complete.pdf")
	completePDFContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Metadata 4 0 R
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
/Contents 5 0 R
>>
endobj
4 0 obj
<<
/Type /Metadata
/Subtype /XML
/Length 200
>>
stream
<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>测试PDF文档</dc:title>
<dc:creator>PDF合并工具</dc:creator>
</rdf:Description>
</rdf:RDF>
</x:xmpmeta>
endstream
endobj
5 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello World) Tj
ET
endstream
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000074 00000 n 
0000000120 00000 n 
0000000179 00000 n 
0000000565 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
659
%%EOF`)
	os.WriteFile(completePDFPath, completePDFContent, 0644)
	
	// 获取PDF信息
	fmt.Println("   PDF基本信息:")
	if info, err := service.GetPDFInfo(completePDFPath); err == nil {
		fmt.Printf("   - 文件路径: %s\n", info.FilePath)
		fmt.Printf("   - 页数: %d\n", info.PageCount)
		fmt.Printf("   - 文件大小: %d 字节\n", info.FileSize)
		fmt.Printf("   - 是否加密: %t\n", info.IsEncrypted)
		fmt.Printf("   - PDF版本: %s\n", info.Version)
		fmt.Printf("   - 标题: %s\n", info.Title)
		fmt.Printf("   - 作者: %s\n", info.Author)
		fmt.Printf("   - 创建时间: %s\n", info.CreationDate.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("   获取PDF信息失败: %v\n", err)
	}
	
	// 获取PDF元数据
	fmt.Println("\n   PDF元数据:")
	if metadata, err := service.GetPDFMetadata(completePDFPath); err == nil {
		for key, value := range metadata {
			fmt.Printf("   - %s: %s\n", key, value)
		}
	} else {
		fmt.Printf("   获取PDF元数据失败: %v\n", err)
	}
	
	// 检查加密状态
	fmt.Println("\n   加密状态检查:")
	if isEncrypted, err := service.IsPDFEncrypted(completePDFPath); err == nil {
		fmt.Printf("   - 文件是否加密: %t\n", isEncrypted)
	} else {
		fmt.Printf("   检查加密状态失败: %v\n", err)
	}
	
	fmt.Println()
}

func demonstrateComprehensiveFileProcessing() {
	fmt.Println("4. 综合文件处理流程演示:")
	
	// 创建文件管理器和PDF服务
	fm := file.NewFileManager("")
	service := pdf.NewPDFService()
	
	// 创建测试文件
	tempDir, _ := os.MkdirTemp("", "comprehensive-demo")
	defer os.RemoveAll(tempDir)
	
	// 创建测试PDF文件
	testPDFPath := filepath.Join(tempDir, "test.pdf")
	testPDFContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n>>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000074 00000 n \n0000000120 00000 n \ntrailer\n<<\n/Size 4\n/Root 1 0 R\n>>\nstartxref\n179\n%%EOF")
	os.WriteFile(testPDFPath, testPDFContent, 0644)
	
	fmt.Println("   完整的文件处理流程:")
	
	// 步骤1: 基本文件验证
	fmt.Printf("   步骤1: 基本文件验证...")
	if err := fm.ValidateFile(testPDFPath); err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Println(" 通过 ✓")
	
	// 步骤2: PDF格式验证
	fmt.Printf("   步骤2: PDF格式验证...")
	if err := service.ValidatePDF(testPDFPath); err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Println(" 通过 ✓")
	
	// 步骤3: 获取文件基本信息
	fmt.Printf("   步骤3: 获取文件基本信息...")
	fileInfo, err := fm.GetFileInfo(testPDFPath)
	if err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Printf(" 成功 ✓ (大小: %d 字节)\n", fileInfo.Size)
	
	// 步骤4: 获取PDF详细信息
	fmt.Printf("   步骤4: 获取PDF详细信息...")
	pdfInfo, err := service.GetPDFInfo(testPDFPath)
	if err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Printf(" 成功 ✓ (页数: %d, 加密: %t)\n", pdfInfo.PageCount, pdfInfo.IsEncrypted)
	
	// 步骤5: 创建临时副本
	fmt.Printf("   步骤5: 创建临时副本...")
	tempCopyPath, err := fm.CopyToTempFile(testPDFPath, "processed_")
	if err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Printf(" 成功 ✓ (临时文件: %s)\n", filepath.Base(tempCopyPath))
	
	// 步骤6: 验证临时副本
	fmt.Printf("   步骤6: 验证临时副本...")
	if err := service.ValidatePDF(tempCopyPath); err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Println(" 通过 ✓")
	
	// 步骤7: 清理临时文件
	fmt.Printf("   步骤7: 清理临时文件...")
	if err := fm.CleanupTempFiles(); err != nil {
		fmt.Printf(" 失败: %v\n", err)
		return
	}
	fmt.Println(" 完成 ✓")
	
	fmt.Println("\n   综合处理流程完成 🎉")
	fmt.Println("   所有验证步骤都已通过，文件处理正常")
	
	fmt.Println()
}

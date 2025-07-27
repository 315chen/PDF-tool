//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	fmt.Println("PDF读取器示例")
	fmt.Println("=============")

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "pdf-reader-example")
	fmt.Printf("创建临时目录: %s\n", tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := createTestFile(tempDir)
	fmt.Printf("\n创建测试文件: %s\n", filepath.Base(testFile))

	// 使用PDFReader读取文件
	fmt.Println("\n使用PDFReader读取文件:")
	reader, err := pdf.NewPDFReader(testFile)
	if err != nil {
		fmt.Printf("创建PDF读取器失败: %v\n", err)
		return
	}
	defer reader.Close()

	// 获取PDF信息
	info, err := reader.GetInfo()
	if err != nil {
		fmt.Printf("获取PDF信息失败: %v\n", err)
		return
	}

	// 显示PDF信息
	fmt.Printf("  页数: %d\n", info.PageCount)
	fmt.Printf("  文件大小: %d 字节\n", info.FileSize)
	fmt.Printf("  标题: %s\n", info.Title)
	fmt.Printf("  是否加密: %v\n", info.IsEncrypted)

	// 验证PDF结构
	fmt.Println("\n验证PDF结构:")
	err = reader.ValidateStructure()
	if err != nil {
		fmt.Printf("结构验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 结构验证通过")
	}

	// 获取元数据
	fmt.Println("\n获取PDF元数据:")
	metadata, err := reader.GetMetadata()
	if err != nil {
		fmt.Printf("获取元数据失败: %v\n", err)
	} else {
		if len(metadata) == 0 {
			fmt.Println("  没有可用的元数据")
		} else {
			for key, value := range metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// 使用PDFService
	fmt.Println("\n使用PDFService接口:")
	pdfService := pdf.NewPDFService()

	// 验证PDF文件
	fmt.Println("验证PDF文件:")
	err = pdfService.ValidatePDF(testFile)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 文件验证通过")
	}

	// 验证PDF结构
	fmt.Println("\n验证PDF结构:")
	err = pdfService.ValidatePDFStructure(testFile)
	if err != nil {
		fmt.Printf("结构验证失败: %v\n", err)
	} else {
		fmt.Println("✓ 结构验证通过")
	}

	// 获取PDF元数据
	fmt.Println("\n获取PDF元数据:")
	serviceMetadata, err := pdfService.GetPDFMetadata(testFile)
	if err != nil {
		fmt.Printf("获取元数据失败: %v\n", err)
	} else {
		if len(serviceMetadata) == 0 {
			fmt.Println("  没有可用的元数据")
		} else {
			for key, value := range serviceMetadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	fmt.Println("\n示例完成")
}

// createTestFile 创建一个测试PDF文件
func createTestFile(tempDir string) string {
	// 创建一个有效的PDF内容
	validPDFContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
/Info 5 0 R
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
/F1 24 Tf
100 700 Td
(测试PDF文件) Tj
ET
endstream
endobj
5 0 obj
<<
/Title (测试PDF文档)
/Author (PDF合并工具)
/Subject (PDF读取器测试)
/Creator (PDF合并工具测试程序)
/Producer (PDF合并工具)
/CreationDate (D:20250724000000+08'00')
>>
endobj
xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000079 00000 n 
0000000141 00000 n 
0000000226 00000 n 
0000000321 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
/Info 5 0 R
>>
startxref
500
%%EOF`

	// 写入文件
	filePath := filepath.Join(tempDir, "test.pdf")
	os.WriteFile(filePath, []byte(validPDFContent), 0644)
	return filePath
}
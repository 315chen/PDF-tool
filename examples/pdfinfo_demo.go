//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/user/pdf-merger/pkg/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run pdfinfo_demo.go <pdf-file>")
		os.Exit(1)
	}

	pdfFile := os.Args[1]

	// 创建PDF服务
	service := pdf.NewPDFServiceImpl()

	// 获取PDF信息
	info, err := service.GetPDFInfo(pdfFile)
	if err != nil {
		log.Fatalf("Failed to get PDF info: %v", err)
	}

	// 显示基本信息
	fmt.Printf("=== PDF文件信息 ===\n")
	fmt.Printf("文件路径: %s\n", info.FilePath)
	fmt.Printf("文件大小: %s\n", info.GetFormattedSize())
	fmt.Printf("页数: %d\n", info.PageCount)
	fmt.Printf("PDF版本: %s\n", info.Version)
	fmt.Printf("是否加密: %t\n", info.IsEncrypted)

	// 显示元数据
	fmt.Printf("\n=== 元数据 ===\n")
	metadata := info.GetMetadataMap()
	if len(metadata) > 0 {
		for key, value := range metadata {
			fmt.Printf("%s: %s\n", key, value)
		}
	} else {
		fmt.Println("无元数据")
	}

	// 显示pdfcpu特有信息
	fmt.Printf("\n=== PDFCPU信息 ===\n")
	if info.PDFCPUVersion != "" {
		fmt.Printf("PDFCPU版本: %s\n", info.PDFCPUVersion)
	}

	// 显示加密信息
	if info.IsEncrypted {
		fmt.Printf("\n=== 加密信息 ===\n")
		encInfo := info.GetEncryptionInfo()
		if method, ok := encInfo["method"].(string); ok && method != "" {
			fmt.Printf("加密方法: %s\n", method)
		}
		if keyLen, ok := encInfo["key_length"].(int); ok && keyLen > 0 {
			fmt.Printf("密钥长度: %d位\n", keyLen)
		}
		if userPwd, ok := encInfo["user_password"].(bool); ok {
			fmt.Printf("用户密码: %t\n", userPwd)
		}
		if ownerPwd, ok := encInfo["owner_password"].(bool); ok {
			fmt.Printf("所有者密码: %t\n", ownerPwd)
		}
	}

	// 显示权限信息
	fmt.Printf("\n=== 权限信息 ===\n")
	if info.HasRestrictedPermissions() {
		fmt.Println("文档有权限限制")
		permFlags := info.GetPermissionFlags()
		fmt.Printf("打印: %t\n", permFlags["print"])
		fmt.Printf("修改: %t\n", permFlags["modify"])
		fmt.Printf("复制: %t\n", permFlags["copy"])
		fmt.Printf("注释: %t\n", permFlags["annotate"])
		fmt.Printf("填写表单: %t\n", permFlags["fill_forms"])
		fmt.Printf("提取内容: %t\n", permFlags["extract"])
		fmt.Printf("组装文档: %t\n", permFlags["assemble"])
		fmt.Printf("高质量打印: %t\n", permFlags["print_high_quality"])
	} else {
		fmt.Println("文档无权限限制")
	}

	// 显示权限摘要
	fmt.Printf("\n权限摘要: %s\n", info.GetPermissionSummary())

	// 显示原始权限列表
	if len(info.Permissions) > 0 {
		fmt.Printf("原始权限列表: %v\n", info.Permissions)
	}

	// 验证信息完整性
	fmt.Printf("\n=== 验证 ===\n")
	fmt.Printf("信息有效: %t\n", info.IsValid())
	fmt.Printf("包含元数据: %t\n", info.HasMetadata())

	fmt.Printf("\n=== 演示完成 ===\n")
}
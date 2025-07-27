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
		fmt.Println("Usage: go run permissions_demo.go <pdf-file>")
		os.Exit(1)
	}

	pdfFile := os.Args[1]

	// 创建PDF读取器
	reader, err := pdf.NewPDFReader(pdfFile)
	if err != nil {
		log.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	fmt.Printf("=== PDF权限和安全信息分析 ===\n")
	fmt.Printf("文件: %s\n\n", pdfFile)

	// 检查基本权限
	fmt.Printf("=== 基本权限检查 ===\n")
	permissions, err := reader.CheckPermissions()
	if err != nil {
		log.Fatalf("Failed to check permissions: %v", err)
	}

	fmt.Printf("权限列表: %v\n", permissions)
	fmt.Printf("权限数量: %d/8\n\n", len(permissions))

	// 检查具体权限
	fmt.Printf("=== 详细权限检查 ===\n")
	permissionChecks := []struct {
		name   string
		method func() (bool, error)
	}{
		{"打印", reader.CanPrint},
		{"修改", reader.CanModify},
		{"复制", reader.CanCopy},
		{"注释", reader.CanAnnotate},
		{"填写表单", reader.CanFillForms},
		{"提取内容", reader.CanExtract},
		{"组装文档", reader.CanAssemble},
		{"高质量打印", reader.CanPrintHighQuality},
	}

	for _, check := range permissionChecks {
		allowed, err := check.method()
		if err != nil {
			fmt.Printf("%s: 检查失败 (%v)\n", check.name, err)
		} else {
			status := "❌ 禁止"
			if allowed {
				status = "✅ 允许"
			}
			fmt.Printf("%s: %s\n", check.name, status)
		}
	}

	// 获取安全信息
	fmt.Printf("\n=== 安全信息 ===\n")
	securityInfo, err := reader.GetSecurityInfo()
	if err != nil {
		log.Fatalf("Failed to get security info: %v", err)
	}

	if encrypted, ok := securityInfo["encrypted"].(bool); ok {
		fmt.Printf("加密状态: %t\n", encrypted)
		
		if encrypted {
			if version, ok := securityInfo["version"].(int); ok && version > 0 {
				fmt.Printf("加密版本: %d\n", version)
			}
			if revision, ok := securityInfo["revision"].(int); ok && revision > 0 {
				fmt.Printf("加密修订版: %d\n", revision)
			}
			if keyLength, ok := securityInfo["key_length"].(int); ok && keyLength > 0 {
				fmt.Printf("密钥长度: %d位\n", keyLength)
			}
			if handler, ok := securityInfo["security_handler"].(string); ok && handler != "" {
				fmt.Printf("安全处理器: %s\n", handler)
			}
			if filter, ok := securityInfo["filter"].(string); ok && filter != "" {
				fmt.Printf("过滤器: %s\n", filter)
			}
			
			if hasUserPwd, ok := securityInfo["has_user_password"].(bool); ok {
				fmt.Printf("用户密码: %t\n", hasUserPwd)
			}
			if hasOwnerPwd, ok := securityInfo["has_owner_password"].(bool); ok {
				fmt.Printf("所有者密码: %t\n", hasOwnerPwd)
			}
		}
	}

	// 获取详细安全信息
	fmt.Printf("\n=== 详细安全分析 ===\n")
	detailedInfo, err := reader.GetDetailedSecurityInfo()
	if err != nil {
		log.Fatalf("Failed to get detailed security info: %v", err)
	}

	if securityLevel, ok := detailedInfo["security_level"].(string); ok {
		fmt.Printf("安全级别: %s\n", securityLevel)
	}

	// 权限摘要
	if permSummary, ok := detailedInfo["permission_summary"].(map[string]interface{}); ok {
		fmt.Printf("\n=== 权限摘要 ===\n")
		if totalPerms, ok := permSummary["total_permissions"].(int); ok {
			fmt.Printf("总权限数: %d\n", totalPerms)
		}
		if restrictionLevel, ok := permSummary["restriction_level"].(float64); ok {
			fmt.Printf("限制程度: %.1f%%\n", restrictionLevel)
		}
		
		fmt.Printf("\n权限详情:\n")
		permissionDetails := []struct {
			key  string
			name string
		}{
			{"can_print", "打印"},
			{"can_modify", "修改"},
			{"can_copy", "复制"},
			{"can_annotate", "注释"},
			{"can_fill_forms", "填写表单"},
			{"can_extract", "提取内容"},
			{"can_assemble", "组装文档"},
			{"can_print_high_quality", "高质量打印"},
		}
		
		for _, detail := range permissionDetails {
			if allowed, ok := permSummary[detail.key].(bool); ok {
				status := "❌"
				if allowed {
					status = "✅"
				}
				fmt.Printf("  %s %s\n", status, detail.name)
			}
		}
	}

	// 安全建议
	if recommendations, ok := detailedInfo["security_recommendations"].([]string); ok && len(recommendations) > 0 {
		fmt.Printf("\n=== 安全建议 ===\n")
		for i, rec := range recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}

	// 证书和签名信息（如果有）
	if certInfo, ok := securityInfo["certificate_info"].(map[string]interface{}); ok {
		if hasCerts, ok := certInfo["has_certificates"].(bool); ok && hasCerts {
			fmt.Printf("\n=== 证书信息 ===\n")
			if certCount, ok := certInfo["certificate_count"].(int); ok {
				fmt.Printf("证书数量: %d\n", certCount)
			}
		}
	}

	if sigInfo, ok := securityInfo["signature_info"].(map[string]interface{}); ok {
		if hasSigs, ok := sigInfo["has_signatures"].(bool); ok && hasSigs {
			fmt.Printf("\n=== 数字签名信息 ===\n")
			if sigCount, ok := sigInfo["signature_count"].(int); ok {
				fmt.Printf("签名数量: %d\n", sigCount)
			}
		}
	}

	fmt.Printf("\n=== 分析完成 ===\n")
}
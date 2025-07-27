package pdf

import (
	"testing"
)

func TestNewPDFInfo(t *testing.T) {
	filePath := "/test/path/document.pdf"
	info := NewPDFInfo(filePath)

	// 验证基本字段
	if info.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, info.FilePath)
	}

	if info.PageCount != 0 {
		t.Errorf("Expected PageCount 0, got %d", info.PageCount)
	}

	if info.IsEncrypted != false {
		t.Errorf("Expected IsEncrypted false, got %t", info.IsEncrypted)
	}

	// 验证默认权限设置
	if !info.PrintAllowed {
		t.Error("Expected PrintAllowed to be true by default")
	}

	if !info.ModifyAllowed {
		t.Error("Expected ModifyAllowed to be true by default")
	}

	if !info.CopyAllowed {
		t.Error("Expected CopyAllowed to be true by default")
	}

	// 验证权限切片初始化
	if info.Permissions == nil {
		t.Error("Expected Permissions slice to be initialized")
	}

	if len(info.Permissions) != 0 {
		t.Errorf("Expected empty Permissions slice, got length %d", len(info.Permissions))
	}
}

func TestMapPDFInfo(t *testing.T) {
	filePath := "/test/path/document.pdf"
	basicInfo := map[string]interface{}{
		"PageCount":     10,
		"IsEncrypted":   true,
		"FileSize":      int64(1024000),
		"Title":         "Test Document",
		"Author":        "Test Author",
		"Version":       "1.7",
		"PDFCPUVersion": "0.3.13",
		"Permissions":   []string{"print", "copy"},
		"Keywords":      "test, document",
		"KeyLength":     128,
		"UserPassword":  true,
		"PrintAllowed":  true,
		"CopyAllowed":   true,
		"ModifyAllowed": false,
	}

	info := mapPDFInfo(filePath, basicInfo)

	// 验证基本映射
	if info.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, info.FilePath)
	}

	if info.PageCount != 10 {
		t.Errorf("Expected PageCount 10, got %d", info.PageCount)
	}

	if !info.IsEncrypted {
		t.Error("Expected IsEncrypted to be true")
	}

	if info.FileSize != 1024000 {
		t.Errorf("Expected FileSize 1024000, got %d", info.FileSize)
	}

	if info.Title != "Test Document" {
		t.Errorf("Expected Title 'Test Document', got '%s'", info.Title)
	}

	if info.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", info.Author)
	}

	if info.Version != "1.7" {
		t.Errorf("Expected Version '1.7', got '%s'", info.Version)
	}

	// 验证pdfcpu特有字段
	if info.PDFCPUVersion != "0.3.13" {
		t.Errorf("Expected PDFCPUVersion '0.3.13', got '%s'", info.PDFCPUVersion)
	}

	if len(info.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(info.Permissions))
	}

	if info.Keywords != "test, document" {
		t.Errorf("Expected Keywords 'test, document', got '%s'", info.Keywords)
	}

	if info.KeyLength != 128 {
		t.Errorf("Expected KeyLength 128, got %d", info.KeyLength)
	}

	if !info.UserPassword {
		t.Error("Expected UserPassword to be true")
	}

	if !info.PrintAllowed {
		t.Error("Expected PrintAllowed to be true")
	}

	if !info.CopyAllowed {
		t.Error("Expected CopyAllowed to be true")
	}

	if info.ModifyAllowed {
		t.Error("Expected ModifyAllowed to be false")
	}
}

func TestMapPDFCPUInfo(t *testing.T) {
	filePath := "/test/path/document.pdf"
	pdfcpuOutput := `PDF version: 1.7
Page count: 5
Encrypted: true
Title: Sample Document
Author: John Doe
Subject: Test Subject
Creator: Test Creator
Producer: Test Producer
Keywords: sample, test, document
Trapped: False
Encryption method: AES-256
Key length: 256
User password: true
Owner password: true
Permissions: print, copy, annotate`

	info := mapPDFCPUInfo(filePath, pdfcpuOutput)

	// 验证解析结果
	if info.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, info.FilePath)
	}

	if info.Version != "1.7" {
		t.Errorf("Expected Version '1.7', got '%s'", info.Version)
	}

	if info.PageCount != 5 {
		t.Errorf("Expected PageCount 5, got %d", info.PageCount)
	}

	if !info.IsEncrypted {
		t.Error("Expected IsEncrypted to be true")
	}

	if info.Title != "Sample Document" {
		t.Errorf("Expected Title 'Sample Document', got '%s'", info.Title)
	}

	if info.Author != "John Doe" {
		t.Errorf("Expected Author 'John Doe', got '%s'", info.Author)
	}

	if info.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got '%s'", info.Subject)
	}

	if info.Creator != "Test Creator" {
		t.Errorf("Expected Creator 'Test Creator', got '%s'", info.Creator)
	}

	if info.Producer != "Test Producer" {
		t.Errorf("Expected Producer 'Test Producer', got '%s'", info.Producer)
	}

	if info.Keywords != "sample, test, document" {
		t.Errorf("Expected Keywords 'sample, test, document', got '%s'", info.Keywords)
	}

	if info.Trapped != "False" {
		t.Errorf("Expected Trapped 'False', got '%s'", info.Trapped)
	}

	if info.EncryptionMethod != "AES-256" {
		t.Errorf("Expected EncryptionMethod 'AES-256', got '%s'", info.EncryptionMethod)
	}

	if info.KeyLength != 256 {
		t.Errorf("Expected KeyLength 256, got %d", info.KeyLength)
	}

	if !info.UserPassword {
		t.Error("Expected UserPassword to be true")
	}

	if !info.OwnerPassword {
		t.Error("Expected OwnerPassword to be true")
	}

	// 验证权限解析
	expectedPermissions := []string{"print", "copy", "annotate"}
	if len(info.Permissions) != len(expectedPermissions) {
		t.Errorf("Expected %d permissions, got %d", len(expectedPermissions), len(info.Permissions))
	}

	for i, expected := range expectedPermissions {
		if i < len(info.Permissions) && info.Permissions[i] != expected {
			t.Errorf("Expected permission[%d] '%s', got '%s'", i, expected, info.Permissions[i])
		}
	}

	// 验证权限标志
	if !info.PrintAllowed {
		t.Error("Expected PrintAllowed to be true")
	}

	if !info.CopyAllowed {
		t.Error("Expected CopyAllowed to be true")
	}

	if !info.AnnotateAllowed {
		t.Error("Expected AnnotateAllowed to be true")
	}

	// 这些权限不在列表中，应该为false
	if info.ModifyAllowed {
		t.Error("Expected ModifyAllowed to be false")
	}

	if info.FillFormsAllowed {
		t.Error("Expected FillFormsAllowed to be false")
	}
}

func TestPDFInfoHelperMethods(t *testing.T) {
	info := NewPDFInfo("/test/document.pdf")
	info.Title = "Test Document"
	info.Author = "Test Author"
	info.IsEncrypted = true
	info.EncryptionMethod = "AES-128"
	info.KeyLength = 128
	info.UserPassword = true
	info.OwnerPassword = false
	info.PrintAllowed = true
	info.ModifyAllowed = false
	info.CopyAllowed = true

	// 测试IsValid方法
	info.PageCount = 10
	info.FileSize = 1024
	if !info.IsValid() {
		t.Error("Expected IsValid to return true")
	}

	// 测试HasMetadata方法
	if !info.HasMetadata() {
		t.Error("Expected HasMetadata to return true")
	}

	// 测试GetEncryptionInfo方法
	encInfo := info.GetEncryptionInfo()
	if encInfo["encrypted"] != true {
		t.Error("Expected encryption info to show encrypted=true")
	}

	if encInfo["method"] != "AES-128" {
		t.Errorf("Expected encryption method 'AES-128', got '%v'", encInfo["method"])
	}

	if encInfo["key_length"] != 128 {
		t.Errorf("Expected key length 128, got %v", encInfo["key_length"])
	}

	// 测试GetPermissionFlags方法
	permFlags := info.GetPermissionFlags()
	if permFlags["print"] != true {
		t.Error("Expected print permission to be true")
	}

	if permFlags["modify"] != false {
		t.Error("Expected modify permission to be false")
	}

	if permFlags["copy"] != true {
		t.Error("Expected copy permission to be true")
	}

	// 测试HasRestrictedPermissions方法
	if !info.HasRestrictedPermissions() {
		t.Error("Expected HasRestrictedPermissions to return true")
	}

	// 测试GetMetadataMap方法
	metadata := info.GetMetadataMap()
	if metadata["Title"] != "Test Document" {
		t.Errorf("Expected metadata Title 'Test Document', got '%s'", metadata["Title"])
	}

	if metadata["Author"] != "Test Author" {
		t.Errorf("Expected metadata Author 'Test Author', got '%s'", metadata["Author"])
	}
}

func TestPDFInfoClone(t *testing.T) {
	original := NewPDFInfo("/test/document.pdf")
	original.Title = "Original Title"
	original.PageCount = 10
	original.Permissions = []string{"print", "copy"}

	clone := original.Clone()

	// 验证克隆的基本字段
	if clone.Title != original.Title {
		t.Errorf("Expected cloned Title '%s', got '%s'", original.Title, clone.Title)
	}

	if clone.PageCount != original.PageCount {
		t.Errorf("Expected cloned PageCount %d, got %d", original.PageCount, clone.PageCount)
	}

	// 验证权限切片的深拷贝
	if len(clone.Permissions) != len(original.Permissions) {
		t.Errorf("Expected cloned Permissions length %d, got %d", len(original.Permissions), len(clone.Permissions))
	}

	// 修改原始对象的权限，不应该影响克隆
	original.Permissions[0] = "modified"
	if clone.Permissions[0] == "modified" {
		t.Error("Clone should have independent Permissions slice")
	}
}

func TestUpdateFromPDFCPU(t *testing.T) {
	info := NewPDFInfo("/test/document.pdf")

	pdfcpuInfo := map[string]interface{}{
		"pdfcpu_version":    "0.3.13",
		"permissions":       []string{"print", "copy", "annotate"},
		"encryption_method": "AES-256",
		"key_length":        256,
		"user_password":     true,
		"owner_password":    false,
	}

	info.UpdateFromPDFCPU(pdfcpuInfo)

	// 验证更新结果
	if info.PDFCPUVersion != "0.3.13" {
		t.Errorf("Expected PDFCPUVersion '0.3.13', got '%s'", info.PDFCPUVersion)
	}

	if info.EncryptionMethod != "AES-256" {
		t.Errorf("Expected EncryptionMethod 'AES-256', got '%s'", info.EncryptionMethod)
	}

	if info.KeyLength != 256 {
		t.Errorf("Expected KeyLength 256, got %d", info.KeyLength)
	}

	if !info.UserPassword {
		t.Error("Expected UserPassword to be true")
	}

	if info.OwnerPassword {
		t.Error("Expected OwnerPassword to be false")
	}

	// 验证权限更新
	expectedPermissions := []string{"print", "copy", "annotate"}
	if len(info.Permissions) != len(expectedPermissions) {
		t.Errorf("Expected %d permissions, got %d", len(expectedPermissions), len(info.Permissions))
	}

	// 验证权限标志更新
	if !info.PrintAllowed {
		t.Error("Expected PrintAllowed to be true")
	}

	if !info.CopyAllowed {
		t.Error("Expected CopyAllowed to be true")
	}

	if !info.AnnotateAllowed {
		t.Error("Expected AnnotateAllowed to be true")
	}

	if info.ModifyAllowed {
		t.Error("Expected ModifyAllowed to be false (not in permissions)")
	}
}

func TestExtractStringValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Title: Sample Document", "Sample Document"},
		{"Author: John Doe", "John Doe"},
		{"Version: 1.7", "1.7"},
		{"Empty:", ""},
		{"No colon", ""},
		{"Multiple: colons: here", "colons: here"},
	}

	for _, test := range tests {
		result := extractStringValue(test.input)
		if result != test.expected {
			t.Errorf("extractStringValue(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestExtractIntValue(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		shouldError bool
	}{
		{"Page count: 10", 10, false},
		{"Key length: 256", 256, false},
		{"Invalid: abc", 0, true},
		{"Empty:", 0, true},
		{"No colon", 0, true},
	}

	for _, test := range tests {
		result, err := extractIntValue(test.input)
		if test.shouldError {
			if err == nil {
				t.Errorf("extractIntValue(%q) should have returned an error", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("extractIntValue(%q) returned unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("extractIntValue(%q) = %d, expected %d", test.input, result, test.expected)
			}
		}
	}
}

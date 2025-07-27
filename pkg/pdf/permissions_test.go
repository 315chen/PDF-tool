package pdf

import (
	"testing"
)

func TestPDFReader_CheckPermissions(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "permissions_test")
	testFile := createTestPDFFile(t, tempDir, "permissions_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	permissions, err := reader.CheckPermissions()
	if err != nil {
		t.Fatalf("Failed to check permissions: %v", err)
	}

	// 对于未加密的测试文件，应该有所有权限
	expectedPermissions := []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"}
	if len(permissions) != len(expectedPermissions) {
		t.Errorf("Expected %d permissions, got %d", len(expectedPermissions), len(permissions))
	}

	// 验证每个权限都存在
	permissionMap := make(map[string]bool)
	for _, perm := range permissions {
		permissionMap[perm] = true
	}

	for _, expected := range expectedPermissions {
		if !permissionMap[expected] {
			t.Errorf("Expected permission '%s' not found", expected)
		}
	}
}

func TestPDFReader_IndividualPermissionChecks(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "individual_permissions_test")
	testFile := createTestPDFFile(t, tempDir, "individual_permissions_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	// 测试各种权限检查方法
	tests := []struct {
		name   string
		method func() (bool, error)
	}{
		{"CanPrint", reader.CanPrint},
		{"CanModify", reader.CanModify},
		{"CanCopy", reader.CanCopy},
		{"CanAnnotate", reader.CanAnnotate},
		{"CanFillForms", reader.CanFillForms},
		{"CanExtract", reader.CanExtract},
		{"CanAssemble", reader.CanAssemble},
		{"CanPrintHighQuality", reader.CanPrintHighQuality},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			canDo, err := test.method()
			if err != nil {
				t.Errorf("Error checking %s: %v", test.name, err)
			}

			// 对于未加密的测试文件，所有权限都应该为true
			if !canDo {
				t.Errorf("Expected %s to be true for unencrypted PDF", test.name)
			}
		})
	}
}

func TestPDFReader_GetSecurityInfo(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "security_info_test")
	testFile := createTestPDFFile(t, tempDir, "security_info_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	securityInfo, err := reader.GetSecurityInfo()
	if err != nil {
		t.Fatalf("Failed to get security info: %v", err)
	}

	// 验证基本安全信息
	encrypted, ok := securityInfo["encrypted"].(bool)
	if !ok {
		t.Error("Expected 'encrypted' field to be boolean")
	}

	// 对于测试文件，应该是未加密的
	if encrypted {
		t.Error("Expected test PDF to be unencrypted")
	}

	// 验证权限信息
	permissions, ok := securityInfo["permissions"].([]string)
	if !ok {
		t.Error("Expected 'permissions' field to be string slice")
	}

	if len(permissions) == 0 {
		t.Error("Expected permissions to be non-empty")
	}

	// 验证密码状态
	hasUserPwd, ok := securityInfo["has_user_password"].(bool)
	if !ok {
		t.Error("Expected 'has_user_password' field to be boolean")
	}

	if hasUserPwd {
		t.Error("Expected test PDF to have no user password")
	}

	hasOwnerPwd, ok := securityInfo["has_owner_password"].(bool)
	if !ok {
		t.Error("Expected 'has_owner_password' field to be boolean")
	}

	if hasOwnerPwd {
		t.Error("Expected test PDF to have no owner password")
	}
}

func TestPDFReader_GetDetailedSecurityInfo(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "detailed_security_test")
	testFile := createTestPDFFile(t, tempDir, "detailed_security_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	detailedInfo, err := reader.GetDetailedSecurityInfo()
	if err != nil {
		t.Fatalf("Failed to get detailed security info: %v", err)
	}

	// 验证安全级别
	securityLevel, ok := detailedInfo["security_level"].(string)
	if !ok {
		t.Error("Expected 'security_level' field to be string")
	}

	if securityLevel != "无保护" {
		t.Errorf("Expected security level '无保护', got '%s'", securityLevel)
	}

	// 验证权限摘要
	permissionSummary, ok := detailedInfo["permission_summary"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'permission_summary' field to be map")
	}

	totalPermissions, ok := permissionSummary["total_permissions"].(int)
	if !ok {
		t.Error("Expected 'total_permissions' to be int")
	}

	if totalPermissions != 8 {
		t.Errorf("Expected 8 total permissions, got %d", totalPermissions)
	}

	// 验证安全建议
	recommendations, ok := detailedInfo["security_recommendations"].([]string)
	if !ok {
		t.Error("Expected 'security_recommendations' field to be string slice")
	}

	if len(recommendations) == 0 {
		t.Error("Expected security recommendations to be non-empty")
	}
}

func TestPDFReader_PermissionMethods_ClosedReader(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "closed_reader_test")
	testFile := createTestPDFFile(t, tempDir, "closed_reader_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}

	// 关闭读取器
	reader.Close()

	// 测试在关闭状态下调用权限方法
	_, err = reader.CheckPermissions()
	if err == nil {
		t.Error("Expected error when checking permissions on closed reader")
	}

	_, err = reader.GetSecurityInfo()
	if err == nil {
		t.Error("Expected error when getting security info on closed reader")
	}

	_, err = reader.GetDetailedSecurityInfo()
	if err == nil {
		t.Error("Expected error when getting detailed security info on closed reader")
	}

	_, err = reader.CanPrint()
	if err == nil {
		t.Error("Expected error when checking print permission on closed reader")
	}
}

func TestPDFCPUCLIAdapter_GetPermissions(t *testing.T) {
	adapter, err := NewPDFCPUCLIAdapter()
	if err != nil {
		t.Skipf("Skipping test: pdfcpu CLI not available: %v", err)
	}
	defer adapter.Close()

	if !adapter.IsAvailable() {
		t.Skip("Skipping test: pdfcpu CLI not available")
	}

	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "cli_permissions_test")
	testFile := createTestPDFFile(t, tempDir, "cli_permissions_test.pdf")

	permissions, err := adapter.GetPermissions(testFile)
	if err != nil {
		t.Fatalf("Failed to get permissions: %v", err)
	}

	// 验证返回的权限信息结构
	if permissions == nil {
		t.Fatal("Expected permissions map to be non-nil")
	}

	// 验证必需的字段
	requiredFields := []string{"encrypted", "permissions", "has_user_password", "has_owner_password"}
	for _, field := range requiredFields {
		if _, exists := permissions[field]; !exists {
			t.Errorf("Expected field '%s' to exist in permissions", field)
		}
	}
}

func TestPDFCPUCLIAdapter_GetSecurityDetails(t *testing.T) {
	adapter, err := NewPDFCPUCLIAdapter()
	if err != nil {
		t.Skipf("Skipping test: pdfcpu CLI not available: %v", err)
	}
	defer adapter.Close()

	if !adapter.IsAvailable() {
		t.Skip("Skipping test: pdfcpu CLI not available")
	}

	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "cli_security_test")
	testFile := createTestPDFFile(t, tempDir, "cli_security_test.pdf")

	securityDetails, err := adapter.GetSecurityDetails(testFile)
	if err != nil {
		t.Fatalf("Failed to get security details: %v", err)
	}

	// 验证返回的安全详情结构
	if securityDetails == nil {
		t.Fatal("Expected security details map to be non-nil")
	}

	// 验证基本字段
	encrypted, ok := securityDetails["encrypted"].(bool)
	if !ok {
		t.Error("Expected 'encrypted' field to be boolean")
	}

	// 对于测试文件，应该是未加密的
	if encrypted {
		t.Error("Expected test PDF to be unencrypted")
	}
}

func TestParsePermissionFlags(t *testing.T) {
	adapter := &PDFCPUCLIAdapter{}

	tests := []struct {
		flags    int
		expected []string
	}{
		{0, []string{}},                   // 无权限
		{4, []string{"print"}},            // 只有打印权限 (位2)
		{12, []string{"print", "modify"}}, // 打印和修改权限 (位2和3)
		{-4, []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"}}, // 所有权限
	}

	for _, test := range tests {
		result := adapter.parsePermissionFlags(test.flags)

		if len(result) != len(test.expected) {
			t.Errorf("For flags %d, expected %d permissions, got %d",
				test.flags, len(test.expected), len(result))
			continue
		}

		// 创建映射以便比较
		resultMap := make(map[string]bool)
		for _, perm := range result {
			resultMap[perm] = true
		}

		for _, expected := range test.expected {
			if !resultMap[expected] {
				t.Errorf("For flags %d, expected permission '%s' not found", test.flags, expected)
			}
		}
	}
}

func TestAnalyzeSecurityLevel(t *testing.T) {
	// 创建测试目录和PDF文件
	tempDir := createTempDir(t, "security_level_test")
	testFile := createTestPDFFile(t, tempDir, "security_level_test.pdf")

	reader, err := NewPDFReader(testFile)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	tests := []struct {
		name         string
		securityInfo map[string]interface{}
		expected     string
	}{
		{
			name: "无加密",
			securityInfo: map[string]interface{}{
				"encrypted": false,
			},
			expected: "无保护",
		},
		{
			name: "高级加密",
			securityInfo: map[string]interface{}{
				"encrypted":  true,
				"key_length": 256,
			},
			expected: "高级加密",
		},
		{
			name: "标准加密",
			securityInfo: map[string]interface{}{
				"encrypted":  true,
				"key_length": 128,
			},
			expected: "标准加密",
		},
		{
			name: "中级加密",
			securityInfo: map[string]interface{}{
				"encrypted": true,
				"version":   4,
			},
			expected: "中级加密",
		},
		{
			name: "基础加密",
			securityInfo: map[string]interface{}{
				"encrypted": true,
				"version":   2,
			},
			expected: "基础加密",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := reader.analyzeSecurityLevel(test.securityInfo)
			if result != test.expected {
				t.Errorf("Expected security level '%s', got '%s'", test.expected, result)
			}
		})
	}
}

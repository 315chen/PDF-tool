package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PDFReader 提供增强的PDF读取功能
type PDFReader struct {
	filePath   string
	info       *PDFInfo
	isOpen     bool
	cliAdapter *PDFCPUCLIAdapter
	useCLI     bool
}

// NewPDFReader 创建一个新的PDF读取器
func NewPDFReader(filePath string) (*PDFReader, error) {
	reader := &PDFReader{
		filePath: filePath,
		isOpen:   false,
		useCLI:   false,
	}

	// 尝试初始化CLI适配器
	if cliAdapter, err := NewPDFCPUCLIAdapter(); err == nil && cliAdapter.IsAvailable() {
		reader.cliAdapter = cliAdapter
		reader.useCLI = true
	}

	if err := reader.Open(); err != nil {
		return nil, err
	}

	return reader, nil
}

// Open 打开PDF文件进行读取
func (r *PDFReader) Open() error {
	if r.isOpen {
		return nil
	}

	// 验证文件是否存在
	if _, err := os.Stat(r.filePath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法访问PDF文件",
			File:    r.filePath,
			Cause:   err,
		}
	}

	// 基本PDF文件验证
	if err := r.basicPDFValidation(); err != nil {
		return err
	}

	// 如果使用CLI，验证文件
	if r.useCLI && r.cliAdapter != nil {
		if err := r.cliAdapter.ValidateFile(r.filePath); err != nil {
			return &PDFError{
				Type:    ErrorInvalidFile,
				Message: "PDF文件验证失败",
				File:    r.filePath,
				Cause:   err,
			}
		}
	}

	r.isOpen = true
	return nil
}

// Close 关闭PDF读取器并释放资源
func (r *PDFReader) Close() error {
	if !r.isOpen {
		return nil
	}

	// 关闭CLI适配器
	if r.cliAdapter != nil {
		r.cliAdapter.Close()
	}

	r.info = nil
	r.isOpen = false

	return nil
}

// GetInfo 获取PDF文件的详细信息
func (r *PDFReader) GetInfo() (*PDFInfo, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	if r.info != nil {
		return r.info, nil
	}

	// 如果使用CLI，从CLI获取信息
	if r.useCLI && r.cliAdapter != nil {
		info, err := r.cliAdapter.GetFileInfo(r.filePath)
		if err != nil {
			return nil, &PDFError{
				Type:    ErrorProcessing,
				Message: "无法获取PDF信息",
				File:    r.filePath,
				Cause:   err,
			}
		}
		r.info = info
		return r.info, nil
	}

	// 回退到基本信息提取
	fileInfo, err := os.Stat(r.filePath)
	if err != nil {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "无法获取文件信息",
			File:    r.filePath,
			Cause:   err,
		}
	}

	r.info = &PDFInfo{
		FilePath:      r.filePath,
		PageCount:     1,     // 默认值，实际需要解析PDF
		IsEncrypted:   false, // 默认值，实际需要检查
		FileSize:      fileInfo.Size(),
		Title:         r.extractTitle(),
		Version:       "1.4", // 默认PDF版本
		Author:        "",
		Subject:       "",
		Creator:       "",
		Producer:      "",
		CreationDate:  fileInfo.ModTime(),
		ModDate:       fileInfo.ModTime(),
		PDFCPUVersion: "",
		Permissions:   []string{},
	}

	return r.info, nil
}

// GetPageCount 获取PDF页数
func (r *PDFReader) GetPageCount() (int, error) {
	if !r.isOpen {
		return 0, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	info, err := r.GetInfo()
	if err != nil {
		return 0, err
	}

	return info.PageCount, nil
}

// ValidatePage 验证指定页面是否存在
func (r *PDFReader) ValidatePage(pageNum int) error {
	if !r.isOpen {
		return &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	info, err := r.GetInfo()
	if err != nil {
		return err
	}

	if pageNum < 1 || pageNum > info.PageCount {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: fmt.Sprintf("页码超出范围: %d (总页数: %d)", pageNum, info.PageCount),
			File:    r.filePath,
		}
	}

	return nil
}

// ValidateStructure 验证PDF文件结构完整性
func (r *PDFReader) ValidateStructure() error {
	if !r.isOpen {
		return &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	// 如果使用CLI，使用CLI验证
	if r.useCLI && r.cliAdapter != nil {
		if err := r.cliAdapter.ValidateFile(r.filePath); err != nil {
			return &PDFError{
				Type:    ErrorCorrupted,
				Message: "PDF结构验证失败",
				File:    r.filePath,
				Cause:   err,
			}
		}
		return nil
	}

	// 回退到基本验证
	info, err := r.GetInfo()
	if err != nil {
		return err
	}

	if info.PageCount <= 0 {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "PDF文件没有有效页面",
			File:    r.filePath,
		}
	}

	return nil
}

// IsEncrypted 检查PDF是否加密
func (r *PDFReader) IsEncrypted() (bool, error) {
	if !r.isOpen {
		return false, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	info, err := r.GetInfo()
	if err != nil {
		return false, err
	}

	return info.IsEncrypted, nil
}

// GetFilePath 获取文件路径
func (r *PDFReader) GetFilePath() string {
	return r.filePath
}

// IsOpen 检查读取器是否已打开
func (r *PDFReader) IsOpen() bool {
	return r.isOpen
}

// extractTitle 提取PDF标题
func (r *PDFReader) extractTitle() string {
	// 如果没有标题，使用文件名（不含扩展名）
	fileName := filepath.Base(r.filePath)
	if ext := filepath.Ext(fileName); ext != "" {
		fileName = fileName[:len(fileName)-len(ext)]
	}

	return fileName
}

// StreamPages 流式处理页面，避免一次性加载所有页面到内存
func (r *PDFReader) StreamPages(processor func(pageNum int) error) error {
	if !r.isOpen {
		return &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	info, err := r.GetInfo()
	if err != nil {
		return err
	}

	for i := 1; i <= info.PageCount; i++ {
		if err := processor(i); err != nil {
			return err
		}
	}

	return nil
}

// GetMetadata 获取PDF元数据
func (r *PDFReader) GetMetadata() (map[string]string, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	metadata := make(map[string]string)

	// 基本信息
	info, err := r.GetInfo()
	if err == nil {
		metadata["Title"] = info.Title
		metadata["PageCount"] = strconv.Itoa(info.PageCount)
		metadata["FileSize"] = strconv.FormatInt(info.FileSize, 10)
		metadata["IsEncrypted"] = strconv.FormatBool(info.IsEncrypted)
	}

	// 如果使用CLI，可以尝试获取更多信息
	if r.useCLI && r.cliAdapter != nil {
		// CLI适配器目前不支持详细元数据提取
		// 这里可以在未来扩展
	}

	return metadata, nil
}

// CheckPermissions 检查PDF权限设置
func (r *PDFReader) CheckPermissions() ([]string, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	// 如果使用CLI适配器，获取详细权限信息
	if r.useCLI && r.cliAdapter != nil {
		permissions, err := r.cliAdapter.GetPermissions(r.filePath)
		if err != nil {
			return nil, &PDFError{
				Type:    ErrorProcessing,
				Message: "无法获取PDF权限信息",
				File:    r.filePath,
				Cause:   err,
			}
		}

		if perms, ok := permissions["permissions"].([]string); ok {
			return perms, nil
		}
	}

	// 回退到基本权限检查
	isEncrypted, err := r.IsEncrypted()
	if err != nil {
		return nil, err
	}

	// 如果文件未加密，则具有所有权限
	if !isEncrypted {
		return []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"}, nil
	}

	// 对于加密文件，返回受限权限
	return []string{"print", "copy"}, nil
}

// checkPermissions 内部方法，检查特定权限
func (r *PDFReader) checkPermissions(permission string) (bool, error) {
	permissions, err := r.CheckPermissions()
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permission {
			return true, nil
		}
	}

	return false, nil
}

// CanPrint 检查是否允许打印
func (r *PDFReader) CanPrint() (bool, error) {
	return r.checkPermissions("print")
}

// CanModify 检查是否允许修改
func (r *PDFReader) CanModify() (bool, error) {
	return r.checkPermissions("modify")
}

// CanCopy 检查是否允许复制
func (r *PDFReader) CanCopy() (bool, error) {
	return r.checkPermissions("copy")
}

// CanAnnotate 检查是否允许注释
func (r *PDFReader) CanAnnotate() (bool, error) {
	return r.checkPermissions("annotate")
}

// CanFillForms 检查是否允许填写表单
func (r *PDFReader) CanFillForms() (bool, error) {
	return r.checkPermissions("fill")
}

// CanExtract 检查是否允许提取内容
func (r *PDFReader) CanExtract() (bool, error) {
	return r.checkPermissions("extract")
}

// CanAssemble 检查是否允许组装文档
func (r *PDFReader) CanAssemble() (bool, error) {
	return r.checkPermissions("assemble")
}

// CanPrintHighQuality 检查是否允许高质量打印
func (r *PDFReader) CanPrintHighQuality() (bool, error) {
	return r.checkPermissions("print_high")
}

// GetSecurityInfo 获取PDF安全设置信息
func (r *PDFReader) GetSecurityInfo() (map[string]interface{}, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	// 如果使用CLI适配器，获取详细安全信息
	if r.useCLI && r.cliAdapter != nil {
		securityInfo, err := r.cliAdapter.GetSecurityDetails(r.filePath)
		if err != nil {
			return nil, &PDFError{
				Type:    ErrorProcessing,
				Message: "无法获取PDF安全信息",
				File:    r.filePath,
				Cause:   err,
			}
		}
		return securityInfo, nil
	}

	// 回退到基本安全信息
	securityInfo := make(map[string]interface{})

	// 检查加密状态
	isEncrypted, err := r.IsEncrypted()
	if err != nil {
		return nil, err
	}

	securityInfo["encrypted"] = isEncrypted

	if isEncrypted {
		// 对于加密文件，提供基本信息
		securityInfo["version"] = "unknown"
		securityInfo["revision"] = "unknown"
		securityInfo["key_length"] = "unknown"
		securityInfo["security_handler"] = "unknown"
		securityInfo["filter"] = "unknown"

		// 权限信息
		permissions, err := r.CheckPermissions()
		if err == nil {
			securityInfo["permissions"] = permissions
		}

		securityInfo["has_user_password"] = true
		securityInfo["has_owner_password"] = true
	} else {
		securityInfo["permissions"] = []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"}
		securityInfo["has_user_password"] = false
		securityInfo["has_owner_password"] = false
		securityInfo["version"] = 0
		securityInfo["revision"] = 0
		securityInfo["key_length"] = 0
	}

	return securityInfo, nil
}

// GetDetailedSecurityInfo 获取详细的安全信息，包括加密级别分析
func (r *PDFReader) GetDetailedSecurityInfo() (map[string]interface{}, error) {
	if !r.isOpen {
		return nil, &PDFError{
			Type:    ErrorIO,
			Message: "PDF读取器未打开",
			File:    r.filePath,
		}
	}

	securityInfo, err := r.GetSecurityInfo()
	if err != nil {
		return nil, err
	}

	// 添加安全级别分析
	securityLevel := r.analyzeSecurityLevel(securityInfo)
	securityInfo["security_level"] = securityLevel

	// 添加权限摘要
	permissionSummary := r.generatePermissionSummary(securityInfo)
	securityInfo["permission_summary"] = permissionSummary

	// 添加安全建议
	securityRecommendations := r.generateSecurityRecommendations(securityInfo)
	securityInfo["security_recommendations"] = securityRecommendations

	return securityInfo, nil
}

// analyzeSecurityLevel 分析安全级别
func (r *PDFReader) analyzeSecurityLevel(securityInfo map[string]interface{}) string {
	encrypted, _ := securityInfo["encrypted"].(bool)

	if !encrypted {
		return "无保护"
	}

	keyLength, _ := securityInfo["key_length"].(int)
	version, _ := securityInfo["version"].(int)

	switch {
	case keyLength >= 256:
		return "高级加密"
	case keyLength >= 128:
		return "标准加密"
	case version >= 4:
		return "中级加密"
	default:
		return "基础加密"
	}
}

// generatePermissionSummary 生成权限摘要
func (r *PDFReader) generatePermissionSummary(securityInfo map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{})

	permissions, ok := securityInfo["permissions"].([]string)
	if !ok {
		return summary
	}

	summary["total_permissions"] = len(permissions)
	summary["can_print"] = r.containsPermission(permissions, "print")
	summary["can_modify"] = r.containsPermission(permissions, "modify")
	summary["can_copy"] = r.containsPermission(permissions, "copy")
	summary["can_annotate"] = r.containsPermission(permissions, "annotate")
	summary["can_fill_forms"] = r.containsPermission(permissions, "fill")
	summary["can_extract"] = r.containsPermission(permissions, "extract")
	summary["can_assemble"] = r.containsPermission(permissions, "assemble")
	summary["can_print_high_quality"] = r.containsPermission(permissions, "print_high")

	// 计算权限限制程度
	totalPossible := 8
	allowed := 0
	for _, perm := range []string{"print", "modify", "copy", "annotate", "fill", "extract", "assemble", "print_high"} {
		if r.containsPermission(permissions, perm) {
			allowed++
		}
	}

	summary["restriction_level"] = float64(totalPossible-allowed) / float64(totalPossible) * 100

	return summary
}

// generateSecurityRecommendations 生成安全建议
func (r *PDFReader) generateSecurityRecommendations(securityInfo map[string]interface{}) []string {
	var recommendations []string

	encrypted, _ := securityInfo["encrypted"].(bool)
	keyLength, _ := securityInfo["key_length"].(int)

	if !encrypted {
		recommendations = append(recommendations, "建议对敏感文档启用加密保护")
	} else {
		if keyLength < 128 {
			recommendations = append(recommendations, "建议使用更强的加密算法（至少128位）")
		}

		hasUserPwd, _ := securityInfo["has_user_password"].(bool)
		hasOwnerPwd, _ := securityInfo["has_owner_password"].(bool)

		if !hasUserPwd {
			recommendations = append(recommendations, "建议设置用户密码以限制文档访问")
		}

		if !hasOwnerPwd {
			recommendations = append(recommendations, "建议设置所有者密码以保护文档权限设置")
		}
	}

	permissions, ok := securityInfo["permissions"].([]string)
	if ok && len(permissions) == 8 {
		recommendations = append(recommendations, "文档具有完全权限，考虑根据需要限制某些操作")
	}

	return recommendations
}

// containsPermission 检查权限列表是否包含特定权限
func (r *PDFReader) containsPermission(permissions []string, permission string) bool {
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// OpenWithPassword 使用密码打开加密的PDF文件
func (r *PDFReader) OpenWithPassword(password string) error {
	if r.isOpen {
		r.Close()
	}

	// 验证文件是否存在
	if _, err := os.Stat(r.filePath); err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法访问PDF文件",
			File:    r.filePath,
			Cause:   err,
		}
	}

	// 如果使用CLI，尝试解密文件
	if r.useCLI && r.cliAdapter != nil {
		// 创建临时解密文件
		tempFile := filepath.Join(r.cliAdapter.GetTempDir(), "decrypted_"+filepath.Base(r.filePath))

		if err := r.cliAdapter.DecryptFile(r.filePath, tempFile, password); err != nil {
			return &PDFError{
				Type:    ErrorEncrypted,
				Message: "密码错误或解密失败",
				File:    r.filePath,
				Cause:   err,
			}
		}

		// 验证解密后的文件
		if err := r.cliAdapter.ValidateFile(tempFile); err != nil {
			return &PDFError{
				Type:    ErrorCorrupted,
				Message: "解密后的文件损坏",
				File:    r.filePath,
				Cause:   err,
			}
		}
	}

	r.isOpen = true
	return nil
}

// basicPDFValidation 基本PDF文件验证
func (r *PDFReader) basicPDFValidation() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return &PDFError{
			Type:    ErrorIO,
			Message: "无法打开PDF文件",
			File:    r.filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 检查PDF头部
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return &PDFError{
			Type:    ErrorCorrupted,
			Message: "无法读取文件头部",
			File:    r.filePath,
			Cause:   err,
		}
	}

	if !strings.HasPrefix(string(header), "%PDF-") {
		return &PDFError{
			Type:    ErrorInvalidFile,
			Message: "不是有效的PDF文件",
			File:    r.filePath,
		}
	}

	return nil
}

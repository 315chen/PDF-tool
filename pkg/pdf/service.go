package pdf

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// PDFInfo 定义PDF文件信息（保持向后兼容）
type PDFInfo struct {
	// 基本信息（保持兼容）
	PageCount   int
	IsEncrypted bool
	FileSize    int64
	Title       string
	
	// 扩展信息
	FilePath     string
	Version      string
	Author       string
	Subject      string
	Creator      string
	Producer     string
	CreationDate time.Time
	ModDate      time.Time
	
	// pdfcpu特有信息
	PDFCPUVersion string
	Permissions   []string
	
	// 额外的pdfcpu特有字段
	Keywords      string
	Trapped       string
	EncryptionMethod string
	KeyLength     int
	UserPassword  bool
	OwnerPassword bool
	PrintAllowed  bool
	ModifyAllowed bool
	CopyAllowed   bool
	AnnotateAllowed bool
	FillFormsAllowed bool
	ExtractAllowed bool
	AssembleAllowed bool
	PrintHighQualityAllowed bool
}

// PDFService 定义PDF处理服务接口
type PDFService interface {
	// ValidatePDF 验证PDF文件格式是否有效
	ValidatePDF(filePath string) error
	
	// GetPDFInfo 获取PDF文件的基本信息
	GetPDFInfo(filePath string) (*PDFInfo, error)
	
	// GetPDFMetadata 获取PDF文件元数据
	GetPDFMetadata(filePath string) (map[string]string, error)
	
	// IsPDFEncrypted 检查PDF文件是否加密
	IsPDFEncrypted(filePath string) (bool, error)
	
	// MergePDFs 将多个PDF文件合并为一个
	MergePDFs(mainFile string, additionalFiles []string, outputPath string, progressWriter io.Writer) error
}

// mapPDFInfo 将基本PDF信息映射到扩展的PDFInfo结构
func mapPDFInfo(filePath string, basicInfo map[string]interface{}) *PDFInfo {
	info := &PDFInfo{
		FilePath: filePath,
	}
	
	// 映射基本字段
	if pageCount, ok := basicInfo["PageCount"].(int); ok {
		info.PageCount = pageCount
	}
	
	if isEncrypted, ok := basicInfo["IsEncrypted"].(bool); ok {
		info.IsEncrypted = isEncrypted
	}
	
	if fileSize, ok := basicInfo["FileSize"].(int64); ok {
		info.FileSize = fileSize
	}
	
	if title, ok := basicInfo["Title"].(string); ok {
		info.Title = title
	}
	
	// 映射扩展字段
	if version, ok := basicInfo["Version"].(string); ok {
		info.Version = version
	}
	
	if author, ok := basicInfo["Author"].(string); ok {
		info.Author = author
	}
	
	if subject, ok := basicInfo["Subject"].(string); ok {
		info.Subject = subject
	}
	
	if creator, ok := basicInfo["Creator"].(string); ok {
		info.Creator = creator
	}
	
	if producer, ok := basicInfo["Producer"].(string); ok {
		info.Producer = producer
	}
	
	if creationDate, ok := basicInfo["CreationDate"].(time.Time); ok {
		info.CreationDate = creationDate
	}
	
	if modDate, ok := basicInfo["ModDate"].(time.Time); ok {
		info.ModDate = modDate
	}
	
	// 映射pdfcpu特有字段
	if pdfcpuVersion, ok := basicInfo["PDFCPUVersion"].(string); ok {
		info.PDFCPUVersion = pdfcpuVersion
	}
	
	if permissions, ok := basicInfo["Permissions"].([]string); ok {
		info.Permissions = permissions
	}
	
	// 映射额外的pdfcpu特有字段
	if keywords, ok := basicInfo["Keywords"].(string); ok {
		info.Keywords = keywords
	}
	
	if trapped, ok := basicInfo["Trapped"].(string); ok {
		info.Trapped = trapped
	}
	
	if encryptionMethod, ok := basicInfo["EncryptionMethod"].(string); ok {
		info.EncryptionMethod = encryptionMethod
	}
	
	if keyLength, ok := basicInfo["KeyLength"].(int); ok {
		info.KeyLength = keyLength
	}
	
	// 映射密码和权限标志
	if userPassword, ok := basicInfo["UserPassword"].(bool); ok {
		info.UserPassword = userPassword
	}
	
	if ownerPassword, ok := basicInfo["OwnerPassword"].(bool); ok {
		info.OwnerPassword = ownerPassword
	}
	
	if printAllowed, ok := basicInfo["PrintAllowed"].(bool); ok {
		info.PrintAllowed = printAllowed
	}
	
	if modifyAllowed, ok := basicInfo["ModifyAllowed"].(bool); ok {
		info.ModifyAllowed = modifyAllowed
	}
	
	if copyAllowed, ok := basicInfo["CopyAllowed"].(bool); ok {
		info.CopyAllowed = copyAllowed
	}
	
	if annotateAllowed, ok := basicInfo["AnnotateAllowed"].(bool); ok {
		info.AnnotateAllowed = annotateAllowed
	}
	
	if fillFormsAllowed, ok := basicInfo["FillFormsAllowed"].(bool); ok {
		info.FillFormsAllowed = fillFormsAllowed
	}
	
	if extractAllowed, ok := basicInfo["ExtractAllowed"].(bool); ok {
		info.ExtractAllowed = extractAllowed
	}
	
	if assembleAllowed, ok := basicInfo["AssembleAllowed"].(bool); ok {
		info.AssembleAllowed = assembleAllowed
	}
	
	if printHighQualityAllowed, ok := basicInfo["PrintHighQualityAllowed"].(bool); ok {
		info.PrintHighQualityAllowed = printHighQualityAllowed
	}
	
	return info
}

// mapPDFCPUInfo 专门用于映射pdfcpu输出的信息到PDFInfo结构
func mapPDFCPUInfo(filePath string, pdfcpuOutput string) *PDFInfo {
	info := NewPDFInfo(filePath)
	
	// 解析pdfcpu info命令的输出
	lines := strings.Split(pdfcpuOutput, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 解析基本信息
		if strings.HasPrefix(line, "Page count:") {
			if count, err := extractIntValue(line); err == nil {
				info.PageCount = count
			}
		} else if strings.HasPrefix(line, "Encrypted:") {
			info.IsEncrypted = strings.Contains(strings.ToLower(line), "true") || 
							   strings.Contains(strings.ToLower(line), "yes")
		} else if strings.HasPrefix(line, "PDF version:") {
			info.Version = extractStringValue(line)
		} else if strings.HasPrefix(line, "Title:") {
			info.Title = extractStringValue(line)
		} else if strings.HasPrefix(line, "Author:") {
			info.Author = extractStringValue(line)
		} else if strings.HasPrefix(line, "Subject:") {
			info.Subject = extractStringValue(line)
		} else if strings.HasPrefix(line, "Creator:") {
			info.Creator = extractStringValue(line)
		} else if strings.HasPrefix(line, "Producer:") {
			info.Producer = extractStringValue(line)
		} else if strings.HasPrefix(line, "Keywords:") {
			info.Keywords = extractStringValue(line)
		} else if strings.HasPrefix(line, "Trapped:") {
			info.Trapped = extractStringValue(line)
		} else if strings.HasPrefix(line, "Encryption method:") {
			info.EncryptionMethod = extractStringValue(line)
		} else if strings.HasPrefix(line, "Key length:") {
			if keyLen, err := extractIntValue(line); err == nil {
				info.KeyLength = keyLen
			}
		} else if strings.HasPrefix(line, "User password:") {
			info.UserPassword = strings.Contains(strings.ToLower(line), "true") ||
								strings.Contains(strings.ToLower(line), "yes")
		} else if strings.HasPrefix(line, "Owner password:") {
			info.OwnerPassword = strings.Contains(strings.ToLower(line), "true") ||
								 strings.Contains(strings.ToLower(line), "yes")
		} else if strings.HasPrefix(line, "Permissions:") {
			// 解析权限列表
			permStr := extractStringValue(line)
			if permStr != "" {
				info.Permissions = strings.Split(permStr, ",")
				for i, perm := range info.Permissions {
					info.Permissions[i] = strings.TrimSpace(perm)
				}
				info.updatePermissionFlags(info.Permissions)
			}
		}
	}
	
	return info
}

// extractStringValue 从形如 "Key: Value" 的行中提取值
func extractStringValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// extractIntValue 从形如 "Key: 123" 的行中提取整数值
func extractIntValue(line string) (int, error) {
	valueStr := extractStringValue(line)
	if valueStr == "" {
		return 0, fmt.Errorf("no value found")
	}
	return strconv.Atoi(valueStr)
}

// CreatePDFInfoFromMap 从映射创建PDFInfo（用于向后兼容）
func CreatePDFInfoFromMap(filePath string, infoMap map[string]interface{}) *PDFInfo {
	return mapPDFInfo(filePath, infoMap)
}

// NewPDFInfo 创建一个新的PDFInfo实例，确保向后兼容性
func NewPDFInfo(filePath string) *PDFInfo {
	return &PDFInfo{
		FilePath:      filePath,
		PageCount:     0,
		IsEncrypted:   false,
		FileSize:      0,
		Title:         "",
		Version:       "",
		Author:        "",
		Subject:       "",
		Creator:       "",
		Producer:      "",
		CreationDate:  time.Time{},
		ModDate:       time.Time{},
		PDFCPUVersion: "",
		Permissions:   []string{},
		
		// 初始化新的pdfcpu特有字段
		Keywords:         "",
		Trapped:          "",
		EncryptionMethod: "",
		KeyLength:        0,
		UserPassword:     false,
		OwnerPassword:    false,
		PrintAllowed:     true,  // 默认允许打印
		ModifyAllowed:    true,  // 默认允许修改
		CopyAllowed:      true,  // 默认允许复制
		AnnotateAllowed:  true,  // 默认允许注释
		FillFormsAllowed: true,  // 默认允许填写表单
		ExtractAllowed:   true,  // 默认允许提取
		AssembleAllowed:  true,  // 默认允许组装
		PrintHighQualityAllowed: true, // 默认允许高质量打印
	}
}

// IsValid 检查PDFInfo是否包含有效信息
func (info *PDFInfo) IsValid() bool {
	return info.PageCount > 0 && info.FileSize > 0
}

// HasMetadata 检查是否包含元数据信息
func (info *PDFInfo) HasMetadata() bool {
	return info.Title != "" || info.Author != "" || info.Subject != "" || 
		   info.Creator != "" || info.Producer != ""
}

// GetFormattedSize 获取格式化的文件大小字符串
func (info *PDFInfo) GetFormattedSize() string {
	const unit = 1024
	if info.FileSize < unit {
		return fmt.Sprintf("%d B", info.FileSize)
	}
	div, exp := int64(unit), 0
	for n := info.FileSize / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(info.FileSize)/float64(div), "KMGTPE"[exp])
}

// GetPermissionSummary 获取权限摘要
func (info *PDFInfo) GetPermissionSummary() string {
	if len(info.Permissions) == 0 {
		if info.IsEncrypted {
			return "受限权限"
		}
		return "完全权限"
	}
	
	return fmt.Sprintf("%d项权限", len(info.Permissions))
}

// Clone 创建PDFInfo的副本
func (info *PDFInfo) Clone() *PDFInfo {
	clone := *info
	// 深拷贝权限切片
	if info.Permissions != nil {
		clone.Permissions = make([]string, len(info.Permissions))
		copy(clone.Permissions, info.Permissions)
	}
	return &clone
}

// GetEncryptionInfo 获取加密信息摘要
func (info *PDFInfo) GetEncryptionInfo() map[string]interface{} {
	encInfo := make(map[string]interface{})
	
	encInfo["encrypted"] = info.IsEncrypted
	encInfo["method"] = info.EncryptionMethod
	encInfo["key_length"] = info.KeyLength
	encInfo["user_password"] = info.UserPassword
	encInfo["owner_password"] = info.OwnerPassword
	
	return encInfo
}

// GetPermissionFlags 获取权限标志摘要
func (info *PDFInfo) GetPermissionFlags() map[string]bool {
	return map[string]bool{
		"print":              info.PrintAllowed,
		"modify":             info.ModifyAllowed,
		"copy":               info.CopyAllowed,
		"annotate":           info.AnnotateAllowed,
		"fill_forms":         info.FillFormsAllowed,
		"extract":            info.ExtractAllowed,
		"assemble":           info.AssembleAllowed,
		"print_high_quality": info.PrintHighQualityAllowed,
	}
}

// HasRestrictedPermissions 检查是否有权限限制
func (info *PDFInfo) HasRestrictedPermissions() bool {
	return !info.PrintAllowed || !info.ModifyAllowed || !info.CopyAllowed ||
		   !info.AnnotateAllowed || !info.FillFormsAllowed || !info.ExtractAllowed ||
		   !info.AssembleAllowed || !info.PrintHighQualityAllowed
}

// GetMetadataMap 获取所有元数据的映射
func (info *PDFInfo) GetMetadataMap() map[string]string {
	metadata := make(map[string]string)
	
	if info.Title != "" {
		metadata["Title"] = info.Title
	}
	if info.Author != "" {
		metadata["Author"] = info.Author
	}
	if info.Subject != "" {
		metadata["Subject"] = info.Subject
	}
	if info.Creator != "" {
		metadata["Creator"] = info.Creator
	}
	if info.Producer != "" {
		metadata["Producer"] = info.Producer
	}
	if info.Keywords != "" {
		metadata["Keywords"] = info.Keywords
	}
	if info.Trapped != "" {
		metadata["Trapped"] = info.Trapped
	}
	
	return metadata
}

// UpdateFromPDFCPU 从pdfcpu特定信息更新PDFInfo
func (info *PDFInfo) UpdateFromPDFCPU(pdfcpuInfo map[string]interface{}) {
	// 更新pdfcpu版本
	if version, ok := pdfcpuInfo["pdfcpu_version"].(string); ok {
		info.PDFCPUVersion = version
	}
	
	// 更新权限信息
	if permissions, ok := pdfcpuInfo["permissions"].([]string); ok {
		info.Permissions = make([]string, len(permissions))
		copy(info.Permissions, permissions)
		
		// 根据权限字符串更新权限标志
		info.updatePermissionFlags(permissions)
	}
	
	// 更新加密信息
	if encMethod, ok := pdfcpuInfo["encryption_method"].(string); ok {
		info.EncryptionMethod = encMethod
	}
	
	if keyLen, ok := pdfcpuInfo["key_length"].(int); ok {
		info.KeyLength = keyLen
	}
	
	// 更新密码状态
	if userPwd, ok := pdfcpuInfo["user_password"].(bool); ok {
		info.UserPassword = userPwd
	}
	
	if ownerPwd, ok := pdfcpuInfo["owner_password"].(bool); ok {
		info.OwnerPassword = ownerPwd
	}
}

// updatePermissionFlags 根据权限字符串更新权限标志
func (info *PDFInfo) updatePermissionFlags(permissions []string) {
	// 重置所有权限为false
	info.PrintAllowed = false
	info.ModifyAllowed = false
	info.CopyAllowed = false
	info.AnnotateAllowed = false
	info.FillFormsAllowed = false
	info.ExtractAllowed = false
	info.AssembleAllowed = false
	info.PrintHighQualityAllowed = false
	
	// 根据权限列表设置标志
	for _, perm := range permissions {
		switch perm {
		case "print":
			info.PrintAllowed = true
		case "modify":
			info.ModifyAllowed = true
		case "copy":
			info.CopyAllowed = true
		case "annotate":
			info.AnnotateAllowed = true
		case "fill":
			info.FillFormsAllowed = true
		case "extract":
			info.ExtractAllowed = true
		case "assemble":
			info.AssembleAllowed = true
		case "print_high":
			info.PrintHighQualityAllowed = true
		}
	}
}
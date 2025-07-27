package encryption

// EncryptionHandler 定义加密PDF处理的核心功能接口
type EncryptionHandler interface {
	// TryAutoDecrypt 尝试使用常见密码自动解密PDF文件
	TryAutoDecrypt(filePath string) (string, error)
	
	// DecryptWithPassword 使用指定密码解密PDF文件
	DecryptWithPassword(filePath, password string) (string, error)
	
	// GetCommonPasswords 获取常用密码列表
	GetCommonPasswords() []string
	
	// RememberPassword 记住特定文件的密码
	RememberPassword(filePath, password string)
	
	// GetRememberedPassword 获取之前记住的密码
	GetRememberedPassword(filePath string) (string, bool)
}
package pdf

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// PasswordManager 密码管理器
type PasswordManager struct {
	cache           map[string]string // 文件路径 -> 密码的缓存
	commonPasswords []string          // 常用密码字典
	passwordStats   map[string]int    // 密码使用统计
	mutex           sync.RWMutex
	cacheFile       string // 缓存文件路径
	statsFile       string // 统计文件路径
	enableCache     bool   // 是否启用缓存
	enableStats     bool   // 是否启用统计
}

// PasswordManagerOptions 密码管理器选项
type PasswordManagerOptions struct {
	CacheDirectory  string   // 缓存目录
	CommonPasswords []string // 常用密码列表
	EnableCache     bool     // 是否启用缓存
	EnableStats     bool     // 是否启用统计
}

// PasswordStrength 密码强度
type PasswordStrength struct {
	Score       int      // 强度分数 (0-100)
	Level       string   // 强度等级 (weak/medium/strong)
	Suggestions []string // 改进建议
}

// PasswordStats 密码统计信息
type PasswordStats struct {
	TotalAttempts     int                  // 总尝试次数
	SuccessCount      int                  // 成功次数
	SuccessRate       float64              // 成功率
	MostUsedPasswords map[string]int       // 最常用密码
	FileStats         map[string]FileStats // 文件统计
}

// FileStats 文件统计信息
type FileStats struct {
	Attempts     int       // 尝试次数
	Success      bool      // 是否成功
	UsedPassword string    // 使用的密码
	LastAttempt  time.Time // 最后尝试时间
}

// NewPasswordManager 创建新的密码管理器
func NewPasswordManager(options *PasswordManagerOptions) *PasswordManager {
	if options == nil {
		options = &PasswordManagerOptions{
			CacheDirectory:  os.TempDir(),
			CommonPasswords: getDefaultCommonPasswords(),
			EnableCache:     true,
			EnableStats:     true,
		}
	}

	pm := &PasswordManager{
		cache:           make(map[string]string),
		commonPasswords: make([]string, len(options.CommonPasswords)),
		passwordStats:   make(map[string]int),
		cacheFile:       filepath.Join(options.CacheDirectory, "pdf_password_cache.json"),
		statsFile:       filepath.Join(options.CacheDirectory, "pdf_password_stats.json"),
		enableCache:     options.EnableCache,
		enableStats:     options.EnableStats,
	}

	// 复制常用密码列表
	copy(pm.commonPasswords, options.CommonPasswords)

	// 加载缓存和统计
	if pm.enableCache {
		pm.loadCache()
	}
	if pm.enableStats {
		pm.loadStats()
	}

	return pm
}

// GetPassword 获取文件密码（从缓存）
func (pm *PasswordManager) GetPassword(filePath string) (string, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 计算文件哈希作为缓存键
	fileHash := pm.getFileHash(filePath)
	password, exists := pm.cache[fileHash]
	return password, exists
}

// SetPassword 设置文件密码（到缓存）
func (pm *PasswordManager) SetPassword(filePath, password string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	fileHash := pm.getFileHash(filePath)
	pm.cache[fileHash] = password

	// 更新统计
	if pm.enableStats {
		pm.passwordStats[password]++
	}

	// 保存缓存
	if pm.enableCache {
		pm.saveCache()
	}
	if pm.enableStats {
		pm.saveStats()
	}
}

// RemovePassword 移除文件密码缓存
func (pm *PasswordManager) RemovePassword(filePath string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	fileHash := pm.getFileHash(filePath)
	delete(pm.cache, fileHash)

	if pm.enableCache {
		pm.saveCache()
	}
}

// ClearCache 清空所有缓存
func (pm *PasswordManager) ClearCache() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.cache = make(map[string]string)
	pm.passwordStats = make(map[string]int)

	if pm.enableCache {
		pm.saveCache()
	}
	if pm.enableStats {
		pm.saveStats()
	}
}

// GetCommonPasswords 获取常用密码列表
func (pm *PasswordManager) GetCommonPasswords() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make([]string, len(pm.commonPasswords))
	copy(result, pm.commonPasswords)
	return result
}

// SetCommonPasswords 设置常用密码列表
func (pm *PasswordManager) SetCommonPasswords(passwords []string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.commonPasswords = make([]string, len(passwords))
	copy(pm.commonPasswords, passwords)
}

// AddCommonPassword 添加常用密码
func (pm *PasswordManager) AddCommonPassword(password string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查是否已存在
	for _, existing := range pm.commonPasswords {
		if existing == password {
			return
		}
	}

	pm.commonPasswords = append(pm.commonPasswords, password)
}

// RemoveCommonPassword 移除常用密码
func (pm *PasswordManager) RemoveCommonPassword(password string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for i, existing := range pm.commonPasswords {
		if existing == password {
			pm.commonPasswords = append(pm.commonPasswords[:i], pm.commonPasswords[i+1:]...)
			return
		}
	}
}

// GetOptimizedPasswordList 获取优化的密码列表（基于统计）
func (pm *PasswordManager) GetOptimizedPasswordList() []string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	// 创建密码统计的副本
	stats := make(map[string]int)
	for k, v := range pm.passwordStats {
		stats[k] = v
	}

	// 按使用频率排序
	type passwordCount struct {
		password string
		count    int
	}

	var sortedPasswords []passwordCount
	for password, count := range stats {
		sortedPasswords = append(sortedPasswords, passwordCount{password, count})
	}

	sort.Slice(sortedPasswords, func(i, j int) bool {
		return sortedPasswords[i].count > sortedPasswords[j].count
	})

	// 构建优化列表：高频密码在前，常用密码在后
	result := make([]string, 0, len(sortedPasswords)+len(pm.commonPasswords))

	// 添加高频密码
	for _, pc := range sortedPasswords {
		result = append(result, pc.password)
	}

	// 添加常用密码（去重）
	usedPasswords := make(map[string]bool)
	for _, password := range result {
		usedPasswords[password] = true
	}

	for _, password := range pm.commonPasswords {
		if !usedPasswords[password] {
			result = append(result, password)
		}
	}

	return result
}

// ValidatePasswordStrength 验证密码强度
func (pm *PasswordManager) ValidatePasswordStrength(password string) *PasswordStrength {
	score := 0
	var suggestions []string

	// 长度检查
	if len(password) < 6 {
		suggestions = append(suggestions, "密码长度应至少6位")
	} else if len(password) >= 8 {
		score += 20
	} else {
		score += 10
	}

	// 字符类型检查
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if hasLower {
		score += 15
	} else {
		suggestions = append(suggestions, "应包含小写字母")
	}

	if hasUpper {
		score += 15
	} else {
		suggestions = append(suggestions, "应包含大写字母")
	}

	if hasDigit {
		score += 15
	} else {
		suggestions = append(suggestions, "应包含数字")
	}

	if hasSpecial {
		score += 20
	} else {
		suggestions = append(suggestions, "应包含特殊字符")
	}

	// 重复字符检查
	if hasRepeatingChars(password) {
		score -= 10
		suggestions = append(suggestions, "避免重复字符")
	}

	// 常见密码检查
	if isCommonPassword(password) {
		score -= 20
		suggestions = append(suggestions, "避免使用常见密码")
	}

	// 长密码奖励（超过12位）
	if len(password) >= 12 {
		score += 15
	}
	// 超长密码奖励（超过50位）
	if len(password) >= 50 {
		score += 25
	}

	// 确定强度等级
	var level string
	switch {
	case score >= 80:
		level = "strong"
	case score >= 45:
		level = "medium"
	default:
		level = "weak"
	}

	return &PasswordStrength{
		Score:       score,
		Level:       level,
		Suggestions: suggestions,
	}
}

// GetPasswordStats 获取密码统计信息
func (pm *PasswordManager) GetPasswordStats() *PasswordStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := &PasswordStats{
		MostUsedPasswords: make(map[string]int),
		FileStats:         make(map[string]FileStats),
	}

	// 复制统计信息
	for k, v := range pm.passwordStats {
		stats.MostUsedPasswords[k] = v
		stats.TotalAttempts += v
	}

	// 计算成功率（这里简化处理，实际应该从解密结果统计）
	if stats.TotalAttempts > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalAttempts) * 100
	}

	return stats
}

// BatchTryPasswords 批量尝试密码
func (pm *PasswordManager) BatchTryPasswords(filePath string, passwords []string,
	decryptFunc func(string, string) (string, error)) (string, string, error) {

	// 首先检查缓存
	if cachedPassword, exists := pm.GetPassword(filePath); exists {
		if _, err := decryptFunc(filePath, cachedPassword); err == nil {
			return filePath, cachedPassword, nil
		}
		// 缓存密码无效，移除缓存
		pm.RemovePassword(filePath)
	}

	// 尝试密码列表
	for _, password := range passwords {
		if decryptedPath, err := decryptFunc(filePath, password); err == nil {
			// 成功解密，缓存密码
			pm.SetPassword(filePath, password)
			return decryptedPath, password, nil
		}
	}

	return "", "", fmt.Errorf("所有密码尝试失败")
}

// getFileHash 计算文件哈希
func (pm *PasswordManager) getFileHash(filePath string) string {
	// 使用文件路径计算哈希，避免文件内容变化影响缓存
	hash := md5.Sum([]byte(filePath))
	return hex.EncodeToString(hash[:])
}

// loadCache 加载缓存
func (pm *PasswordManager) loadCache() {
	// 这里简化实现，实际应该从JSON文件加载
	// 为了演示，使用空实现
}

// saveCache 保存缓存
func (pm *PasswordManager) saveCache() {
	// 这里简化实现，实际应该保存到JSON文件
	// 为了演示，使用空实现
}

// loadStats 加载统计
func (pm *PasswordManager) loadStats() {
	// 这里简化实现，实际应该从JSON文件加载
	// 为了演示，使用空实现
}

// saveStats 保存统计
func (pm *PasswordManager) saveStats() {
	// 这里简化实现，实际应该保存到JSON文件
	// 为了演示，使用空实现
}

// hasRepeatingChars 检查是否有重复字符
func hasRepeatingChars(password string) bool {
	if len(password) < 3 {
		return false
	}

	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i] == password[i+2] {
			return true
		}
	}
	return false
}

// isCommonPassword 检查是否为常见密码
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"qwerty", "abc123", "111111", "123123", "admin",
		"letmein", "welcome", "monkey", "1234", "dragon",
		"pass", "master", "hello", "freedom", "whatever",
	}

	password = strings.ToLower(password)
	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}
	return false
}

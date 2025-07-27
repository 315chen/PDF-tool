package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigChangeCallback 配置变更回调函数类型
type ConfigChangeCallback func(oldConfig, newConfig *Config)

// ConfigManager 定义配置管理器
type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
	callbacks  []ConfigChangeCallback
	watching   bool
	stopWatch  chan bool
}

// NewConfigManager 创建一个新的配置管理器
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		config:     DefaultConfig(),
		configPath: configPath,
		callbacks:  make([]ConfigChangeCallback, 0),
		watching:   false,
		stopWatch:  make(chan bool, 1),
	}
}

// LoadConfig 从文件加载配置
func (cm *ConfigManager) LoadConfig() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		return nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// 合并默认配置和加载的配置
	cm.mergeWithDefaults(&config)
	cm.config = &config

	return nil
}

// SaveConfig 保存配置到文件
func (cm *ConfigManager) SaveConfig() error {
	// 确保配置目录存在
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(config *Config) {
	cm.config = config
}

// SetMaxMemoryUsage 设置最大内存使用量
func (cm *ConfigManager) SetMaxMemoryUsage(size int64) {
	cm.config.MaxMemoryUsage = size
}

// SetTempDirectory 设置临时目录
func (cm *ConfigManager) SetTempDirectory(dir string) {
	cm.config.TempDirectory = dir
}

// SetOutputDirectory 设置默认输出目录
func (cm *ConfigManager) SetOutputDirectory(dir string) {
	cm.config.OutputDirectory = dir
}

// SetAutoDecrypt 设置是否启用自动解密
func (cm *ConfigManager) SetAutoDecrypt(enabled bool) {
	cm.config.EnableAutoDecrypt = enabled
}

// AddCommonPassword 添加常用密码
func (cm *ConfigManager) AddCommonPassword(password string) {
	// 检查密码是否已存在
	for _, existing := range cm.config.CommonPasswords {
		if existing == password {
			return
		}
	}
	
	cm.config.CommonPasswords = append(cm.config.CommonPasswords, password)
}

// RemoveCommonPassword 移除常用密码
func (cm *ConfigManager) RemoveCommonPassword(password string) {
	for i, existing := range cm.config.CommonPasswords {
		if existing == password {
			cm.config.CommonPasswords = append(
				cm.config.CommonPasswords[:i],
				cm.config.CommonPasswords[i+1:]...,
			)
			return
		}
	}
}

// SetWindowSize 设置窗口大小
func (cm *ConfigManager) SetWindowSize(width, height int) {
	cm.config.WindowWidth = width
	cm.config.WindowHeight = height
}

// mergeWithDefaults 将加载的配置与默认配置合并
func (cm *ConfigManager) mergeWithDefaults(config *Config) {
	defaults := DefaultConfig()

	// 如果某些字段为空或零值，使用默认值
	if config.MaxMemoryUsage <= 0 {
		config.MaxMemoryUsage = defaults.MaxMemoryUsage
	}

	if config.TempDirectory == "" {
		config.TempDirectory = defaults.TempDirectory
	}

	if config.OutputDirectory == "" {
		config.OutputDirectory = defaults.OutputDirectory
	}
	
	if len(config.CommonPasswords) == 0 {
		config.CommonPasswords = defaults.CommonPasswords
	}
	
	// 注意：我们不覆盖EnableAutoDecrypt的值，因为布尔值没有明确的"未设置"状态
	
	if config.WindowWidth <= 0 {
		config.WindowWidth = defaults.WindowWidth
	}
	
	if config.WindowHeight <= 0 {
		config.WindowHeight = defaults.WindowHeight
	}

	if len(config.CommonPasswords) == 0 {
		config.CommonPasswords = defaults.CommonPasswords
	}

	if config.WindowWidth <= 0 {
		config.WindowWidth = defaults.WindowWidth
	}

	if config.WindowHeight <= 0 {
		config.WindowHeight = defaults.WindowHeight
	}
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".pdf-merger", "config.json"), nil
}

// AddConfigChangeCallback 添加配置变更回调
func (cm *ConfigManager) AddConfigChangeCallback(callback ConfigChangeCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.callbacks = append(cm.callbacks, callback)
}

// RemoveConfigChangeCallback 移除配置变更回调
func (cm *ConfigManager) RemoveConfigChangeCallback(callback ConfigChangeCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 注意：这里使用函数指针比较，在实际使用中可能需要使用ID或其他标识
	for i, cb := range cm.callbacks {
		if &cb == &callback {
			cm.callbacks = append(cm.callbacks[:i], cm.callbacks[i+1:]...)
			break
		}
	}
}

// notifyConfigChange 通知配置变更
func (cm *ConfigManager) notifyConfigChange(oldConfig, newConfig *Config) {
	cm.mutex.RLock()
	callbacks := make([]ConfigChangeCallback, len(cm.callbacks))
	copy(callbacks, cm.callbacks)
	cm.mutex.RUnlock()

	for _, callback := range callbacks {
		go callback(oldConfig, newConfig)
	}
}

// UpdateConfigWithNotification 更新配置并通知变更
func (cm *ConfigManager) UpdateConfigWithNotification(config *Config) {
	cm.mutex.Lock()
	oldConfig := cm.config
	cm.config = config
	cm.mutex.Unlock()

	cm.notifyConfigChange(oldConfig, config)
}

// StartWatching 开始监听配置文件变更
func (cm *ConfigManager) StartWatching() error {
	cm.mutex.Lock()
	if cm.watching {
		cm.mutex.Unlock()
		return nil // 已经在监听
	}
	cm.watching = true
	cm.mutex.Unlock()

	go cm.watchConfigFile()
	return nil
}

// StopWatching 停止监听配置文件变更
func (cm *ConfigManager) StopWatching() {
	cm.mutex.Lock()
	if !cm.watching {
		cm.mutex.Unlock()
		return
	}
	cm.watching = false
	cm.mutex.Unlock()

	select {
	case cm.stopWatch <- true:
	default:
	}
}

// watchConfigFile 监听配置文件变更的内部方法
func (cm *ConfigManager) watchConfigFile() {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
	defer ticker.Stop()

	var lastModTime time.Time
	if info, err := os.Stat(cm.configPath); err == nil {
		lastModTime = info.ModTime()
	}

	for {
		select {
		case <-cm.stopWatch:
			return
		case <-ticker.C:
			if info, err := os.Stat(cm.configPath); err == nil {
				if info.ModTime().After(lastModTime) {
					lastModTime = info.ModTime()
					cm.reloadConfig()
				}
			}
		}
	}
}

// reloadConfig 重新加载配置文件
func (cm *ConfigManager) reloadConfig() {
	cm.mutex.Lock()
	oldConfig := cm.config
	cm.mutex.Unlock()

	if err := cm.LoadConfig(); err != nil {
		// 加载失败，保持原配置
		return
	}

	cm.mutex.RLock()
	newConfig := cm.config
	cm.mutex.RUnlock()

	// 检查配置是否真的发生了变化
	if !cm.configsEqual(oldConfig, newConfig) {
		cm.notifyConfigChange(oldConfig, newConfig)
	}
}

// configsEqual 比较两个配置是否相等
func (cm *ConfigManager) configsEqual(config1, config2 *Config) bool {
	if config1 == nil || config2 == nil {
		return config1 == config2
	}

	return config1.MaxMemoryUsage == config2.MaxMemoryUsage &&
		config1.TempDirectory == config2.TempDirectory &&
		config1.OutputDirectory == config2.OutputDirectory &&
		config1.EnableAutoDecrypt == config2.EnableAutoDecrypt &&
		config1.WindowWidth == config2.WindowWidth &&
		config1.WindowHeight == config2.WindowHeight &&
		cm.slicesEqual(config1.CommonPasswords, config2.CommonPasswords)
}

// slicesEqual 比较两个字符串切片是否相等
func (cm *ConfigManager) slicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}

	return true
}

// GetConfigSafely 线程安全地获取配置副本
func (cm *ConfigManager) GetConfigSafely() *Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建配置的深拷贝
	configCopy := *cm.config
	configCopy.CommonPasswords = make([]string, len(cm.config.CommonPasswords))
	copy(configCopy.CommonPasswords, cm.config.CommonPasswords)

	return &configCopy
}

// IsWatching 检查是否正在监听配置文件
func (cm *ConfigManager) IsWatching() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.watching
}
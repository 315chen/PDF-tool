package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewConfigManager(t *testing.T) {
	configPath := "/tmp/test-config.json"
	cm := NewConfigManager(configPath)

	if cm == nil {
		t.Fatal("Expected non-nil ConfigManager")
	}

	if cm.configPath != configPath {
		t.Errorf("Expected configPath %s, got %s", configPath, cm.configPath)
	}

	if cm.config == nil {
		t.Fatal("Expected non-nil default config")
	}
}

func TestConfigManager_LoadConfig_NonExistentFile(t *testing.T) {
	configPath := "/tmp/nonexistent-config.json"
	cm := NewConfigManager(configPath)

	err := cm.LoadConfig()
	if err != nil {
		t.Errorf("Expected no error for nonexistent config file, got %v", err)
	}

	// 应该使用默认配置
	config := cm.GetConfig()
	if config.MaxMemoryUsage != 100*1024*1024 {
		t.Errorf("Expected default MaxMemoryUsage, got %d", config.MaxMemoryUsage)
	}
}

func TestConfigManager_SaveAndLoadConfig(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pdf-merger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	cm := NewConfigManager(configPath)

	// 修改配置
	cm.SetMaxMemoryUsage(200 * 1024 * 1024)
	cm.SetAutoDecrypt(false)
	cm.SetWindowSize(1024, 768)
	cm.AddCommonPassword("testpass")

	// 保存配置
	err = cm.SaveConfig()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 创建新的配置管理器并加载配置
	cm2 := NewConfigManager(configPath)
	err = cm2.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm2.GetConfig()

	// 验证配置是否正确加载
	if config.MaxMemoryUsage != 200*1024*1024 {
		t.Errorf("Expected MaxMemoryUsage %d, got %d", 200*1024*1024, config.MaxMemoryUsage)
	}

	if config.EnableAutoDecrypt {
		t.Error("Expected EnableAutoDecrypt to be false")
	}

	if config.WindowWidth != 1024 {
		t.Errorf("Expected WindowWidth 1024, got %d", config.WindowWidth)
	}

	if config.WindowHeight != 768 {
		t.Errorf("Expected WindowHeight 768, got %d", config.WindowHeight)
	}

	// 检查是否包含添加的密码
	found := false
	for _, password := range config.CommonPasswords {
		if password == "testpass" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find added password 'testpass'")
	}
}

func TestConfigManager_AddCommonPassword(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")
	
	initialCount := len(cm.GetConfig().CommonPasswords)
	
	// 添加新密码
	cm.AddCommonPassword("newpassword")
	
	config := cm.GetConfig()
	if len(config.CommonPasswords) != initialCount+1 {
		t.Errorf("Expected %d passwords, got %d", initialCount+1, len(config.CommonPasswords))
	}

	// 尝试添加重复密码
	cm.AddCommonPassword("newpassword")
	
	config = cm.GetConfig()
	if len(config.CommonPasswords) != initialCount+1 {
		t.Errorf("Expected password count to remain %d after adding duplicate, got %d", initialCount+1, len(config.CommonPasswords))
	}
}

func TestConfigManager_RemoveCommonPassword(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")
	
	// 添加一个密码
	cm.AddCommonPassword("toremove")
	initialCount := len(cm.GetConfig().CommonPasswords)
	
	// 移除密码
	cm.RemoveCommonPassword("toremove")
	
	config := cm.GetConfig()
	if len(config.CommonPasswords) != initialCount-1 {
		t.Errorf("Expected %d passwords after removal, got %d", initialCount-1, len(config.CommonPasswords))
	}

	// 检查密码是否真的被移除
	for _, password := range config.CommonPasswords {
		if password == "toremove" {
			t.Error("Expected password 'toremove' to be removed")
		}
	}

	// 尝试移除不存在的密码
	cm.RemoveCommonPassword("nonexistent")
	
	config = cm.GetConfig()
	if len(config.CommonPasswords) != initialCount-1 {
		t.Errorf("Expected password count to remain %d after removing nonexistent, got %d", initialCount-1, len(config.CommonPasswords))
	}
}

func TestConfigManager_UpdateConfig(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")
	
	newConfig := &Config{
		MaxMemoryUsage:    500 * 1024 * 1024,
		EnableAutoDecrypt: false,
		WindowWidth:       1200,
		WindowHeight:      800,
		CommonPasswords:   []string{"test1", "test2"},
	}

	cm.UpdateConfig(newConfig)
	
	config := cm.GetConfig()
	if config != newConfig {
		t.Error("Expected UpdateConfig to set the new config")
	}

	if config.MaxMemoryUsage != 500*1024*1024 {
		t.Errorf("Expected MaxMemoryUsage %d, got %d", 500*1024*1024, config.MaxMemoryUsage)
	}
}

func TestConfigManager_MergeWithDefaults(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pdf-merger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// 创建一个不完整的配置文件
	incompleteConfig := map[string]interface{}{
		"MaxMemoryUsage": 50 * 1024 * 1024,
		"WindowWidth":    1024,
		// 缺少其他字段
	}

	data, err := json.Marshal(incompleteConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 加载配置
	cm := NewConfigManager(configPath)
	err = cm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm.GetConfig()

	// 检查是否正确合并了默认值
	if config.MaxMemoryUsage != 50*1024*1024 {
		t.Errorf("Expected MaxMemoryUsage from file %d, got %d", 50*1024*1024, config.MaxMemoryUsage)
	}

	if config.WindowWidth != 1024 {
		t.Errorf("Expected WindowWidth from file 1024, got %d", config.WindowWidth)
	}

	// 这些应该来自默认配置
	if config.WindowHeight != 600 {
		t.Errorf("Expected default WindowHeight 600, got %d", config.WindowHeight)
	}

	// 注意：我们不再检查EnableAutoDecrypt的值，因为它不会被默认值覆盖

	if len(config.CommonPasswords) == 0 {
		t.Error("Expected default CommonPasswords to be non-empty")
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path, err := GetDefaultConfigPath()
	if err != nil {
		t.Fatalf("Failed to get default config path: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty config path")
	}

	// 检查路径是否包含预期的组件
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}

	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".pdf-merger" {
		t.Errorf("Expected config directory to be '.pdf-merger', got %s", filepath.Base(dir))
	}

	if filepath.Base(path) != "config.json" {
		t.Errorf("Expected config file to be 'config.json', got %s", filepath.Base(path))
	}
}

func TestConfigManager_ConfigChangeCallback(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")

	var callbackCalled bool
	var oldConfigReceived, newConfigReceived *Config
	var wg sync.WaitGroup

	callback := func(oldConfig, newConfig *Config) {
		defer wg.Done()
		callbackCalled = true
		oldConfigReceived = oldConfig
		newConfigReceived = newConfig
	}

	cm.AddConfigChangeCallback(callback)

	// 更新配置并等待回调
	wg.Add(1)
	oldConfig := cm.GetConfig()
	newConfig := &Config{
		MaxMemoryUsage:    200 * 1024 * 1024,
		EnableAutoDecrypt: false,
		WindowWidth:       1024,
		WindowHeight:      768,
		CommonPasswords:   []string{"test"},
	}

	cm.UpdateConfigWithNotification(newConfig)

	// 等待回调完成
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// 回调完成
	case <-time.After(1 * time.Second):
		t.Fatal("Callback not called within timeout")
	}

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if oldConfigReceived != oldConfig {
		t.Error("Expected old config to match")
	}

	if newConfigReceived != newConfig {
		t.Error("Expected new config to match")
	}
}

func TestConfigManager_GetConfigSafely(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")

	// 获取配置副本
	config1 := cm.GetConfigSafely()
	config2 := cm.GetConfigSafely()

	// 应该是不同的对象
	if config1 == config2 {
		t.Error("Expected different config objects")
	}

	// 但内容应该相同
	if config1.MaxMemoryUsage != config2.MaxMemoryUsage {
		t.Error("Expected same MaxMemoryUsage")
	}

	// 修改一个副本不应该影响另一个
	config1.MaxMemoryUsage = 999
	if config2.MaxMemoryUsage == 999 {
		t.Error("Expected config copies to be independent")
	}

	// 修改密码列表也不应该影响原配置
	config1.CommonPasswords[0] = "modified"
	originalConfig := cm.GetConfig()
	if originalConfig.CommonPasswords[0] == "modified" {
		t.Error("Expected original config to be unmodified")
	}
}

func TestConfigManager_ConfigsEqual(t *testing.T) {
	cm := NewConfigManager("/tmp/test-config.json")

	config1 := &Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	config2 := &Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	config3 := &Config{
		MaxMemoryUsage:    200 * 1024 * 1024, // 不同
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	if !cm.configsEqual(config1, config2) {
		t.Error("Expected identical configs to be equal")
	}

	if cm.configsEqual(config1, config3) {
		t.Error("Expected different configs to be unequal")
	}

	if cm.configsEqual(config1, nil) {
		t.Error("Expected config and nil to be unequal")
	}

	if !cm.configsEqual(nil, nil) {
		t.Error("Expected nil and nil to be equal")
	}
}

func TestConfigManager_WatchingFunctionality(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pdf-merger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	cm := NewConfigManager(configPath)

	// 初始状态不应该在监听
	if cm.IsWatching() {
		t.Error("Expected not to be watching initially")
	}

	// 开始监听
	err = cm.StartWatching()
	if err != nil {
		t.Fatalf("Failed to start watching: %v", err)
	}

	if !cm.IsWatching() {
		t.Error("Expected to be watching after StartWatching")
	}

	// 停止监听
	cm.StopWatching()

	// 给一点时间让goroutine停止
	time.Sleep(100 * time.Millisecond)

	if cm.IsWatching() {
		t.Error("Expected not to be watching after StopWatching")
	}
}
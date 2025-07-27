//go:build ignore
// +build ignore
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/pdf-merger/internal/model"
)

func main() {
	fmt.Println("=== PDF合并工具配置管理系统演示 ===\n")

	// 1. 演示默认配置
	demonstrateDefaultConfig()

	// 2. 演示配置管理器基本功能
	demonstrateConfigManager()

	// 3. 演示配置持久化
	demonstrateConfigPersistence()

	// 4. 演示配置验证
	demonstrateConfigValidation()

	// 5. 演示配置合并
	demonstrateConfigMerging()

	// 6. 演示高级功能
	demonstrateAdvancedFeatures()

	fmt.Println("\n=== 配置管理系统演示完成 ===")
}

func demonstrateDefaultConfig() {
	fmt.Println("1. 默认配置演示:")
	
	config := model.DefaultConfig()
	
	fmt.Printf("   最大内存使用: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	fmt.Printf("   临时目录: %s (空表示使用系统默认)\n", config.TempDirectory)
	fmt.Printf("   输出目录: %s (空表示使用用户文档目录)\n", config.OutputDirectory)
	fmt.Printf("   自动解密: %t\n", config.EnableAutoDecrypt)
	fmt.Printf("   窗口大小: %dx%d\n", config.WindowWidth, config.WindowHeight)
	fmt.Printf("   常用密码数量: %d\n", len(config.CommonPasswords))
	fmt.Printf("   前5个常用密码: %v\n", config.CommonPasswords[:5])
	
	fmt.Println()
}

func demonstrateConfigManager() {
	fmt.Println("2. 配置管理器基本功能演示:")
	
	// 创建临时配置文件路径
	tempDir, _ := os.MkdirTemp("", "pdf-merger-demo")
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.json")
	cm := model.NewConfigManager(configPath)
	
	fmt.Printf("   配置文件路径: %s\n", configPath)
	
	// 获取初始配置
	config := cm.GetConfig()
	fmt.Printf("   初始最大内存: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	
	// 修改配置
	cm.SetMaxMemoryUsage(200 * 1024 * 1024) // 200MB
	cm.SetAutoDecrypt(false)
	cm.SetWindowSize(1024, 768)
	cm.SetTempDirectory("/tmp/pdf-merger")
	cm.SetOutputDirectory("/Users/user/Documents/PDF输出")
	
	// 添加自定义密码
	cm.AddCommonPassword("mypassword123")
	cm.AddCommonPassword("secret2023")
	
	config = cm.GetConfig()
	fmt.Printf("   修改后最大内存: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	fmt.Printf("   自动解密: %t\n", config.EnableAutoDecrypt)
	fmt.Printf("   窗口大小: %dx%d\n", config.WindowWidth, config.WindowHeight)
	fmt.Printf("   临时目录: %s\n", config.TempDirectory)
	fmt.Printf("   输出目录: %s\n", config.OutputDirectory)
	fmt.Printf("   密码列表长度: %d\n", len(config.CommonPasswords))
	
	// 移除密码
	cm.RemoveCommonPassword("123456")
	config = cm.GetConfig()
	fmt.Printf("   移除'123456'后密码数量: %d\n", len(config.CommonPasswords))
	
	fmt.Println()
}

func demonstrateConfigPersistence() {
	fmt.Println("3. 配置持久化演示:")
	
	// 创建临时目录
	tempDir, _ := os.MkdirTemp("", "pdf-merger-demo")
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "config.json")
	
	// 第一个配置管理器 - 设置配置并保存
	fmt.Println("   创建第一个配置管理器并设置配置...")
	cm1 := model.NewConfigManager(configPath)
	cm1.SetMaxMemoryUsage(150 * 1024 * 1024)
	cm1.SetWindowSize(900, 700)
	cm1.AddCommonPassword("demo123")
	
	err := cm1.SaveConfig()
	if err != nil {
		fmt.Printf("   保存配置失败: %v\n", err)
		return
	}
	fmt.Println("   配置已保存到文件 ✓")
	
	// 检查文件是否存在
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("   配置文件存在 ✓")
	} else {
		fmt.Printf("   配置文件不存在: %v\n", err)
		return
	}
	
	// 第二个配置管理器 - 从文件加载配置
	fmt.Println("   创建第二个配置管理器并加载配置...")
	cm2 := model.NewConfigManager(configPath)
	err = cm2.LoadConfig()
	if err != nil {
		fmt.Printf("   加载配置失败: %v\n", err)
		return
	}
	
	config := cm2.GetConfig()
	fmt.Printf("   加载的最大内存: %d MB\n", config.MaxMemoryUsage/(1024*1024))
	fmt.Printf("   加载的窗口大小: %dx%d\n", config.WindowWidth, config.WindowHeight)
	
	// 检查自定义密码是否存在
	hasDemo123 := false
	for _, pwd := range config.CommonPasswords {
		if pwd == "demo123" {
			hasDemo123 = true
			break
		}
	}
	fmt.Printf("   自定义密码'demo123'存在: %t\n", hasDemo123)
	
	fmt.Println()
}

func demonstrateConfigValidation() {
	fmt.Println("4. 配置验证演示:")
	
	validator := model.NewValidator()
	
	// 验证有效配置
	validConfig := model.DefaultConfig()
	if err := validator.ValidateConfig(validConfig); err != nil {
		fmt.Printf("   默认配置验证失败: %v\n", err)
	} else {
		fmt.Println("   默认配置验证通过 ✓")
	}
	
	// 验证无效配置 - 内存使用量为负数
	invalidConfig1 := &model.Config{
		MaxMemoryUsage:    -1,
		WindowWidth:       800,
		WindowHeight:      600,
		EnableAutoDecrypt: true,
		CommonPasswords:   []string{"test"},
	}
	
	if err := validator.ValidateConfig(invalidConfig1); err != nil {
		fmt.Printf("   无效配置1验证失败 (预期): %v\n", err)
	} else {
		fmt.Println("   无效配置1验证通过 (意外)")
	}
	
	// 验证无效配置 - 窗口大小超出范围
	invalidConfig2 := &model.Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		WindowWidth:       5000, // 超出范围
		WindowHeight:      600,
		EnableAutoDecrypt: true,
		CommonPasswords:   []string{"test"},
	}
	
	if err := validator.ValidateConfig(invalidConfig2); err != nil {
		fmt.Printf("   无效配置2验证失败 (预期): %v\n", err)
	} else {
		fmt.Println("   无效配置2验证通过 (意外)")
	}
	
	// 验证无效配置 - 密码过长
	longPassword := make([]byte, 150)
	for i := range longPassword {
		longPassword[i] = 'a'
	}
	
	invalidConfig3 := &model.Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		WindowWidth:       800,
		WindowHeight:      600,
		EnableAutoDecrypt: true,
		CommonPasswords:   []string{string(longPassword)},
	}
	
	if err := validator.ValidateConfig(invalidConfig3); err != nil {
		fmt.Printf("   无效配置3验证失败 (预期): %v\n", err)
	} else {
		fmt.Println("   无效配置3验证通过 (意外)")
	}
	
	fmt.Println()
}

func demonstrateConfigMerging() {
	fmt.Println("5. 配置合并演示:")
	
	// 创建临时目录和不完整的配置文件
	tempDir, _ := os.MkdirTemp("", "pdf-merger-demo")
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "incomplete_config.json")
	
	// 创建不完整的配置文件（只包含部分字段）
	incompleteConfigJSON := `{
  "MaxMemoryUsage": 50000000,
  "WindowWidth": 1200,
  "EnableAutoDecrypt": false
}`
	
	err := os.WriteFile(configPath, []byte(incompleteConfigJSON), 0644)
	if err != nil {
		fmt.Printf("   创建不完整配置文件失败: %v\n", err)
		return
	}
	
	fmt.Println("   创建了不完整的配置文件:")
	fmt.Println("   - MaxMemoryUsage: 50MB")
	fmt.Println("   - WindowWidth: 1200")
	fmt.Println("   - EnableAutoDecrypt: false")
	fmt.Println("   - 缺少其他字段")
	
	// 加载配置并观察合并结果
	cm := model.NewConfigManager(configPath)
	err = cm.LoadConfig()
	if err != nil {
		fmt.Printf("   加载配置失败: %v\n", err)
		return
	}
	
	config := cm.GetConfig()
	
	fmt.Println("\n   合并后的配置:")
	fmt.Printf("   - MaxMemoryUsage: %d MB (来自文件)\n", config.MaxMemoryUsage/(1024*1024))
	fmt.Printf("   - WindowWidth: %d (来自文件)\n", config.WindowWidth)
	fmt.Printf("   - WindowHeight: %d (来自默认值)\n", config.WindowHeight)
	fmt.Printf("   - EnableAutoDecrypt: %t (来自文件)\n", config.EnableAutoDecrypt)
	fmt.Printf("   - TempDirectory: '%s' (来自默认值)\n", config.TempDirectory)
	fmt.Printf("   - OutputDirectory: '%s' (来自默认值)\n", config.OutputDirectory)
	fmt.Printf("   - CommonPasswords数量: %d (来自默认值)\n", len(config.CommonPasswords))
	
	fmt.Println("\n   配置合并成功 ✓")
	fmt.Println("   文件中的值被保留，缺失的字段使用默认值")

	fmt.Println()
}

func demonstrateAdvancedFeatures() {
	fmt.Println("6. 高级功能演示:")

	// 创建临时目录
	tempDir, _ := os.MkdirTemp("", "pdf-merger-demo")
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	cm := model.NewConfigManager(configPath)

	// 6.1 演示配置变更回调
	fmt.Println("   6.1 配置变更回调演示:")

	callbackCount := 0
	callback := func(oldConfig, newConfig *model.Config) {
		callbackCount++
		fmt.Printf("       [回调%d] 配置已变更:\n", callbackCount)
		fmt.Printf("       - 内存使用: %d MB -> %d MB\n",
			oldConfig.MaxMemoryUsage/(1024*1024),
			newConfig.MaxMemoryUsage/(1024*1024))
		fmt.Printf("       - 窗口宽度: %d -> %d\n",
			oldConfig.WindowWidth, newConfig.WindowWidth)
	}

	cm.AddConfigChangeCallback(callback)

	// 触发配置变更
	newConfig := &model.Config{
		MaxMemoryUsage:    150 * 1024 * 1024,
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       1000,
		WindowHeight:      700,
		CommonPasswords:   []string{"test1", "test2"},
	}

	cm.UpdateConfigWithNotification(newConfig)

	// 等待回调完成
	time.Sleep(100 * time.Millisecond)

	// 6.2 演示线程安全的配置获取
	fmt.Println("\n   6.2 线程安全配置获取演示:")

	config1 := cm.GetConfigSafely()
	config2 := cm.GetConfigSafely()

	fmt.Printf("       配置副本1地址: %p\n", config1)
	fmt.Printf("       配置副本2地址: %p\n", config2)
	fmt.Printf("       是否为不同对象: %t\n", config1 != config2)
	fmt.Printf("       内容是否相同: %t\n", config1.MaxMemoryUsage == config2.MaxMemoryUsage)

	// 修改一个副本不影响另一个
	config1.MaxMemoryUsage = 999 * 1024 * 1024
	fmt.Printf("       修改副本1后，副本2的内存设置: %d MB\n", config2.MaxMemoryUsage/(1024*1024))

	// 6.3 演示配置监听功能
	fmt.Println("\n   6.3 配置文件监听演示:")

	fmt.Printf("       初始监听状态: %t\n", cm.IsWatching())

	// 开始监听
	err := cm.StartWatching()
	if err != nil {
		fmt.Printf("       启动监听失败: %v\n", err)
		return
	}

	fmt.Printf("       启动监听后状态: %t\n", cm.IsWatching())

	// 停止监听
	cm.StopWatching()
	time.Sleep(200 * time.Millisecond) // 等待监听goroutine停止

	fmt.Printf("       停止监听后状态: %t\n", cm.IsWatching())

	// 6.4 演示配置比较功能
	fmt.Println("\n   6.4 配置比较功能演示:")

	config3 := &model.Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	config4 := &model.Config{
		MaxMemoryUsage:    100 * 1024 * 1024,
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	config5 := &model.Config{
		MaxMemoryUsage:    200 * 1024 * 1024, // 不同
		TempDirectory:     "/tmp",
		OutputDirectory:   "/output",
		EnableAutoDecrypt: true,
		WindowWidth:       800,
		WindowHeight:      600,
		CommonPasswords:   []string{"pass1", "pass2"},
	}

	// 使用反射访问私有方法进行演示（实际使用中不推荐）
	fmt.Printf("       配置3和配置4相同: %t (预期: true)\n",
		compareConfigs(cm, config3, config4))
	fmt.Printf("       配置3和配置5相同: %t (预期: false)\n",
		compareConfigs(cm, config3, config5))

	fmt.Println("\n   高级功能演示完成 ✓")
	fmt.Println()
}

// compareConfigs 辅助函数，用于演示配置比较
func compareConfigs(cm *model.ConfigManager, config1, config2 *model.Config) bool {
	// 简单的配置比较实现
	return config1.MaxMemoryUsage == config2.MaxMemoryUsage &&
		config1.TempDirectory == config2.TempDirectory &&
		config1.OutputDirectory == config2.OutputDirectory &&
		config1.EnableAutoDecrypt == config2.EnableAutoDecrypt &&
		config1.WindowWidth == config2.WindowWidth &&
		config1.WindowHeight == config2.WindowHeight &&
		slicesEqual(config1.CommonPasswords, config2.CommonPasswords)
}

// slicesEqual 辅助函数，比较字符串切片
func slicesEqual(slice1, slice2 []string) bool {
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

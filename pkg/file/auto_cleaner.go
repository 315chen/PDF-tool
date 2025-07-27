package file

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// AutoCleaner 自动资源清理器
type AutoCleaner struct {
	resourceManager *ResourceManager
	signalChan      chan os.Signal
	cleanupDone     chan struct{}
	mutex           sync.Mutex
	isSetup         bool
}

// NewAutoCleaner 创建一个新的自动资源清理器
func NewAutoCleaner() *AutoCleaner {
	return &AutoCleaner{
		resourceManager: NewResourceManager(),
		signalChan:      make(chan os.Signal, 1),
		cleanupDone:     make(chan struct{}),
	}
}

// Setup 设置信号处理
func (ac *AutoCleaner) Setup() {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	if ac.isSetup {
		return
	}

	// 监听中断信号
	signal.Notify(ac.signalChan, os.Interrupt, syscall.SIGTERM)

	// 启动信号处理协程
	go func() {
		<-ac.signalChan
		fmt.Println("\n正在清理临时资源...")
		ac.Cleanup()
		close(ac.cleanupDone)
		os.Exit(0)
	}()

	ac.isSetup = true
}

// AddResource 添加资源
func (ac *AutoCleaner) AddResource(resource Resource) {
	ac.resourceManager.AddResource(resource)
}

// AddFile 添加文件资源
func (ac *AutoCleaner) AddFile(path string, priority int) {
	ac.resourceManager.AddFile(path, priority)
}

// AddDirectory 添加目录资源
func (ac *AutoCleaner) AddDirectory(path string, priority int) {
	ac.resourceManager.AddDirectory(path, priority)
}

// AddCustom 添加自定义资源
func (ac *AutoCleaner) AddCustom(cleanup CleanupFunc, priority int) {
	ac.resourceManager.AddCustom(cleanup, priority)
}

// Cleanup 清理所有资源
func (ac *AutoCleaner) Cleanup() []error {
	return ac.resourceManager.Cleanup()
}

// Wait 等待清理完成
func (ac *AutoCleaner) Wait() {
	<-ac.cleanupDone
}

// GetResourceCount 获取资源数量
func (ac *AutoCleaner) GetResourceCount() int {
	return ac.resourceManager.GetResourceCount()
}

// DefaultAutoCleaner 默认的自动资源清理器实例
var DefaultAutoCleaner = NewAutoCleaner()

// SetupDefaultAutoCleaner 设置默认的自动资源清理器
func SetupDefaultAutoCleaner() {
	DefaultAutoCleaner.Setup()
}

// AddFileToAutoClean 添加文件到默认的自动资源清理器
func AddFileToAutoClean(path string, priority int) {
	DefaultAutoCleaner.AddFile(path, priority)
}

// AddDirectoryToAutoClean 添加目录到默认的自动资源清理器
func AddDirectoryToAutoClean(path string, priority int) {
	DefaultAutoCleaner.AddDirectory(path, priority)
}

// AddCustomToAutoClean 添加自定义资源到默认的自动资源清理器
func AddCustomToAutoClean(cleanup CleanupFunc, priority int) {
	DefaultAutoCleaner.AddCustom(cleanup, priority)
}

// CleanupAll 清理默认的自动资源清理器中的所有资源
func CleanupAll() []error {
	return DefaultAutoCleaner.Cleanup()
}

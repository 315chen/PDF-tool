package file

import (
	"fmt"
	"os"
	"sync"
)

// ResourceType 定义资源类型
type ResourceType int

const (
	// ResourceFile 表示文件资源
	ResourceFile ResourceType = iota
	// ResourceDir 表示目录资源
	ResourceDir
	// ResourceCustom 表示自定义资源
	ResourceCustom
)

// CleanupFunc 定义资源清理函数
type CleanupFunc func() error

// Resource 定义需要管理的资源
type Resource struct {
	Type     ResourceType
	Path     string
	Cleanup  CleanupFunc
	Priority int // 清理优先级，数字越大优先级越高
}

// ResourceManager 管理应用程序资源
type ResourceManager struct {
	resources []Resource
	mutex     sync.Mutex
}

// NewResourceManager 创建一个新的资源管理器
func NewResourceManager() *ResourceManager {
	rm := &ResourceManager{
		resources: make([]Resource, 0),
	}
	return rm
}

// AddResource 添加资源到管理器
func (rm *ResourceManager) AddResource(resource Resource) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.resources = append(rm.resources, resource)
}

// AddFile 添加文件资源
func (rm *ResourceManager) AddFile(path string, priority int) {
	rm.AddResource(Resource{
		Type:     ResourceFile,
		Path:     path,
		Priority: priority,
		Cleanup: func() error {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("无法删除文件 %s: %v", path, err)
			}
			return nil
		},
	})
}

// AddDirectory 添加目录资源
func (rm *ResourceManager) AddDirectory(path string, priority int) {
	rm.AddResource(Resource{
		Type:     ResourceDir,
		Path:     path,
		Priority: priority,
		Cleanup: func() error {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("无法删除目录 %s: %v", path, err)
			}
			return nil
		},
	})
}

// AddCustom 添加自定义资源
func (rm *ResourceManager) AddCustom(cleanup CleanupFunc, priority int) {
	rm.AddResource(Resource{
		Type:     ResourceCustom,
		Priority: priority,
		Cleanup:  cleanup,
	})
}

// Cleanup 清理所有资源
func (rm *ResourceManager) Cleanup() []error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 按优先级排序资源
	resources := make([]Resource, len(rm.resources))
	copy(resources, rm.resources)

	// 按优先级从高到低排序
	for i := 0; i < len(resources); i++ {
		for j := i + 1; j < len(resources); j++ {
			if resources[i].Priority < resources[j].Priority {
				resources[i], resources[j] = resources[j], resources[i]
			}
		}
	}

	var errors []error
	for _, resource := range resources {
		if err := resource.Cleanup(); err != nil {
			errors = append(errors, err)
		}
	}

	// 清空资源列表
	rm.resources = rm.resources[:0]

	return errors
}

// CleanupResource 清理特定资源
func (rm *ResourceManager) CleanupResource(path string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	for i, resource := range rm.resources {
		if resource.Path == path {
			err := resource.Cleanup()
			// 从列表中移除资源
			rm.resources = append(rm.resources[:i], rm.resources[i+1:]...)
			return err
		}
	}

	return fmt.Errorf("资源不存在: %s", path)
}

// GetResourceCount 获取资源数量
func (rm *ResourceManager) GetResourceCount() int {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	return len(rm.resources)
}

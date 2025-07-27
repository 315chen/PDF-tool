package file

import (
	"os"
	"path/filepath"
)

// GetDirectoryFromPath 从文件路径中提取目录部分
func GetDirectoryFromPath(path string) string {
	return filepath.Dir(path)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists 检查目录是否存在
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

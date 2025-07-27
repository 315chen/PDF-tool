package pdf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// RollbackManager 回滚与恢复管理器
// 支持文件级备份、恢复、回滚

type RollbackManager struct {
	backupDir string
	mutex     sync.Mutex
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager(backupDir string) *RollbackManager {
	return &RollbackManager{backupDir: backupDir}
}

// BackupFile 备份文件，返回备份路径
func (rm *RollbackManager) BackupFile(filePath string) (string, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("待备份文件不存在: %s", filePath)
	}
	backupName := filepath.Base(filePath) + ".bak"
	backupPath := filepath.Join(rm.backupDir, backupName)
	if err := copyFileForRollback(filePath, backupPath); err != nil {
		return "", fmt.Errorf("备份失败: %v", err)
	}
	return backupPath, nil
}

// RestoreFile 恢复文件（用备份覆盖原文件）
func (rm *RollbackManager) RestoreFile(backupPath, targetPath string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("备份文件不存在: %s", backupPath)
	}
	return copyFileForRollback(backupPath, targetPath)
}

// RollbackIfFailed 操作失败时自动回滚
func (rm *RollbackManager) RollbackIfFailed(targetPath string, op func() error) error {
	backupPath, err := rm.BackupFile(targetPath)
	if err != nil {
		return fmt.Errorf("回滚前备份失败: %v", err)
	}
	err = op()
	if err != nil {
		_ = rm.RestoreFile(backupPath, targetPath)
		return fmt.Errorf("操作失败，已回滚: %v", err)
	}
	return nil
}

// copyFileForRollback 回滚专用文件复制函数
func copyFileForRollback(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

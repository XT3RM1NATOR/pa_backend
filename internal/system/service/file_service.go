package service

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileServiceImpl struct {
	BaseDir string
}

func NewFileServiceImpl(baseDir string) *FileServiceImpl {
	return &FileServiceImpl{BaseDir: baseDir}
}

func (fs *FileServiceImpl) SaveFile(filename string, content []byte) error {
	filePath := filepath.Join(fs.BaseDir, filename+".jpg")
	return os.WriteFile(filePath, content, 0644)
}

func (fs *FileServiceImpl) LoadFile(filename string) ([]byte, error) {
	filePath := filepath.Join(fs.BaseDir, filename+".jpg")
	return os.ReadFile(filePath)
}

func (fs *FileServiceImpl) UpdateFileName(oldName, newName string) error {
	oldPath := filepath.Join(fs.BaseDir, oldName+".jpg")
	newPath := filepath.Join(fs.BaseDir, newName+".jpg")

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", oldName)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename file from %s to %s: %w", oldName, newName, err)
	}

	return nil
}

func (fs *FileServiceImpl) UpdateFile(newFileBytes []byte, fileName string) error {
	filePath := filepath.Join(fs.BaseDir, fileName+".jpg")

	return os.WriteFile(filePath, newFileBytes, 0644)
}

func (fs *FileServiceImpl) DeleteFile(filename string) error {
	filePath := filepath.Join(fs.BaseDir, filename+".jpg")
	return os.Remove(filePath)
}

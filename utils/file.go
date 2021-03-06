package utils

import (
	"io"
	"os"

	"github.com/google/logger"
)

// 读取文件
func ReadFile(path string) []byte {
	fileInfo, err := os.OpenFile(path, os.O_RDONLY, 0600)

	if err != nil {
		logger.Errorln("打开文件出错:", err)
		return nil
	}
	data, err := io.ReadAll(fileInfo)

	if err != nil {
		logger.Errorln("读取文件出错:", err)
		return nil
	}
	return data
}

// 检测文件是否存在
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// 创建文件
func CreateFile(dstPath string, content []byte) error {
	fileInfo, err := os.Create(dstPath)

	if err != nil {
		logger.Errorln("创建文件失败:", err)
		return err
	} else {
		fileInfo.Write(content)
	}
	fileInfo.Close()
	logger.Info("创建文件成功")
	return nil
}

// 复制文件
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

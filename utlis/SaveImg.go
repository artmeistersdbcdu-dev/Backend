package utlis

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type Save interface {
	SaveLocal()
}

func isPath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false

}

func SaveLocal(file multipart.File, fileName string, path string) (string, error) {
	is := isPath(path)
	if !is {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	filePath := fmt.Sprintf("%s/%s.png", path, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
func deleteFile(path string) error {
	err := os.Remove(path)
	return err
}
func deleteFolder(path string) error {
	err := os.RemoveAll(path)
	return err
}
func DeleteLocal(path string) error {
	mainPath := "uploads/" + path
	info, err := os.Stat(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}
		return err
	}
	if info.IsDir() {
		return deleteFolder(mainPath)
	}
	err = deleteFile(mainPath)
	return err
}

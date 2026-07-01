package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var path string = "logs"
var appName string = "Art Meister"

type Logger struct {
	service string
}

var (
	log    *Logger
	once   sync.Once
	logErr error
)

func GetLogger() (*Logger, error) {
	once.Do(func() {
		log, logErr = loadLogger(appName)
	})
	return log, logErr
}
func isPath(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false

}
func createFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}
func initLogsFile() error {
	files := []string{
		"err.log",
		"info.log",
	}
	for _, f := range files {
		filePath := fmt.Sprintf("%s/%s", path, f)
		_, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				err = createFile(filePath)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}
func loadLogger(str string) (*Logger, error) {
	ok := isPath(path)
	if !ok {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	err := initLogsFile()
	if err != nil {
		return nil, err
	}
	return &Logger{
		service: str,
	}, nil
}
func (l *Logger) Info(msg string) {

	filePath := fmt.Sprintf("%s/%s", path, "info.log")
	file, err := os.OpenFile(
		filePath,
		os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return
	}
	defer file.Close()
	logLine := fmt.Sprintf(
		"[%s] [INFO] [%s] %s\n",
		time.Now().Format(time.RFC3339),
		l.service,
		msg,
	)
	file.WriteString(logLine)

}

func (l *Logger) Error(msg string) {

	filePath := fmt.Sprintf("%s/%s", path, "err.log")
	file, err := os.OpenFile(
		filePath,
		os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return
	}
	defer file.Close()
	logLine := fmt.Sprintf(
		"[%s] [ERROR] [%s] %s\n",
		time.Now().Format(time.RFC3339),
		l.service,
		msg,
	)
	file.WriteString(logLine)
}

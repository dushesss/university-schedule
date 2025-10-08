package logger

import (
	"log"
	"os"
	"path/filepath"
)

// SetupLogger создаёт простой логгер
func SetupLogger(logFile string) *log.Logger {
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		log.Fatal("Не удалось создать папку для логов:", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Не удалось открыть файл логов:", err)
	}

	return log.New(file, "", log.LstdFlags)
}

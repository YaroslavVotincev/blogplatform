package filelogger

import (
	"log"
	"log/slog"
	"os"
)

type FileLogger struct {
	fileLogger    *slog.Logger
	consoleLogger *slog.Logger
	consoleLog    bool
}

func NewFileLogger(filepath string) *FileLogger {
	handlerOptions := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}
	consoleHandler := slog.NewTextHandler(os.Stdout, &handlerOptions)
	consoleLogger := slog.New(consoleHandler)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error occurred while opening log file: ", err)
	}
	fileHandler := slog.NewTextHandler(file, &handlerOptions)
	fileLogger := slog.New(fileHandler)

	return &FileLogger{fileLogger: fileLogger, consoleLogger: consoleLogger, consoleLog: true}

}

func (f *FileLogger) EnableConsoleLog() {
	f.consoleLog = true
}

func (f *FileLogger) DisableConsoleLog() {
	f.consoleLog = false
}

func (f *FileLogger) Debug(msg string, params map[string]interface{}) {
	keyValues := make([]interface{}, 0)
	for k, v := range params {
		keyValues = append(keyValues, k, v)
	}
	if f.consoleLog {
		f.consoleLogger.Debug(msg, keyValues...)
	}
	f.fileLogger.Debug(msg, keyValues...)
}

func (f *FileLogger) Info(msg string, params map[string]interface{}) {
	keyValues := make([]interface{}, 0)
	for k, v := range params {
		keyValues = append(keyValues, k, v)
	}
	if f.consoleLog {
		f.consoleLogger.Info(msg, keyValues...)
	}
	f.fileLogger.Info(msg, keyValues...)
}

func (f *FileLogger) Warn(msg string, params map[string]interface{}) {
	keyValues := make([]interface{}, 0)
	for k, v := range params {
		keyValues = append(keyValues, k, v)
	}
	if f.consoleLog {
		f.consoleLogger.Warn(msg, keyValues...)
	}
	f.fileLogger.Warn(msg, keyValues...)
}

func (f *FileLogger) Error(msg string, params map[string]interface{}) {
	keyValues := make([]interface{}, 0)
	for k, v := range params {
		keyValues = append(keyValues, k, v)
	}
	if f.consoleLog {
		f.consoleLogger.Error(msg, keyValues...)
	}
	f.fileLogger.Error(msg, keyValues...)
}

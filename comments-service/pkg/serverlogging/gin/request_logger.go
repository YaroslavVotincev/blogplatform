package serverlogging

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	LogLevelNone  = "none"
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// NewRequestLogger returns a gin.HandlerFunc that logs HTTP responses.
//
// It takes a FileLogger and a QueueLogger as parameters.
func NewRequestLogger(fileLogger FileLogger, queueLogger QueueLogger) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next()
		contextCopy := context.Copy()
		go func() {
			loggingMap := GetLoggingMap(contextCopy)
			status := contextCopy.Writer.Status()
			loggingMap["response_status"] = status
			loggingMap["request_method"] = contextCopy.Request.Method
			loggingMap["request_path"] = contextCopy.Request.URL.String()
			userId := loggingMap.GetUserId()
			logLevel := loggingMap.getLogLevel()

			if logLevel != nil {
				switch *logLevel {
				case LogLevelDebug:
					logDebug(fileLogger, queueLogger, userId, loggingMap)
				case LogLevelInfo:
					logInfo(fileLogger, queueLogger, userId, loggingMap)
				case LogLevelWarn:
					logWarn(fileLogger, queueLogger, userId, loggingMap)
				case LogLevelError:
					logError(fileLogger, queueLogger, userId, loggingMap)
				default:
					return
				}
				return
			}

			switch {
			case status == 200 || status >= 300 && status < 400:
				logDebug(fileLogger, queueLogger, userId, loggingMap)
			case status > 200 && status < 300:
				logInfo(fileLogger, queueLogger, userId, loggingMap)
			case status >= 400 && status < 500:
				logWarn(fileLogger, queueLogger, userId, loggingMap)
			default:
				logError(fileLogger, queueLogger, userId, loggingMap)
			}
		}()
	}
}

// GetLoggingMap retrieves the LoggingMap from the gin context.
func GetLoggingMap(ctx *gin.Context) LoggingMap {
	obj, ok := ctx.Get("logging_map")
	if !ok {
		logInfo := make(map[string]any)
		ctx.Set("logging_map", logInfo)
		return logInfo
	}
	logInfo, ok := obj.(map[string]any)
	if !ok {
		logInfo := make(map[string]any)
		ctx.Set("logging_map", logInfo)
		return logInfo
	}
	return logInfo
}

func logDebug(fileLogger FileLogger, queueLogger QueueLogger, userId *uuid.UUID, logInfo LoggingMap) {
	fileLogger.Debug("handled request", logInfo)
	err := queueLogger.Debug(userId, logInfo)
	if err != nil {
		fileLogger.Error("failed to push log message to mq", logInfo)
	}
}

func logInfo(fileLogger FileLogger, queueLogger QueueLogger, userId *uuid.UUID, logInfo LoggingMap) {
	fileLogger.Info("handled request", logInfo)
	err := queueLogger.Info(userId, logInfo)
	if err != nil {
		fileLogger.Error("failed to push log message to mq", logInfo)
	}
}

func logWarn(fileLogger FileLogger, queueLogger QueueLogger, userId *uuid.UUID, logInfo LoggingMap) {
	fileLogger.Warn("handled request", logInfo)
	err := queueLogger.Warn(userId, logInfo)
	if err != nil {
		fileLogger.Error("failed to push log message to mq", logInfo)
	}
}

func logError(fileLogger FileLogger, queueLogger QueueLogger, userId *uuid.UUID, logInfo LoggingMap) {
	fileLogger.Error("handled request", logInfo)
	err := queueLogger.Error(userId, logInfo)
	if err != nil {
		fileLogger.Error("failed to push log message to mq", logInfo)
	}
}

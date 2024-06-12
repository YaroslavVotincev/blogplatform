package serverlogging

import "github.com/google/uuid"

type FileLogger interface {
	Info(msg string, params map[string]any)
	Debug(msg string, params map[string]any)
	Warn(msg string, params map[string]any)
	Error(msg string, params map[string]any)
}

type QueueLogger interface {
	Info(userId *uuid.UUID, data map[string]any) error
	Debug(userId *uuid.UUID, data map[string]any) error
	Warn(userId *uuid.UUID, data map[string]any) error
	Error(userId *uuid.UUID, data map[string]any) error
}

type LoggingMap map[string]any

// SetUserId sets the user ID in the LoggingMap.
func (loggingMap LoggingMap) SetUserId(userId *uuid.UUID) {
	if userId != nil {
		loggingMap["user_id"] = userId.String()
	}
}

// GetUserId returns the user ID from the LoggingMap.
func (loggingMap LoggingMap) GetUserId() *uuid.UUID {
	obj, ok := loggingMap["user_id"]
	if ok {
		strObj, ok := obj.(string)
		if ok {
			userId, err := uuid.Parse(strObj)
			if err != nil {
				return nil
			}
			return &userId
		} else {
			return nil
		}
	} else {
		return nil
	}
}

// SetMessage sets the message in the LoggingMap.
func (loggingMap LoggingMap) SetMessage(message string) {
	loggingMap["message"] = message
}

// SetError sets the error in the LoggingMap.
func (loggingMap LoggingMap) SetError(message string) {
	loggingMap["error"] = message
}

// None sets the logging level to none in the LoggingMap.
func (loggingMap LoggingMap) None() {
	loggingMap["logging_level"] = LogLevelNone
}

// Debug sets the logging level to debug in the LoggingMap.
func (loggingMap LoggingMap) Debug() {
	loggingMap["logging_level"] = LogLevelDebug
}

// Info sets the logging level to info in the LoggingMap.
func (loggingMap LoggingMap) Info() {
	loggingMap["logging_level"] = LogLevelInfo
}

// Warn sets the logging level to warn in the LoggingMap.
func (loggingMap LoggingMap) Warn() {
	loggingMap["logging_level"] = LogLevelWarn
}

// Error sets the logging level to error in the LoggingMap.
func (loggingMap LoggingMap) Error() {
	loggingMap["logging_level"] = LogLevelError
}

func (loggingMap LoggingMap) getLogLevel() *string {
	obj, ok := loggingMap["logging_level"]
	if ok {
		logLevelString, ok := obj.(string)
		if !ok {
			return nil
		}
		switch logLevelString {
		case LogLevelNone, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
			return &logLevelString
		default:
			return nil
		}
	}
	return nil
}

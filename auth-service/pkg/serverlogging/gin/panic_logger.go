package serverlogging

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
)

// NewPanicLogger returns a function that logs a panic and aborts the request with a 500 status code.
func NewPanicLogger(fileLogger FileLogger, queueLogger QueueLogger) func(c *gin.Context, err any) {
	return func(ctx *gin.Context, err any) {
		loggingMap := GetLoggingMap(ctx)
		loggingMap["message"] = "runtime error occurred"
		loggingMap["error"] = fmt.Sprintf("%v", err)
		loggingMap["stack_trace"] = string(debug.Stack())
		loggingMap["response_status"] = http.StatusInternalServerError
		loggingMap["request_method"] = ctx.Request.Method
		loggingMap["request_path"] = ctx.Request.URL.String()
		userId := loggingMap.GetUserId()
		logError(fileLogger, queueLogger, userId, loggingMap)
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

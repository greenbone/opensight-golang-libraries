package ginMiddleware

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/greenbone/opensight-golang-libraries/pkg/logs"
	"github.com/rs/zerolog"
)

const (
	// correlationIDHeader is the header key for the correlation ID
	correlationIDHeader = "X-Correlation-ID"
)

// Logging returns a Gin ginMiddleware handler that injects structured logging
// using zerolog into the request lifecycle. It ensures each request is assigned
// a correlation ID, which is included in both the request context and the response
// headers.
//
// This is useful for tracing and debugging distributed systems, where tracking
// requests across services is critical.
func Logging() gin.HandlerFunc {
	logHandler := logger.SetLogger(
		logger.WithLogger(func(c *gin.Context, _ zerolog.Logger) zerolog.Logger {
			correlationID := c.GetHeader(correlationIDHeader)
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			ctx := logs.WithCorrelationID(c.Request.Context(), correlationID)
			// update request context
			c.Request = c.Request.WithContext(ctx)

			// add correlation ID to response header
			c.Writer.Header().Set(correlationIDHeader, correlationID)

			// return context logger
			return *zerolog.Ctx(ctx)
		}),

		logger.WithUTC(true),
		// exclude alive endpoint to avoid log spam
		logger.WithSkipPath([]string{"/health/alive"}),
	)
	return logHandler
}

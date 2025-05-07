// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logs provides a thin wrapper around the zerolog logging library,
// enabling structured, context logging across an application.
//
// It includes:
//   - Setting the global logging level.
//   - Extracting loggers from context, with fallback to the global logger.
//   - Attaching metadata (e.g., job name, correlation IDs) to loggers via context.
//
// This package ensures consistent logging behavior and makes it easy to trace
// log output using structured fields like correlation IDs, even across goroutines or services.
package logs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	AllowedLogLevels string = "disabled|trace|debug|info|warn|error|fatal|panic"

	correlationIDKey = "correlation_id"
	jobKey           = "job"
)

// SetupLogger configures the global log level for the zerolog logger.
//
// It takes a string `logLevel` representing the desired logging level (e.g. "debug", "info", "warn").
// If the provided log level is invalid, it returns an error including the list of allowed levels.
//
// Parameters:
//   - logLevel: a string indicating the desired log verbosity level.
//
// Returns:
//   - error: if the log level cannot be parsed, an error is returned; otherwise, nil.
//
// Example:
//
//	err := SetupLogger("info")
//	if err != nil {
//	    log.Fatal(err)
//	}
func SetupLogger(logLevel string) error {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed parsing log level: %w; only allowed levels are: %s", err, AllowedLogLevels)
	}

	zerolog.SetGlobalLevel(level)
	return nil
}

// Ctx retrieves the zerolog logger from the given context.
//
// If the context does not have a logger attached, zerolog.Ctx(ctx) returns a disabled logger.
// In such cases, this function returns the global logger instead.
//
// This ensures that a valid logger is always returned, preventing logging from being silently disabled.
//
// Parameters:
//   - ctx: the context from which to retrieve the logger.
//
// Returns:
//   - *zerolog.Logger: the logger extracted from the context, or the global logger if none is found.
//
// Example:
//
//	logger := Ctx(r.Context())
//	logger.Info().Msg("request received")
func Ctx(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx)
	if l.GetLevel() == zerolog.Disabled {
		// zerolog.Ctx() can return disabled logger, if no logger was previously attached to the context,
		// in such case return global logger
		return &log.Logger
	}
	return l
}

// WithCtxField adds a key-value pair to the logger in the given context and returns the updated context.
//
// If the context does not contain a logger, a new logger with the key-value pair is created
// and attached to the context.
//
// Parameters:
//   - ctx: the context to enrich with the key-value field.
//   - key: the field key to add to the logger.
//   - value: the field value to add to the logger.
//
// Returns:
//   - context.Context: the updated context containing the enriched logger.
//
// Example:
//
//	ctx = WithCtxField(ctx, "request_id", "abc123")
func WithCtxField(ctx context.Context, key, value string) context.Context {
	// Retrieve the ctxLogger from the context
	ctxLogger := zerolog.Ctx(ctx)

	// Check if the logger is disabled
	if ctxLogger.GetLevel() == zerolog.Disabled {
		// Create a new logger with the key-value pair
		return log.With().Str(key, value).Logger().WithContext(ctx)
	}
	return ctxLogger.With().Str(key, value).Logger().WithContext(ctx)
}

// WithJob adds a job name field to the logger in the context and returns the updated context.
//
// Internally calls WithCtxField with the appropriate key.
func WithJob(ctx context.Context, jobName string) context.Context {
	return WithCtxField(ctx, jobKey, jobName)
}

// WithCorrelationID adds a correlation ID to the logger in the context and returns the updated context.
//
// Useful for tracing requests across services or components.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return WithCtxField(ctx, correlationIDKey, correlationID)
}

// WithNewCorrelationID generates a new correlation ID, adds it to the logger in the context,
// and returns the updated context.
//
// This is useful for initiating a trace when handling a new incoming request or job.
func WithNewCorrelationID(ctx context.Context) context.Context {
	return WithCorrelationID(ctx, uuid.New().String())
}

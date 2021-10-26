package logging

import (
	log "github.com/golangee/log"
)

// NewLoggerFromEnv creates a new logger based on the current environment.
func NewLoggerFromEnv() log.Logger {
	logger := log.NewLogger()
	logger = log.WithFields(logger, log.V("build_id", "tbd"), log.V("build_tag", "tbd"))

	return logger
}

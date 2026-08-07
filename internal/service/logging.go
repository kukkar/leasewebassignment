package service

import (
	"context"
	"io"
	"time"

	"go.uber.org/zap"
)

type loggingService struct {
	Service
	logger *zap.SugaredLogger
}

// NewLoggingService wraps next with logging. If logger is nil, next is
// returned unwrapped rather than logging to a no-op sink - there's no
// reason to pay the decorator's overhead if nothing will observe it.
func NewLoggingService(next Service, logger *zap.SugaredLogger) Service {
	if logger == nil {
		return next
	}
	return &loggingService{Service: next, logger: logger}
}

func (s *loggingService) UploadServerData(ctx context.Context, filename string, reader io.Reader) error {
	start := time.Now()
	err := s.Service.UploadServerData(ctx, filename, reader)
	logAtLevel(s.logger, err, "UploadServerData",
		"filename", filename,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", errString(err),
	)
	return err
}

func (s *loggingService) LoadServerData(ctx context.Context, path string) error {
	start := time.Now()
	err := s.Service.LoadServerData(ctx, path)
	logAtLevel(s.logger, err, "LoadServerData",
		"path", path,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", errString(err),
	)
	return err
}

// logAtLevel logs at Info on success and Error on failure - uploads and
// startup loads are rare enough that every one is worth a durable,
// always-on log line rather than a debug-only one.
func logAtLevel(logger *zap.SugaredLogger, err error, msg string, kv ...any) {
	if err != nil {
		logger.Errorw(msg, kv...)
		return
	}
	logger.Infow(msg, kv...)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

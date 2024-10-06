package log

import (
	"log/slog"
	"os"
	"strings"
)

func NewDefaultService(level string, appID string, hostname string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: stringToLevel(level),
	}))
	return logger.With(
		slog.String("app_id", appID),
		slog.String("hostname", hostname),
	)
}

func stringToLevel(s string) slog.Level {
	var levelToInt = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	lower := strings.ToLower(s)
	l, ok := levelToInt[lower]
	if !ok {
		return slog.LevelInfo
	}
	return l
}

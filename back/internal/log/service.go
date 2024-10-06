package log

import "log/slog"

type Logger interface {
	slog.Logger
}

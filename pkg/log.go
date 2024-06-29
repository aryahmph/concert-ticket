package pkg

import (
	"context"
	"fmt"
	"log/slog"
)

func LogEventName(name string) slog.Attr {
	return slog.Any("name", name)
}

func LogStatusFailed(key string) slog.Attr {
	return slog.Any(key, "FAILED")
}

func LogStatusSuccess(key string) slog.Attr {
	return slog.Any(key, "SUCCESS")
}

func LogEventState(state string) slog.Attr {
	return slog.Any("state", state)
}

func LogEventPayload(payload interface{}) slog.Attr {
	return slog.Any("payload", payload)
}

func LogInfoWithContext(ctx context.Context, msg string, attrs []slog.Attr) {
	slog.InfoContext(ctx, msg, slog.Any("event", attrs))
}

func LogWarnWithContext(ctx context.Context, msg string, err error, attrs []slog.Attr) {
	slog.WarnContext(ctx, fmt.Sprintf("%s, err: %v", msg, err), slog.Any("event", attrs))
}

func LogErrorWithContext(ctx context.Context, err error, attrs []slog.Attr) {
	slog.ErrorContext(ctx, err.Error(), slog.Any("event", attrs))
}

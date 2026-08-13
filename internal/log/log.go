package log

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func NewHandler(w io.Writer, level slog.Level) slog.Handler {
	return slog.NewTextHandler(
		w,
		&slog.HandlerOptions{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: replaceAttr,
		},
	)
}

func replaceAttr(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.TimeKey:
		return slog.String(
			slog.TimeKey,
			attr.Value.Time().Format("15:04:05"),
		)

	case slog.SourceKey:
		source := attr.Value.Any().(*slog.Source)

		return slog.String(
			slog.SourceKey,
			formatSource(source),
		)
	}

	return attr
}

func formatSource(source *slog.Source) string {
	function := source.Function

	if index := strings.LastIndex(function, "/"); index >= 0 {
		function = function[index+1:]
	}

	if index := strings.Index(function, ".("); index >= 0 {
		if end := strings.Index(function[index:], ")."); end >= 0 {
			end += index
			function = function[:index] + function[end+2:]
		}
	}

	return fmt.Sprintf("%s:%d", function, source.Line)
}

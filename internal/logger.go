package internal

import (
	"log/slog"
	"os"
	"path/filepath"
)

func ConfigureLogger() {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: replace, AddSource: true}))
	slog.SetDefault(logger)
}

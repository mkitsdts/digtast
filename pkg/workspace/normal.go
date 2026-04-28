package workspace

import (
	"fmt"
	"log/slog"
	"os"
)

func PersistFile(path string, content []byte) {
	realpath := fmt.Sprintf("%s/%s", GetCurrentWorkspacePath(), path)
	if err := os.WriteFile(realpath, content, 0644); err != nil {
		slog.Error("cant persist file", "error", err)
	}
	slog.Info("file persisted", "path", realpath)
}

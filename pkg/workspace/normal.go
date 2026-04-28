package workspace

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func PersistFile(path string, content []byte) {
	realpath := fmt.Sprintf("%s/%s", GetWorkspacePath(), path)
	dirPath := filepath.Dir(realpath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		slog.Error("cant create persist file", "path", realpath, "err", err)
	}
	if err := os.WriteFile(realpath, content, 0644); err != nil {
		slog.Error("cant persist file", "path", realpath, "err", err)
	}
	slog.Debug("file persisted", "path", realpath)
}

func AppendFile(path string, content []byte) {
	realpath := fmt.Sprintf("%s/%s", GetWorkspacePath(), path)
	dirPath := filepath.Dir(realpath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		slog.Error("cant create persist file directory", "path", dirPath, "err", err)
		return
	}

	f, err := os.OpenFile(realpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("cant open persist file for append", "path", realpath, "err", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(append(content, '\n')); err != nil {
		slog.Error("cant append to persist file", "path", realpath, "err", err)
	}
	slog.Debug("file appended", "path", realpath)
}

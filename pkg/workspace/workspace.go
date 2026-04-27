package workspace

import (
	"log/slog"
	"os"
	"path"
)

func GetWorkspacePath() string {
	workspacePath, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	if workspacePath == "" {
		workspacePath, err = os.Getwd()
		if err != nil {
			return ""
		}
	}

	workspacePath = path.Join(workspacePath, "/.digtast")

	return workspacePath
}

func GetCurrentWorkspacePath() string {
	wd, err := os.Getwd()
	if err != nil {
		slog.Error("cant get current workspace path", "error", err)
		return ""
	}
	return wd
}

func InitWorkspace() {
	// TODO:
	path := GetWorkspacePath()
	if err := os.MkdirAll(path, 0755); err != nil {
		slog.Error("Failed to create workspace", "path", path, "err", err)
		return
	}

	if err := os.MkdirAll(path+"/prompt", 0755); err != nil {
		slog.Error("Failed to create prompt directory", "path", path+"/prompt", "err", err)
		return
	}

	for _, promptCreator := range promptCreators {
		err := promptCreator.CreatePromptImpl()
		if err != nil {
			slog.Error("Failed create prompt impl", "name", promptCreator.GetPromptName())
		}
	}
}

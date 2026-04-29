package workspace

import (
	"digital-labor/pkg/conf"
	"fmt"
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
	path := GetWorkspacePath()
	if err := os.MkdirAll(path, 0755); err != nil {
		slog.Error("Failed to create workspace", "path", path, "err", err)
		return
	}

	workspaceRoot := fmt.Sprintf("%s/workspace", path)
	if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
		slog.Error("Failed to create workspace root", "path", workspaceRoot, "err", err)
		return
	}

	// Load or create config.json
	configPath := fmt.Sprintf("%s/config.json", path)
	if err := conf.LoadConfig(configPath); err != nil {
		slog.Error("Failed to load config", "path", configPath, "err", err)
	}

	if err := os.MkdirAll(path+"/prompt", 0755); err != nil {
		slog.Error("Failed to create prompt directory", "path", path+"/prompt", "err", err)
		return
	}

	taskPath := fmt.Sprintf("%s/tasks", path)
	if err := os.MkdirAll(taskPath, 0755); err != nil {
		slog.Error("Failed to create task directory", "path", taskPath, "err", err)
		return
	}

	for _, promptCreator := range promptCreators {
		err := promptCreator.CreatePromptImpl()
		if err != nil {
			slog.Error("Failed create prompt impl", "name", promptCreator.GetPromptName())
		}
	}
}

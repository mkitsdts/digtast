package workspace

import (
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
		fmt.Println("Failed to create workspace", "path", path, "err", err)
		return
	}

	logPath := fmt.Sprintf("%s/app.log", path)
	if err := os.MkdirAll(path, 0755); err != nil {
		slog.Error("Failed to create log directory", "path", logPath, "err", err)
		return
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "path", logPath, "err", err)
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(logFile, nil)))
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	workspaceRoot := fmt.Sprintf("%s/workspace", path)
	if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
		slog.Error("Failed to create workspace root", "path", workspaceRoot, "err", err)
		return
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

	skillPath := fmt.Sprintf("%s/skills", path)
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		slog.Error("Failed to create skill directory", "path", skillPath, "err", err)
		return
	}

	for _, promptCreator := range promptCreators {
		err := promptCreator.CreatePromptImpl()
		if err != nil {
			slog.Error("Failed create prompt impl", "name", promptCreator.GetPromptName())
		}
	}
}

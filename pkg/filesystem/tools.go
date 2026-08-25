package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"digital-labor/pkg/llm"
	toolpkg "digital-labor/pkg/tool"
)

type ToolLoader struct {
	Backend  Backend
	Shell    StreamingShell
	ReadOnly bool
}

func (l ToolLoader) LoadTools(context.Context) ([]toolpkg.Tool, error) {
	if l.Backend == nil {
		return nil, errors.New("filesystem backend is required")
	}
	tools := []toolpkg.Tool{
		&backendTool{name: "list_files", description: "List files in a directory.", backend: l.Backend},
		&backendTool{name: "read_file", description: "Read a text file by line offset and limit.", backend: l.Backend},
		&backendTool{name: "grep_files", description: "Search file contents with a regular expression.", backend: l.Backend},
	}
	if !l.ReadOnly {
		tools = append(tools,
			&backendTool{name: "write_file", description: "Write complete text content to a file.", backend: l.Backend},
			&backendTool{name: "edit_file", description: "Replace an exact string in a text file.", backend: l.Backend},
		)
	}
	if l.Shell != nil && !l.ReadOnly {
		tools = append(tools, &shellTool{shell: l.Shell})
	}
	return tools, nil
}

type backendTool struct {
	name        string
	description string
	backend     Backend
}

func (t *backendTool) Definition(context.Context) (llm.ToolDefinition, error) {
	properties := map[string]any{}
	required := []string{}
	switch t.name {
	case "list_files":
		properties["path"] = map[string]any{"type": "string", "default": "."}
	case "read_file":
		properties["path"] = map[string]any{"type": "string"}
		properties["offset"] = map[string]any{"type": "integer", "minimum": 1}
		properties["limit"] = map[string]any{"type": "integer", "minimum": 1}
		required = append(required, "path")
	case "write_file":
		properties["path"] = map[string]any{"type": "string"}
		properties["content"] = map[string]any{"type": "string"}
		required = append(required, "path", "content")
	case "edit_file":
		properties["path"] = map[string]any{"type": "string"}
		properties["old_string"] = map[string]any{"type": "string"}
		properties["new_string"] = map[string]any{"type": "string"}
		properties["replace_all"] = map[string]any{"type": "boolean"}
		required = append(required, "path", "old_string", "new_string")
	case "grep_files":
		properties["pattern"] = map[string]any{"type": "string"}
		properties["path"] = map[string]any{"type": "string", "default": "."}
		properties["glob"] = map[string]any{"type": "string"}
		properties["case_insensitive"] = map[string]any{"type": "boolean"}
		required = append(required, "pattern")
	}
	return llm.ToolDefinition{Name: t.name, Description: t.description, Parameters: toolpkg.ObjectSchema(properties, required...)}, nil
}

func (t *backendTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	var input struct {
		Path            string `json:"path"`
		Content         string `json:"content"`
		Offset          int    `json:"offset"`
		Limit           int    `json:"limit"`
		OldString       string `json:"old_string"`
		NewString       string `json:"new_string"`
		ReplaceAll      bool   `json:"replace_all"`
		Pattern         string `json:"pattern"`
		Glob            string `json:"glob"`
		CaseInsensitive bool   `json:"case_insensitive"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	if input.Path == "" {
		input.Path = "."
	}
	switch t.name {
	case "list_files":
		files, err := t.backend.LsInfo(ctx, &LsInfoRequest{Path: input.Path})
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(files)
		return &toolpkg.Result{Content: string(data)}, err
	case "read_file":
		content, err := t.backend.Read(ctx, &ReadRequest{FilePath: input.Path, Offset: input.Offset, Limit: input.Limit})
		if err != nil {
			return nil, err
		}
		return &toolpkg.Result{Content: content.Content}, nil
	case "write_file":
		if err := t.backend.Write(ctx, &WriteRequest{FilePath: input.Path, Content: input.Content}); err != nil {
			return nil, err
		}
		return &toolpkg.Result{Content: "file written"}, nil
	case "edit_file":
		if err := t.backend.Edit(ctx, &EditRequest{FilePath: input.Path, OldString: input.OldString, NewString: input.NewString, ReplaceAll: input.ReplaceAll}); err != nil {
			return nil, err
		}
		return &toolpkg.Result{Content: "file edited"}, nil
	case "grep_files":
		matches, err := t.backend.GrepRaw(ctx, &GrepRequest{Pattern: input.Pattern, Path: input.Path, Glob: input.Glob, CaseInsensitive: input.CaseInsensitive})
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(matches)
		return &toolpkg.Result{Content: string(data)}, err
	default:
		return nil, fmt.Errorf("unknown filesystem tool %q", t.name)
	}
}

type shellTool struct{ shell StreamingShell }

func (t *shellTool) Definition(context.Context) (llm.ToolDefinition, error) {
	return llm.ToolDefinition{
		Name: "execute", Description: "Execute a shell command in the configured workspace and return combined output.",
		Parameters: toolpkg.ObjectSchema(map[string]any{
			"command":    map[string]any{"type": "string"},
			"background": map[string]any{"type": "boolean"},
		}, "command"),
	}, nil
}

func (t *shellTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	var input struct {
		Command    string `json:"command"`
		Background bool   `json:"background"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Command) == "" {
		return nil, errors.New("command is required")
	}
	stream, err := t.shell.ExecuteStreaming(ctx, &ExecuteRequest{Command: input.Command, RunInBackendGround: input.Background})
	if err != nil {
		return nil, err
	}
	var output strings.Builder
	exitCode := 0
	for {
		frame, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if frame == nil {
			continue
		}
		output.WriteString(frame.Output)
		if frame.ExitCode != nil {
			exitCode = *frame.ExitCode
		}
	}
	result := &toolpkg.Result{Content: output.String(), Data: map[string]any{"exit_code": exitCode}}
	if exitCode != 0 {
		result.IsError = true
	}
	return result, nil
}

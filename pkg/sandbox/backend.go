package sandbox

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/schema"
)

// SandboxedBackend implements filesystem.Backend and filesystem.StreamingShell,
// with all operations executed inside a Docker container.
type SandboxedBackend struct {
	sandbox *Sandbox
}

// WithSandbox creates a Docker sandbox and returns a backend that executes
// all operations inside the container.
func WithSandbox(ctx context.Context, _ filesystem.Backend, hostWorkPath string) (*SandboxedBackend, error) {
	if hostWorkPath == "" {
		return nil, fmt.Errorf("hostWorkPath is required")
	}

	sb, err := New(ctx, &Config{HostWorkPath: hostWorkPath})
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	slog.Info("sandboxed backend created", "hostWorkPath", hostWorkPath)
	return &SandboxedBackend{sandbox: sb}, nil
}

// --- filesystem.Backend ---

func (sb *SandboxedBackend) LsInfo(ctx context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	path := req.Path
	if path == "" {
		path = "."
	}

	output, err := sb.sandbox.execOutput(ctx, fmt.Sprintf("ls -la %s", shellEscape(path)))
	if err != nil {
		return nil, fmt.Errorf("ls failed: %w", err)
	}

	return parseLsOutput(output), nil
}

func (sb *SandboxedBackend) Read(ctx context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	output, err := sb.sandbox.execOutput(ctx, fmt.Sprintf("cat %s", shellEscape(req.FilePath)))
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	lines := strings.Split(output, "\n")

	offset := req.Offset
	if offset <= 0 {
		offset = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 2000
	}

	start := offset - 1
	if start < 0 {
		start = 0
	}
	end := start + limit
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	return &filesystem.FileContent{
		Content: strings.Join(lines[start:end], "\n"),
	}, nil
}

func (sb *SandboxedBackend) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	// mkdir -p parent dir, then write via stdin
	dir := "."
	if idx := strings.LastIndex(req.FilePath, "/"); idx > 0 {
		dir = req.FilePath[:idx]
	}

	err := sb.sandbox.execWithStdin(ctx,
		fmt.Sprintf("mkdir -p %s && cat > %s", shellEscape(dir), shellEscape(req.FilePath)),
		req.Content)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	return nil
}

func (sb *SandboxedBackend) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	if req.OldString == "" {
		return fmt.Errorf("old string is required")
	}
	if req.OldString == req.NewString {
		return fmt.Errorf("new string must be different from old string")
	}

	// Read the file
	content, err := sb.Read(ctx, &filesystem.ReadRequest{FilePath: req.FilePath, Limit: 999999})
	if err != nil {
		return fmt.Errorf("edit read failed: %w", err)
	}

	text := content.Content
	count := strings.Count(text, req.OldString)
	if count == 0 {
		return fmt.Errorf("string not found in file: '%s'", req.OldString)
	}
	if count > 1 && !req.ReplaceAll {
		return fmt.Errorf("string '%s' appears multiple times. Use replace_all=true to replace all occurrences", req.OldString)
	}

	var newText string
	if req.ReplaceAll {
		newText = strings.Replace(text, req.OldString, req.NewString, -1)
	} else {
		newText = strings.Replace(text, req.OldString, req.NewString, 1)
	}

	return sb.Write(ctx, &filesystem.WriteRequest{
		FilePath: req.FilePath,
		Content:  newText,
	})
}

func (sb *SandboxedBackend) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	if req.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	cmd := []string{"rg", "--json"}
	if req.CaseInsensitive {
		cmd = append(cmd, "-i")
	}
	if req.EnableMultiline {
		cmd = append(cmd, "-U", "--multiline-dotall")
	}
	if req.FileType != "" {
		cmd = append(cmd, "--type", req.FileType)
	} else if req.Glob != "" {
		cmd = append(cmd, "--glob", req.Glob)
	}
	if req.AfterLines > 0 {
		cmd = append(cmd, "-A", strconv.Itoa(req.AfterLines))
	}
	if req.BeforeLines > 0 {
		cmd = append(cmd, "-B", strconv.Itoa(req.BeforeLines))
	}
	cmd = append(cmd, req.Pattern)
	if req.Path != "" {
		cmd = append(cmd, req.Path)
	}

	output, err := sb.sandbox.execOutput(ctx, strings.Join(cmd, " "))
	if err != nil {
		// rg exits 1 when no matches
		if strings.Contains(err.Error(), "exited with code 1") {
			return []filesystem.GrepMatch{}, nil
		}
		return nil, fmt.Errorf("grep failed: %w", err)
	}

	return parseRgOutput(output), nil
}

func (sb *SandboxedBackend) GlobInfo(ctx context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	path := req.Path
	if path == "" {
		path = "."
	}

	// Use find and filter with Go's filepath.Match-like logic
	output, err := sb.sandbox.execOutput(ctx, fmt.Sprintf("find %s -maxdepth 10 2>/dev/null", shellEscape(path)))
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	var files []filesystem.FileInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Get relative path from search root
		relPath := line
		if strings.HasPrefix(line, path+"/") {
			relPath = line[len(path)+1:]
		} else if line == path {
			continue
		}

		// Get file info via stat
		statOutput, err := sb.sandbox.execOutput(ctx, fmt.Sprintf("stat -c '%%s %%Y' %s 2>/dev/null", shellEscape(line)))
		if err != nil {
			continue
		}

		fi := filesystem.FileInfo{
			Path: relPath,
		}

		statParts := strings.Fields(strings.TrimSpace(statOutput))
		if len(statParts) >= 2 {
			fi.Size, _ = strconv.ParseInt(statParts[0], 10, 64)
			if ts, err := strconv.ParseInt(statParts[1], 10, 64); err == nil {
				fi.ModifiedAt = time.Unix(ts, 0).UTC().Format(time.RFC3339)
			}
		}

		// Check if directory
		checkDir, _ := sb.sandbox.execOutput(ctx, fmt.Sprintf("test -d %s && echo d", shellEscape(line)))
		fi.IsDir = strings.TrimSpace(checkDir) == "d"

		files = append(files, fi)
	}

	return files, nil
}

// --- filesystem.StreamingShell ---

func (sb *SandboxedBackend) ExecuteStreaming(ctx context.Context, input *filesystem.ExecuteRequest) (*schema.StreamReader[*filesystem.ExecuteResponse], error) {
	return sb.sandbox.Exec(ctx, input.Command)
}

// Close stops and removes the sandbox container.
func (sb *SandboxedBackend) Close() error {
	return sb.sandbox.Close()
}

// --- helpers ---

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func parseLsOutput(output string) []filesystem.FileInfo {
	var files []filesystem.FileInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "total ") || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		name := strings.Join(fields[8:], " ")
		if name == "." || name == ".." {
			continue
		}

		fi := filesystem.FileInfo{
			Path:  name,
			IsDir: fields[0][0] == 'd',
		}
		fi.Size, _ = strconv.ParseInt(fields[4], 10, 64)

		// Parse date: "Jan  2 15:04" or "Jan  2  2025"
		dateStr := strings.Join(fields[5:8], " ")
		if t, err := time.Parse("Jan 2 15:04", dateStr); err == nil {
			fi.ModifiedAt = t.UTC().Format(time.RFC3339)
		} else if t, err := time.Parse("Jan 2 2006", dateStr); err == nil {
			fi.ModifiedAt = t.UTC().Format(time.RFC3339)
		}

		files = append(files, fi)
	}
	return files
}

type rgJSON struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		LineNumber int `json:"line_number"`
		Lines      struct {
			Text string `json:"text"`
		} `json:"lines"`
	} `json:"data"`
}

func parseRgOutput(output string) []filesystem.GrepMatch {
	var matches []filesystem.GrepMatch
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var data rgJSON
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}
		if data.Type == "match" || data.Type == "context" {
			matches = append(matches, filesystem.GrepMatch{
				Path:    data.Data.Path.Text,
				Line:    data.Data.LineNumber,
				Content: strings.TrimRight(data.Data.Lines.Text, "\n"),
			})
		}
	}
	return matches
}

package local

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// checkSecurity implements multi-level security strategy.
// Returns the resolved absolute path and any error (including security denials).
func (s *Local) checkSecurity(ctx context.Context, op string, target string) (string, error) {
	if op == "execute" {
		if dangerousKeywords.MatchString(target) {
			msg := fmt.Sprintf("Dangerous command detected: %s", target)
			if !s.requestConsent(ctx, msg) {
				return "", fmt.Errorf("user refused to execute dangerous command, please try another way")
			}
		}
		return target, nil
	}

	// Path based operations
	var fullPath string
	if filepath.IsAbs(target) {
		fullPath = filepath.Clean(target)
	} else {
		fullPath = filepath.Join(s.workPath, target)
	}

	// Evaluate symlinks to check for workspace escape
	realPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		// If path doesn't exist, we use the cleaned absolute path for prefix check
		realPath = fullPath
	}

	if s.isOutside(realPath) {
		msg := fmt.Sprintf("Accessing path outside workspace: %s (operation: %s)", realPath, op)
		fmt.Println("ACTION: ", msg)
	}

	return fullPath, nil
}

func (s *Local) isOutside(path string) bool {
	rel, err := filepath.Rel(s.workPath, path)
	if err != nil {
		return true
	}
	return strings.HasPrefix(rel, "..") || filepath.IsAbs(rel)
}

func (s *Local) requestConsent(ctx context.Context, msg string) bool {
	fmt.Printf("\n\033[31m[SECURITY DANGER]\033[0m %s\n", msg)
	fmt.Print("Do you want to allow this operation? (y/N): ")

	// Note: In a real server environment, this should be handled by a more sophisticated
	// interaction mechanism. This is a simple terminal implementation for now.
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	return strings.ToLower(response) == "y"
}

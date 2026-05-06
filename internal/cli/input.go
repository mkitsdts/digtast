package cli

import (
	"fmt"
	"strings"
)

func (s *replState) readLine(prompt string) string {
	fmt.Print(prompt)
	if !s.scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(s.scanner.Text())
}

func (s *replState) promptString(label string) string {
	return s.readLine(fmt.Sprintf("Enter %s: ", label))
}

func (s *replState) promptOptional(label string) string {
	return s.readLine(fmt.Sprintf("Enter %s (optional): ", label))
}

package cli

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func printWelcome() {
	banner := `
    ____  _       _ _        _
   |  _ \(_) __ _(_) |_ __ _| |
   | | | | |/ _` + "`" + ` | |
   | |_| | | (_| | | || (_| | |
   |____/|_|\__, |_|\__\__,_|_|
            |___/`
	fmt.Print(colorCyan + banner + colorReset)
	fmt.Printf("%s%s=== Digital Local Interactive Mode ===%s\n", colorBold, colorCyan, colorReset)
	fmt.Println("Type " + colorYellow + "/help" + colorReset + " for available commands, " + colorYellow + "/exit" + colorReset + " to quit.")
	fmt.Println(strings.Repeat("-", 60))
}

func errorMsg(format string, args ...any) {
	fmt.Printf("%s%s[Error] "+format+"%s\n", append([]any{colorBold, colorRed}, append(args, colorReset)...)...)
}

func successMsg(format string, args ...any) {
	fmt.Printf("%s"+format+"%s\n", append([]any{colorGreen}, append(args, colorReset)...)...)
}

func warnMsg(format string, args ...any) {
	fmt.Printf("%s"+format+"%s\n", append([]any{colorYellow}, append(args, colorReset)...)...)
}

func infoMsg(format string, args ...any) {
	fmt.Printf("%s"+format+"%s\n", append([]any{colorCyan}, append(args, colorReset)...)...)
}

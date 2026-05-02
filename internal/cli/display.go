package cli

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

func printWelcome() {
	fmt.Printf("%s=== Digital Labor Local Interactive Mode ===%s\n", colorCyan, colorReset)
	fmt.Println("Type /help for available commands, /exit to quit.")
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  /ls              - List all agents")
	fmt.Println("  /new             - Create a new agent interactively")
	fmt.Println("  /use <id>        - Switch to a different agent")
	fmt.Println("  /reset           - Clear current agent's conversation history")
	fmt.Println("  /model ls        - List configured model providers")
	fmt.Println("  /model add       - Add a new model provider interactively")
	fmt.Println("  /channel <kind>  - Start a channel (qq, telegram, etc.)")
	fmt.Println("  /exit, /quit     - Exit the application")
	fmt.Println("  /help            - Show this help message")
}

func errorMsg(format string, args ...any) {
	fmt.Printf("%s[Error] "+format+"%s\n", append([]any{colorRed}, append(args, colorReset)...)...)
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

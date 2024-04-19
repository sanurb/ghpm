package ui

import "fmt"

var colorMap = map[string]string{
	"reset":   "\033[0m",
	"black":   "\033[30m",
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
	"bold":    "\033[1m",
}

func Colorize(text string, colorName string, bold bool) string {
	colorCode, ok := colorMap[colorName]
	if !ok {
		colorCode = ""
	}

	resetColor := "\033[0m"
	boldCode := ""
	if bold {
		boldCode = "\033[1m"
	}

	return colorCode + boldCode + text + resetColor
}

func PrintError(message string, err error) {
	fmt.Println(Colorize(message+" "+err.Error(), "red", true))
}

func PrintBanner() {
	fmt.Println(Colorize("  ________  ___ _____________  _____   ", "cyan", true))
	fmt.Println(Colorize(" /  _____/ /   |   \\______   \\/     \\  ", "cyan", true))
	fmt.Println(Colorize("/   \\  ___/    ~    \\     ___/  \\ /  \\ ", "cyan", true))
	fmt.Println(Colorize("\\    \\_\\  \\    Y    /    |  /    Y    \\", "cyan", true))
	fmt.Println(Colorize(" \\______  /\\___|_  /|____|  \\____|__  /", "cyan", true))
	fmt.Println(Colorize("        \\/       \\/                 \\/", "cyan", true))
	fmt.Println()
}

func DisplayMenu() {
	fmt.Println(Colorize("Select an option:", "blue", false))
	fmt.Println(Colorize("(1) Clone own repos (requires GitHub CLI)", "green", false))
	fmt.Println(Colorize("(2) Clone public repos of others", "green", false))
	fmt.Println(Colorize("(3) Run command in all repos", "green", false))
	fmt.Println(Colorize("(4) Set SSH remote", "green", false))
	fmt.Println(Colorize("(0) Exit", "red", false))
}

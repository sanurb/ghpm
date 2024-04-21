package ui

import (
	"fmt"
	"github.com/sanurb/ghpm/internal/github"
	"path/filepath"
	"strings"
)

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
	fmt.Println(Colorize("GitHub Project Manager", "green", true))
}

func DisplayMenu() {
	fmt.Println(Colorize("Select an option:", "blue", false))
	fmt.Println(Colorize("(1) Clone own repos (requires GitHub CLI)", "green", false))
	fmt.Println(Colorize("(2) Clone public repos of others", "green", false))
	fmt.Println(Colorize("(3) Run command in all repos", "green", false))
	fmt.Println(Colorize("(4) Set SSH remote", "green", false))
	fmt.Println(Colorize("(0) Exit", "red", false))
}

func DisplayReposTable(repos []github.RepoInfo) {
	separator := strings.Repeat("-", 142)
	header := fmt.Sprintf("| %-48s | %-68s | %-18s |", "Repository Name", "URL", "Last Updated")

	fmt.Println(Colorize(separator, "cyan", false))
	fmt.Println(Colorize(header, "cyan", true))
	fmt.Println(Colorize(separator, "cyan", false))

	for _, repo := range repos {
		repoName := truncateString(filepath.Base(repo.HTMLURL), 48)
		lastUpdated := repo.PushedAt[:10] // YYYY-MM-DD
		line := fmt.Sprintf("| %-48s | %-68s | %-18s |", repoName, repo.HTMLURL, lastUpdated)
		fmt.Println(Colorize(line, "blue", false))
	}

	fmt.Println(Colorize(separator, "cyan", false))
}

func truncateString(str string, num int) string {
	if len(str) > num {
		return str[:num-3] + "..."
	}
	return str
}

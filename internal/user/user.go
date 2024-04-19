package user

import (
	"bufio"
	"fmt"
	"os"
)

var Username string

func RequestUsername() {
	if Username != "" {
		return
	}
	fmt.Print("Enter GitHub Username: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		Username = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading username:", err)
	}
}

func GetUsername() string {
	return Username
}

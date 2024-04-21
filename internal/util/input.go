package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sanurb/ghpm/internal/ui"
)

func ConfirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/n]: ", message)
	response, err := reader.ReadString('\n')
	if err != nil {
		ui.PrintError("Error reading input:", err)
		os.Exit(1)
	}
	response = strings.TrimSpace(response)
	return strings.ToLower(response) == "y"
}

func GetDaysFromUser() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter the number of days to filter by: ")
	daysStr, err := reader.ReadString('\n')
	if err != nil {
		ui.PrintError("Error reading input:", err)
		os.Exit(1)
	}
	days, err := strconv.Atoi(strings.TrimSpace(daysStr))
	if err != nil {
		ui.PrintError("Invalid input for days, please enter a number:", err)
		os.Exit(1)
	}
	return days
}

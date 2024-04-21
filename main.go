package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sanurb/ghpm/internal/command"
	"github.com/sanurb/ghpm/internal/ui"
	"github.com/sanurb/ghpm/internal/user"
	"github.com/sanurb/ghpm/internal/util"
)

func main() {
	ui.PrintBanner()
	flagSet := flag.NewFlagSet("ghpm", flag.ExitOnError)
	help := flagSet.Bool("h", false, "Display this help message")

	commandActions := map[string]func() command.Command{
		"1": func() command.Command {
			if util.ConfirmAction("Filter by recent activity?") {
				days := util.GetDaysFromUser()
				return &command.FilteredCloneCommand{Days: days}
			}
			return &command.CloneCommand{}
		},
		"2": func() command.Command { return &command.CloneOthersCommand{} },
		"3": func() command.Command { return &command.RunCommand{} },
		"4": func() command.Command { return &command.SetSSHCommand{} },
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		ui.PrintError("Error parsing flags:", err)
		os.Exit(1)
	}

	if *help {
		flagSet.Usage()
		return
	}

	user.RequestUsername()

	for {
		ui.DisplayMenu()

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Select an option: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if action, exists := commandActions[input]; exists {
			cmd := action()
			if err := cmd.Execute(); err != nil {
				ui.PrintError("Error executing command:", err)
			}
		} else if input == "0" {
			fmt.Println("Exiting GHPM")
			break
		} else {
			fmt.Println("Invalid option, please try again.")
		}
	}
}

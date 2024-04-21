package command

import (
	"bufio"
	"fmt"
	"github.com/sanurb/ghpm/internal/github"
	"github.com/sanurb/ghpm/internal/repo"
	"github.com/sanurb/ghpm/internal/ui"
	"os"
	"os/exec"
	"strings"
)

type Command interface {
	Execute() error
}

type CloneCommand struct{}

func (c *CloneCommand) Execute() error {
	return repo.CloneRepos()
}

type SetSSHCommand struct{}

func (c *SetSSHCommand) Execute() error {
	return repo.SetSSHRemote()
}

type RunCommand struct {
	CustomCommand string
}

func (c *RunCommand) Execute() error {
	return repo.ExecuteInRepos(c.CustomCommand)
}

type CloneOthersCommand struct {
	Username string
}

func (c *CloneOthersCommand) Execute() error {
	return repo.CloneOthersRepos(c.Username)
}

type FilteredCloneCommand struct {
	Days int
}

func (c *FilteredCloneCommand) Execute() error {

	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return fmt.Errorf("failed to fetch user repos: %w", err)
	}

	filteredRepos, err := repo.FilterRecentRepos(repos, c.Days)
	if err != nil {
		return fmt.Errorf("failed to filter recent repos: %w", err)
	}

	if len(filteredRepos) == 0 {
		fmt.Println("No recent repositories to clone.")
		return nil
	}

	fmt.Printf("Found %d recent repositories to clone.\n", len(filteredRepos))
	ui.DisplayReposTable(filteredRepos)
	if confirmAction("Proceed with cloning?") {
		for _, repo := range filteredRepos {
			fmt.Printf("Cloning %s...\n", repo.HTMLURL)
			cmd := exec.Command("git", "clone", repo.HTMLURL)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", repo, err)
				continue
			}
		}
	}

	return nil
}

func confirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/n]: ", message)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	response = strings.TrimSpace(response)
	return strings.ToLower(response) == "y"
}

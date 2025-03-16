package ghops

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/sanurb/ghpm/internal/github"
)

// CloneRepo uses the GitHub CLI to clone a repository.
func CloneRepo(url, dest string) error {
	cmd := exec.Command("gh", "repo", "clone", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListSelfRepos returns the authenticated user's repositories.
func ListSelfRepos() ([]github.Repo, error) {
	cmd := exec.Command("gh", "repo", "list", "--json", "name,sshUrl", "-L", "500")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}
	var repos []github.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

// ListPublicRepos returns the public repositories for a given username.
func ListPublicRepos(username string) ([]github.Repo, error) {
	cmd := exec.Command("gh", "repo", "list", username, "--public", "--json", "name,sshUrl", "-L", "500")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list public repos: %w", err)
	}
	var repos []github.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

// RunCommandInAllRepos and SetSSHRemote are placeholders for further implementation.
func RunCommandInAllRepos(customCmd string) error {
	// Implementation omitted for brevity.
	return nil
}

func SetSSHRemote(username string) error {
	// Implementation omitted for brevity.
	return nil
}

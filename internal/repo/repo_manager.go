package repo

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sanurb/ghmp/internal/github"
	"os"
	"os/exec"
	"path/filepath"
)

func CloneRepos() error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return errors.Wrap(err, "failed to fetch user repos")
	}

	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo)
		cmd := exec.Command("git", "clone", repo)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", repo, err)
			continue
		}
	}

	fmt.Println("All repositories have been cloned successfully.")
	return nil
}

func SetSSHRemote() error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return errors.Wrap(err, "failed to fetch user repos")
	}

	for _, repo := range repos {
		fmt.Printf("Setting SSH remote for %s...\n", repo)
		cmd := exec.Command("git", "-C", repo, "remote", "set-url", "origin", fmt.Sprintf("git@github.com:%s/%s.git", "USERNAME", repo)) // USERNAME debe ser dinámico
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SSH remote for %s: %v\n", repo, err)
			continue
		}
	}

	fmt.Println("SSH remotes have been set successfully.")
	return nil
}

func ExecuteInRepos(command string) error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		fmt.Printf("Executing command in %s...\n", repo)
		repoPath := filepath.Join("path_to_repos", filepath.Base(repo)) // Asumir un path base para repositorios
		cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s", repoPath, command))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute command in %s: %v\n", repo, err)
			continue
		}
	}

	fmt.Println("Command executed in all repositories successfully.")
	return nil
}

func CloneOthersRepos(username string) error {
	repos, err := github.FetchUserRepos(&username)
	if err != nil {
		return errors.Wrap(err, "failed to fetch user repos")
	}

	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo)
		cmd := exec.Command("git", "clone", repo)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", repo, err)
			continue
		}
	}

	fmt.Println("All repositories have been cloned successfully.")
	return nil
}

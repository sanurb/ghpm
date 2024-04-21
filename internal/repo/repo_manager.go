package repo

import (
	"fmt"
	"github.com/sanurb/ghpm/internal/github"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func CloneRepos() error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return fmt.Errorf("failed to fetch user repos: %w", err)
	}

	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo.HTMLURL)
		cmd := exec.Command("git", "clone", repo.HTMLURL)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", repo.HTMLURL, err)
			continue
		}
	}

	return nil
}

func SetSSHRemote() error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return fmt.Errorf("failed to fetch user repos: %w", err)
	}

	for _, repo := range repos {
		repoName := filepath.Base(repo.HTMLURL)
		fmt.Printf("Setting SSH remote for %s...\n", repoName)
		repoPath := filepath.Join("path_to_repos", repoName)
		cmd := exec.Command("git", "-C", repoPath, "remote", "set-url", "origin", fmt.Sprintf("git@github.com:%s/%s.git", github.GetUsername(), repoName))
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SSH remote for %s: %v\n", repoName, err)
			continue
		}
	}

	fmt.Println("SSH remotes have been set successfully.")
	return nil
}

func ExecuteInRepos(command string) error {
	repos, err := github.FetchUserRepos(nil)
	if err != nil {
		return fmt.Errorf("failed to execute command in repositories: %w", err)
	}

	for _, repo := range repos {
		fmt.Printf("Executing command in %s...\n", repo)
		repoPath := filepath.Join("path_to_repos", filepath.Base(repo.HTMLURL))
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
		return fmt.Errorf("failed to fetch repos for user %s: %w", username, err)
	}

	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo)
		cmd := exec.Command("git", "clone", repo.HTMLURL)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone %s: %v\n", repo, err)
			continue
		}
	}

	fmt.Println("All repositories have been cloned successfully.")
	return nil
}

func FilterRecentRepos(repos []github.RepoInfo, days int) ([]github.RepoInfo, error) {
	var recentRepos []github.RepoInfo
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, repo := range repos {
		lastPushed, err := time.Parse(time.RFC3339, repo.PushedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse push time for repo %s: %w", repo.HTMLURL, err)
		}
		if lastPushed.After(cutoff) {
			recentRepos = append(recentRepos, repo)
		}
	}
	return recentRepos, nil
}

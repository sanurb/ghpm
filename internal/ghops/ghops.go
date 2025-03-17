package ghops

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sanurb/ghpm/internal/github"
)

// GHCliError indicates an error invoking the GitHub CLI.
type GHCliError struct {
	Cmd string
	Err error
}

func (e *GHCliError) Error() string {
	return fmt.Sprintf("GitHub CLI command failed (%s): %v", e.Cmd, e.Err)
}

func (e *GHCliError) Unwrap() error {
	return e.Err
}

// CloneRepo uses the GitHub CLI to clone a repository into "dest".
// It's a normal "git clone" behind the scenes (e.g. "gh repo clone").
func CloneRepo(url, dest string) error {
	if err := runGHCommand("repo", "clone", url, dest); err != nil {
		return fmt.Errorf("failed to clone repo %q: %w", url, err)
	}
	return nil
}

// ListSelfRepos returns the authenticated user's repositories.
// It's effectively: gh repo list --json "name,sshUrl" -L 500
func ListSelfRepos() ([]github.Repo, error) {
	out, err := execGHCommand("repo", "list", "--json", "name,sshUrl", "-L", "500")
	if err != nil {
		return nil, fmt.Errorf("failed to list user repos: %w", err)
	}
	return parseRepoListJSON(out)
}

// ListPublicRepos returns the public repositories for a given username.
// It's effectively: gh repo list <username> --public --json "name,sshUrl" -L 500
func ListPublicRepos(username string) ([]github.Repo, error) {
	if username == "" {
		return nil, fmt.Errorf("no username provided for listing public repos")
	}
	out, err := execGHCommand("repo", "list", username, "--public", "--json", "name,sshUrl", "-L", "500")
	if err != nil {
		return nil, fmt.Errorf("failed to list public repos for %q: %w", username, err)
	}
	return parseRepoListJSON(out)
}

func RunCommandInAllRepos(rootDir, customCmd string) error {
	if customCmd == "" {
		return fmt.Errorf("no custom command specified")
	}

	// A naive approach: find local repos by searching for .git folders.
	repos, err := discoverLocalRepos(rootDir)
	if err != nil {
		return fmt.Errorf("failed discovering repos: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("no git repos found under %s", rootDir)
	}

	for _, repoPath := range repos {
		fmt.Printf("\n[INFO] Running %q in repo: %s\n", customCmd, repoPath)
		// For example, run "sh -c <customCmd>" inside each repo directory
		if err := runCommandInDir(repoPath, customCmd); err != nil {
			// If an error in one repo is fatal, return. Or skip if you prefer partial success
			return fmt.Errorf("error running %q in %s: %w", customCmd, repoPath, err)
		}
	}

	return nil
}

func SetSSHRemote(rootDir, username string) error {
	if username == "" {
		return fmt.Errorf("no GitHub username specified")
	}

	repos, err := discoverLocalRepos(rootDir)
	if err != nil {
		return fmt.Errorf("failed discovering repos: %w", err)
	}
	if len(repos) == 0 {
		return fmt.Errorf("no git repos found under %s", rootDir)
	}

	for _, repoPath := range repos {
		fmt.Printf("\n[INFO] Setting SSH remote for %s\n", repoPath)
		// We try to guess the name of the repo from the existing remote url
		oldRemote, err := getRemoteURL(repoPath, "origin")
		if err != nil {
			return fmt.Errorf("failed to read remote url in %s: %w", repoPath, err)
		}
		// a naive parse from "https://github.com/owner/repo.git" or "git@github.com:owner/repo.git"
		owner, repoName := parseOwnerAndRepoFromURL(oldRemote)
		if owner == "" || repoName == "" {
			// fallback to the user if we can't parse
			owner = username
			repoName = filepath.Base(repoPath)
		}

		newURL := fmt.Sprintf("git@github.com:%s/%s.git", username, repoName)
		// run: git remote set-url origin <newURL>
		if err := runCommandInDir(repoPath, fmt.Sprintf("git remote set-url origin %s", newURL)); err != nil {
			return fmt.Errorf("failed setting ssh remote in %s: %w", repoPath, err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

// execGHCommand runs "gh" with the specified arguments and returns its stdout.
func execGHCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, &GHCliError{
			Cmd: fmt.Sprintf("gh %v", args),
			Err: err,
		}
	}
	return out, nil
}

func runGHCommand(args ...string) error {
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &GHCliError{
			Cmd: fmt.Sprintf("gh %v", args),
			Err: err,
		}
	}
	return nil
}

func parseRepoListJSON(in []byte) ([]github.Repo, error) {
	var repos []github.Repo
	if err := json.Unmarshal(in, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repo list JSON: %w", err)
	}
	return repos, nil
}

func runCommandInDir(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func discoverLocalRepos(root string) ([]string, error) {
	var repos []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // permission error, etc.
		}
		if !d.IsDir() {
			return nil
		}
		// e.g. path/.git
		gitPath := filepath.Join(path, ".git")
		info, e2 := os.Stat(gitPath)
		if e2 == nil && info.IsDir() {
			// This is a repo
			repos = append(repos, path)
			// skip walking deeper into this repo
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// getRemoteURL returns the remote URL for the specified remoteName (e.g. 'origin').
func getRemoteURL(repoPath, remoteName string) (string, error) {
	// e.g. "git remote get-url origin"
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func parseOwnerAndRepoFromURL(url string) (owner, repoName string) {
	// handle https://github.com/<owner>/<repo>.git or git@github.com:<owner>/<repo>.git
	url = strings.TrimSuffix(url, ".git")
	if idx := strings.Index(url, "github.com"); idx != -1 {
		after := url[idx+len("github.com"):]
		after = strings.TrimPrefix(after, ":")
		after = strings.TrimPrefix(after, "/")
		parts := strings.Split(after, "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	// fallback
	return "", ""
}

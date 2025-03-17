package git

import (
	"os/exec"
)

// CloneRepo clones a repository from the given URL into the destination directory.
func CloneRepo(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	return cmd.Run()
}

// BatchPushRepo pushes changes for the repository located at repoDir.
func BatchPushRepo(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "push")
	return cmd.Run()
}

// BatchPullRepo pulls updates for the repository located at repoDir.
func BatchPullRepo(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "pull")
	return cmd.Run()
}

// File: internal/ghops/orgs.go
package ghops

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/sanurb/ghpm/internal/github"
)

// ListUserOrgs returns the organizations that the authenticated user belongs to.
//
// We rely on "gh api user/orgs" rather than "gh org list", because the latter
// does not currently support JSON output. The endpoint "GET /user/orgs" lists
// organizations that the user is a member of.
//
// See: https://docs.github.com/en/rest/orgs/orgs#list-organizations-for-the-authenticated-user
func ListUserOrgs() ([]github.Org, error) {
	// "gh api user/orgs" defaults to GET, returns a JSON array of orgs.
	cmd := exec.Command("gh", "api", "user/orgs")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list user orgs: %w", err)
	}

	var orgs []github.Org
	if err := json.Unmarshal(out, &orgs); err != nil {
		return nil, fmt.Errorf("failed to parse orgs: %w", err)
	}
	return orgs, nil
}

// ListOrgRepos returns repositories belonging to a specific organization.
func ListOrgRepos(orgLogin string) ([]github.Repo, error) {
	cmd := exec.Command("gh", "repo", "list", orgLogin, "--json", "name,sshUrl", "-L", "500")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list repos for org '%s': %w", orgLogin, err)
	}

	var repos []github.Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse org repos: %w", err)
	}
	return repos, nil
}

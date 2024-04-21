package github

import (
	"encoding/json"
	"fmt"
	"github.com/sanurb/ghpm/internal/user"
	"net/http"
)

type RepoInfo struct {
	HTMLURL   string `json:"html_url"`
	UpdatedAt string `json:"updated_at"`
	PushedAt  string `json:"pushed_at"`
}

func FetchUserRepos(username *string) ([]RepoInfo, error) {
	var effectiveUsername string
	if username != nil {
		effectiveUsername = *username
	} else {
		effectiveUsername = GetUsername()
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", effectiveUsername)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()

	var repos []RepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repos, nil
}

func GetUsername() string {
	return user.GetUsername()
}

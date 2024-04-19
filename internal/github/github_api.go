package github

import (
	"encoding/json"
	"fmt"
	"github.com/sanurb/ghpm/internal/user"
	"net/http"
)

func FetchUserRepos(username *string) ([]string, error) {
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

	var repos []struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	repoURLs := make([]string, len(repos))
	for i, repo := range repos {
		repoURLs[i] = repo.HTMLURL
	}

	return repoURLs, nil
}

func GetUsername() string {
	return user.GetUsername()
}

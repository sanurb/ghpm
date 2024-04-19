package github

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sanurb/ghmp/internal/user"
	"net/http"
)

func FetchUserRepos(username *string) ([]string, error) {
	var effectiveUsername string
	if username != nil {
		effectiveUsername = *username
	} else {
		effectiveUsername = user.GetUsername()
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", effectiveUsername)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch repositories")
	}
	defer resp.Body.Close()

	var repos []struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	repoURLs := make([]string, len(repos))
	for i, repo := range repos {
		repoURLs[i] = repo.HTMLURL
	}

	return repoURLs, nil
}

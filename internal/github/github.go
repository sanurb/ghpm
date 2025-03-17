package github

// Repo represents a GitHub repository.
type Repo struct {
	Name   string `json:"name"`
	SSHUrl string `json:"sshUrl"`
}

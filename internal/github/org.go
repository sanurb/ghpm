package github

// Org represents a GitHub organization the user belongs to.
type Org struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

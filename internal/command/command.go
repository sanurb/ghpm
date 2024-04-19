package command

import "github.com/sanurb/ghpm/internal/repo"

type Command interface {
	Execute() error
}

type CloneCommand struct{}

func (c *CloneCommand) Execute() error {
	return repo.CloneRepos()
}

type SetSSHCommand struct{}

func (c *SetSSHCommand) Execute() error {
	return repo.SetSSHRemote()
}

type RunCommand struct {
	CustomCommand string
}

func (c *RunCommand) Execute() error {
	return repo.ExecuteInRepos(c.CustomCommand)
}

type CloneOthersCommand struct {
	Username string
}

func (c *CloneOthersCommand) Execute() error {
	return repo.CloneOthersRepos(c.Username)
}

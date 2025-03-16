package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/sanurb/ghpm/internal/ghops"
	"github.com/sanurb/ghpm/internal/github"
)

// We define a custom message to represent "one repo done cloning".
type clonedRepoMsg string

// We define a function that simulates the "clone" operation asynchronously.
// In real usage, you'd do something like "ghops.CloneRepo" in a goroutine.
func cloneRepoCmd(repoName string) tea.Cmd {
	// Simulate a random wait
	d := time.Millisecond * time.Duration(rand.Intn(1000)+300)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clonedRepoMsg(repoName)
	})
}

// We define a single "multiCloneAllCmd" that starts everything for multiple repos.
func multiCloneAllCmd(repos []string) tea.Cmd {
	// Start the first clone. The TuiModel will handle each clone in turn.
	if len(repos) == 0 {
		// Immediately done if no repos
		return func() tea.Msg { return clonedRepoMsg("") }
	}
	return cloneRepoCmd(repos[0])
}

// updateWelcome transitions to the menu on key press
func (m TuiModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		m.state = StateMenu
	}
	return m, nil
}

func (m TuiModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(m.menuOptions)-1 {
				m.menuCursor++
			}
		case "enter":
			selected := m.menuOptions[m.menuCursor]
			switch selected {
			case "Clone Own Repos":
				m.operation = "cloneOwn"
				m.state = StateRepoFetch
				return m, fetchReposCmd("self", "")

			case "Clone Public Repos":
				m.operation = "clonePublic"
				m.inputForm = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Enter GitHub Username").Key("username"),
					),
				)
				m.state = StateInput

			case "Clone Repos from an Org":
				m.operation = "cloneOrg"
				m.state = StateOrgFetch
				return m, fetchOrgsCmd()

			case "Run Command in All Repos":
				m.operation = "runCommand"
				m.inputForm = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Enter command").Key("command"),
					),
				)
				m.state = StateInput

			case "Set SSH Remote":
				m.operation = "setSSH"
				m.inputForm = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Enter GitHub Username").Key("username"),
					),
				)
				m.state = StateInput

			case "Exit":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m TuiModel) updateOrgFetch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	case orgsMsg:
		m.orgs = []github.Org(msg)
		opts := make([]huh.Option[string], 0, len(m.orgs))
		for _, o := range m.orgs {
			display := o.Login
			if o.Name != "" && o.Name != o.Login {
				display = fmt.Sprintf("%s (%s)", o.Name, o.Login)
			}
			opts = append(opts, huh.NewOption(display, o.Login))
		}
		sel := huh.NewSelect[string]().
			Title("Select an organization").
			Options(opts...).
			Key("selectedOrg")
		form := huh.NewForm(huh.NewGroup(sel))
		m.orgSelectForm = form
		m.state = StateOrgSelect
		return m, nil
	case errMsg:
		m.message = fmt.Sprintf("Error listing orgs: %v", msg.err)
		m.state = StateDone
		return m, nil
	}
	return m, nil
}

func (m TuiModel) updateOrgSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.orgSelectForm == nil {
		return m, nil
	}
	formModel, cmd := m.orgSelectForm.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		m.orgSelectForm = f
		if m.orgSelectForm.State == huh.StateCompleted {
			orgChoice := m.orgSelectForm.GetString("selectedOrg")
			m.selectedOrg = orgChoice
			m.state = StateRepoFetch
			return m, fetchOrgReposCmd(orgChoice)
		}
	}
	return m, cmd
}

func (m TuiModel) updateRepoFetch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	case reposMsg:
		m.repos = []github.Repo(msg)
		if len(m.repos) > 500 {
			m.repos = m.repos[:500]
		}
		items := make([]list.Item, 0, len(m.repos))
		for _, r := range m.repos {
			items = append(items, repoItem{name: r.Name, sshUrl: r.SSHUrl})
		}
		m.repoList.SetItems(items)
		m.repoList.Select(0)
		m.repoList.Paginator.PerPage = 10
		m.state = StateRepoList
		return m, nil
	case errMsg:
		m.message = fmt.Sprintf("Error fetching repos: %v", msg.err)
		m.state = StateDone
		return m, nil
	}
	return m, nil
}

func (m TuiModel) updateRepoList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var listCmd tea.Cmd
	m.repoList, listCmd = m.repoList.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.CloneAll) {
			// NEW: multiple clones with a progress bar
			m.downloadIndex = 0
			m.downloadTarget = len(m.repos)
			m.downloadRepos = make([]string, 0, m.downloadTarget)
			for _, r := range m.repos {
				m.downloadRepos = append(m.downloadRepos, r.Name)
			}
			m.done = false
			m.downloading = true
			m.state = StateDownloading

			// progress starts at 0
			cmd := m.progress.SetPercent(0.0)
			// start the first clone
			cloneCmd := multiCloneAllCmd(m.downloadRepos)
			return m, tea.Batch(cmd, cloneCmd)
		}
		if msg.String() == "enter" {
			if selected, ok := m.repoList.SelectedItem().(repoItem); ok {
				// single clone
				go ghops.CloneRepo(selected.sshUrl, selected.name)
				m.message = fmt.Sprintf("Cloning %s...", selected.name)
				m.state = StateDone
			}
		}
	}
	return m, listCmd
}

// NEW: updateDownloading handles the multi-clone spinner + progress
func (m TuiModel) updateDownloading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// let user abort
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd

	case clonedRepoMsg:
		// one repo done
		if m.downloadIndex >= m.downloadTarget-1 {
			// we're done
			m.done = true
			m.downloading = false
			return m, tea.Quit
		}
		m.downloadIndex++
		percent := float64(m.downloadIndex) / float64(m.downloadTarget)
		progressCmd := m.progress.SetPercent(percent)
		// start next clone
		nextRepo := m.downloadRepos[m.downloadIndex]
		cloneCmd := cloneRepoCmd(nextRepo)

		return m, tea.Batch(progressCmd, cloneCmd)

	case progress.FrameMsg:
		// let the progress bubble handle animation
		newProgress, cmd := m.progress.Update(msg)
		if p, ok := newProgress.(progress.Model); ok {
			m.progress = p
		}
		return m, cmd
	}

	return m, nil
}

func (m TuiModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.state = StateMenu
		m.message = ""
	}
	return m, nil
}

func (m TuiModel) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.inputForm, cmd = m.inputForm.Update(msg)

	type formIntf interface {
		GetString(key string) string
		State() huh.FormState
	}
	if f, ok := m.inputForm.(formIntf); ok && f.State() == huh.StateCompleted {
		username := f.GetString("username")
		cmdStr := f.GetString("command")

		if m.operation == "runCommand" {
			m.command = cmdStr
		}
		switch m.operation {
		case "clonePublic":
			m.state = StateRepoFetch
			return m, fetchReposCmd("public", username)
		case "setSSH":
			go ghops.SetSSHRemote(username)
			m.message = "SSH remote set for all repos."
			m.state = StateDone
		case "runCommand":
			go ghops.RunCommandInAllRepos(m.command)
			m.message = "Command executed in all repos."
			m.state = StateDone
		}
	}
	return m, cmd
}

// fetchOrgsCmd fetches the user’s orgs via “gh api user/orgs”.
func fetchOrgsCmd() tea.Cmd {
	return func() tea.Msg {
		orgs, err := ghops.ListUserOrgs()
		if err != nil {
			out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
			authMsg := strings.TrimSpace(string(out))
			return errMsg{
				fmt.Errorf("%w\nEnsure you're authenticated with gh CLI.\n%s", err, authMsg),
			}
		}
		return orgsMsg(orgs)
	}
}

// fetchOrgReposCmd fetches repos for a chosen org
func fetchOrgReposCmd(orgLogin string) tea.Cmd {
	return func() tea.Msg {
		repos, err := ghops.ListOrgRepos(orgLogin)
		if err != nil {
			out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
			authMsg := strings.TrimSpace(string(out))
			return errMsg{fmt.Errorf("%w\nEnsure you're authenticated.\n%s", err, authMsg)}
		}
		return reposMsg(repos)
	}
}

// fetchReposCmd handles “cloneOwn” or “clonePublic” repos.
func fetchReposCmd(mode, username string) tea.Cmd {
	return func() tea.Msg {
		repos, err := func() ([]github.Repo, error) {
			if mode == "self" {
				return ghops.ListSelfRepos()
			}
			return ghops.ListPublicRepos(username)
		}()
		if err != nil {
			out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
			authMsg := strings.TrimSpace(string(out))
			return errMsg{
				fmt.Errorf("%w\nEnsure you're authenticated with gh CLI.\n%s", err, authMsg),
			}
		}
		return reposMsg(repos)
	}
}

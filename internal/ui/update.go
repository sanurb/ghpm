package ui

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	zone "github.com/lrstanley/bubblezone"
	"github.com/sanurb/ghpm/internal/ghops"
	"github.com/sanurb/ghpm/internal/github"
)

type clonedRepoMsg string

func cloneRepoCmd(repoName string) tea.Cmd {
	d := time.Millisecond * time.Duration(rand.Intn(1000)+300)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clonedRepoMsg(repoName)
	})
}

func multiCloneAllCmd(repos []string) tea.Cmd {
	if len(repos) == 0 {
		return func() tea.Msg { return clonedRepoMsg("") }
	}
	return cloneRepoCmd(repos[0])
}

// =============== MENU ===============

func (m TuiModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft && msg.Button == tea.MouseButtonLeft &&
			msg.Action == tea.MouseActionRelease {
			// check all lines
			for i := range m.menuOptions {
				lineID := fmt.Sprintf("menu-%d", i)
				if zone.Get(lineID).InBounds(msg) {
					m.menuCursor = i
					return m.handleMenuChoice(i)
				}
			}
		}

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		switch msg.String() {
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < len(m.menuOptions)-1 {
				m.menuCursor++
			}
		case "enter":
			return m.handleMenuChoice(m.menuCursor)
		}
	}
	return m, nil
}

// The user’s choice
func (m TuiModel) handleMenuChoice(idx int) (tea.Model, tea.Cmd) {
	switch m.menuOptions[idx] {
	case "Clone Own Repos":
		m.operation = "cloneOwn"
		m.state = StateRepoFetch
		return m, tea.Batch(
			m.sp.Tick,                 // ensures the spinner is advanced
			fetchReposCmd("self", ""), // triggers the GH fetch
		)
	case "Clone Public Repos":
		m.operation = "clonePublic"
		// Validate that user typed a non-empty GitHub username:
		selectField := huh.NewInput().
			Title("Enter GitHub Username").
			Key("username").
			Validate(func(val string) error {
				if strings.TrimSpace(val) == "" {
					return fmt.Errorf("Username cannot be empty")
				}
				return nil
			})

		form := huh.NewForm(huh.NewGroup(selectField))
		m.inputForm = form
		initCmd := m.inputForm.Init()
		m.state = StateInput
		return m, initCmd

	case "Clone Repos from an Org":
		m.operation = "cloneOrg"
		m.state = StateOrgFetch
		return m, tea.Batch(
			m.sp.Tick, // ensures the spinner is advanced
			fetchOrgsCmd(),
		)

	case "Run Command in All Repos":
		m.operation = "runCommand"
		cmdField := huh.NewInput().
			Title("Enter command").
			Key("command").
			Validate(func(val string) error {
				if len(strings.TrimSpace(val)) == 0 {
					return fmt.Errorf("Command cannot be empty")
				}
				return nil
			})

		form := huh.NewForm(huh.NewGroup(cmdField))
		m.inputForm = form
		initCmd := m.inputForm.Init()
		m.state = StateInput
		return m, initCmd

	case "Set SSH Remote":
		m.operation = "setSSH"
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Enter GitHub Username").
				Key("username"),
		))
		m.inputForm = form
		initCmd := m.inputForm.Init()

		m.state = StateInput
		return m, initCmd

	case "Exit":
		return m, tea.Quit
	}
	return m, nil
}

// =============== ORG FETCH ===============
func (m TuiModel) updateOrgFetch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.orgSelectForm = huh.NewForm(huh.NewGroup(sel))
		initCmd := m.orgSelectForm.Init()

		m.state = StateOrgSelect
		return m, initCmd

	case errMsg:
		m.message = fmt.Sprintf("Error listing orgs: %v", msg.err)
		m.state = StateDone
		return m, nil
	}
	return m, nil
}

// =============== ORG SELECT ===============
func (m TuiModel) updateOrgSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.orgSelectForm == nil {
		return m, nil
	}
	formModel, formCmd := m.orgSelectForm.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		m.orgSelectForm = f
		if m.orgSelectForm.State == huh.StateCompleted {
			orgChoice := m.orgSelectForm.GetString("selectedOrg")
			m.selectedOrg = orgChoice
			m.state = StateRepoFetch
			return m, fetchOrgReposCmd(orgChoice)
		}
	}
	return m, formCmd
}

// =============== REPO FETCH ===============
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

		// Force stable page-size
		m.repoList.Paginator.PerPage = m.pageSize
		m.repoList.Paginator.SetTotalPages(len(items))

		m.state = StateRepoList
		return m, nil

	case errMsg:
		m.message = fmt.Sprintf("Error fetching repos: %v", msg.err)
		m.state = StateDone
		return m, nil
	}
	return m, nil
}

// =============== REPO LIST ===============
func (m TuiModel) updateRepoList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var listCmd tea.Cmd
	m.repoList, listCmd = m.repoList.Update(msg)

	// Force stable pagination
	m.repoList.Paginator.PerPage = m.pageSize
	cnt := len(m.repoList.Items())
	m.repoList.Paginator.SetTotalPages(cnt)
	if m.repoList.Paginator.Page >= m.repoList.Paginator.TotalPages {
		m.repoList.Paginator.Page = max(0, m.repoList.Paginator.TotalPages-1)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "?":
			m.showHelp = !m.showHelp
			m.repoList.SetShowHelp(m.showHelp)
			return m, nil
		case "enter":
			if selected, ok := m.repoList.SelectedItem().(repoItem); ok {
				go ghops.CloneRepo(selected.sshUrl, selected.name)
				m.message = fmt.Sprintf("Cloning %s...", selected.name)
				m.state = StateDone
			}
		}
		if key.Matches(msg, m.keys.CloneAll) {
			m.downloadIndex = 0
			m.downloadTarget = len(m.repos)
			m.downloadRepos = make([]string, 0, m.downloadTarget)
			for _, r := range m.repos {
				m.downloadRepos = append(m.downloadRepos, r.Name)
			}
			m.done = false
			m.downloading = true
			m.state = StateDownloading

			cmd := m.progress.SetPercent(0.0)
			cloneCmd := multiCloneAllCmd(m.downloadRepos)
			return m, tea.Batch(cmd, cloneCmd)
		}
	}

	return m, listCmd
}

// =============== DOWNLOADING ===============
func (m TuiModel) updateDownloading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	case clonedRepoMsg:
		if m.downloadIndex >= m.downloadTarget-1 {
			m.done = true
			m.downloading = false
			return m, tea.Quit
		}
		m.downloadIndex++
		percent := float64(m.downloadIndex) / float64(m.downloadTarget)
		progressCmd := m.progress.SetPercent(percent)
		nextRepo := m.downloadRepos[m.downloadIndex]
		cloneCmd := cloneRepoCmd(nextRepo)
		return m, tea.Batch(progressCmd, cloneCmd)
	case progress.FrameMsg:
		newProgress, cmd := m.progress.Update(msg)
		if np, ok := newProgress.(progress.Model); ok {
			m.progress = np
		}
		return m, cmd
	}
	return m, nil
}

// =============== DONE ===============
func (m TuiModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.state = StateMenu
		m.message = ""
	}
	return m, nil
}

// =============== INPUT (HUH FORM) ===============
func (m TuiModel) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 3) We intercept Ctrl+C or "q" here so we can quit:
	switch typed := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(typed, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	// Let the Huh form handle everything else
	var cmd tea.Cmd
	m.inputForm, cmd = m.inputForm.Update(msg)

	// If the Huh form is completed:
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
			// user typed a GitHub username -> fetch repos
			m.state = StateRepoFetch
			return m, tea.Batch(
				m.sp.Tick,
				fetchReposCmd("public", username),
			)
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

// fetch commands remain the same:

func fetchOrgsCmd() tea.Cmd {
	return func() tea.Msg {
		orgs, err := ghops.ListUserOrgs()
		if err != nil {
			out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
			authMsg := strings.TrimSpace(string(out))
			return errMsg{fmt.Errorf("%w\nEnsure you're authenticated.\n%s", err, authMsg)}
		}
		return orgsMsg(orgs)
	}
}

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
			return errMsg{fmt.Errorf("%w\nEnsure you're authenticated.\n%s", err, authMsg)}
		}
		return reposMsg(repos)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

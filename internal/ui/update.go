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
		return func() tea.Msg {
			return clonedRepoMsg("")
		}
	}
	return cloneRepoCmd(repos[0])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =============== MENU ===============
func (m TuiModel) updateMenu(msg tea.Msg) (TuiModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			m.showHelp = !m.showHelp
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

	case tea.MouseMsg:
		// If user clicks a line
		if msg.Type == tea.MouseLeft && msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			for i := range m.menuOptions {
				lineID := fmt.Sprintf("menu-%d", i)
				if zone.Get(lineID).InBounds(msg) {
					m.menuCursor = i
					return m.handleMenuChoice(i)
				}
			}
		}
	}
	return m, nil
}

func (m TuiModel) handleMenuChoice(idx int) (TuiModel, tea.Cmd) {
	switch m.menuOptions[idx] {

	case "Clone Own Repos":
		m.operation = "cloneOwn"
		m.state = StateRepoFetch
		return m, tea.Batch(
			m.sp.Tick,
			fetchReposCmd("self", ""),
		)

	case "Clone Public Repos":
		// This is the single blocking approach
		// So no new state; we do it right here
		usernamePtr := new(string)
		err := huh.NewInput().
			Title("Enter GitHub Username").
			Validate(func(v string) error {
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("Username cannot be empty")
				}
				return nil
			}).
			Value(usernamePtr).
			Run()
		if err != nil {
			m.message = "Input canceled or invalid"
			m.state = StateDone
			return m, nil
		}

		m.operation = "clonePublic"
		m.state = StateRepoFetch
		return m, tea.Batch(
			m.sp.Tick,
			fetchReposCmd("public", *usernamePtr),
		)

	case "Clone Repos from an Org":
		// Non-blocking form approach
		m.operation = "cloneOrg"
		m.state = StateOrgFetch
		return m, tea.Batch(
			m.sp.Tick,
			fetchOrgsCmd(),
		)

	case "Run Command in All Repos":
		// Also do a blocking input
		cmdPtr := new(string)
		err := huh.NewInput().
			Title("Enter command to run").
			Validate(func(v string) error {
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("Command cannot be empty")
				}
				return nil
			}).
			Value(cmdPtr).
			Run()
		if err != nil {
			m.message = "Command input canceled or invalid"
			m.state = StateDone
			return m, nil
		}

		// Fire off background
		go ghops.RunCommandInAllRepos(".", *cmdPtr)
		m.message = "Command executed in all repos."
		m.state = StateDone
		return m, nil

	case "Set SSH Remote":
		// Also do a blocking input
		usrPtr := new(string)
		err := huh.NewInput().
			Title("Enter GitHub Username").
			Validate(func(v string) error {
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("Username cannot be empty")
				}
				return nil
			}).
			Value(usrPtr).
			Run()
		if err != nil {
			m.message = "User input canceled or invalid"
			m.state = StateDone
			return m, nil
		}

		go ghops.SetSSHRemote(".", *usrPtr)
		m.message = "SSH remote set for all repos."
		m.state = StateDone
		return m, nil

	case "Exit":
		return m, tea.Quit
	}
	return m, nil
}

// =============== ORG FETCH ===============
func (m TuiModel) updateOrgFetch(msg tea.Msg) (TuiModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		newSpin, cmd := m.sp.Update(msg)
		m.sp = newSpin
		return m, cmd

	case orgsMsg:
		m.orgs = []github.Org(msg)
		// Build a single select form for the user to pick an org
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
		orgForm := huh.NewForm(huh.NewGroup(sel))
		initCmd := orgForm.Init()
		m.orgSelectForm = orgForm
		m.state = StateOrgSelect
		return m, initCmd

	case errMsg:
		m.message = fmt.Sprintf("Error listing orgs: %v", msg.err)
		m.state = StateDone
	}
	return m, nil
}

// =============== ORG SELECT ===============
func (m TuiModel) updateOrgSelect(msg tea.Msg) (TuiModel, tea.Cmd) {
	if m.orgSelectForm == nil {
		return m, nil
	}
	var cmd tea.Cmd
	formModel, cmd := m.orgSelectForm.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		m.orgSelectForm = f
		if f.State == huh.StateCompleted {
			orgChoice := f.GetString("selectedOrg")
			m.selectedOrg = orgChoice
			m.state = StateRepoFetch
			return m, fetchOrgReposCmd(orgChoice)
		}
	}
	return m, cmd
}

// =============== REPO FETCH ===============
func (m TuiModel) updateRepoFetch(msg tea.Msg) (TuiModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		newSpin, cmd := m.sp.Update(msg)
		m.sp = newSpin
		return m, cmd

	case reposMsg:
		m.repos = msg
		if len(m.repos) > 500 {
			m.repos = m.repos[:500]
		}
		items := make([]list.Item, 0, len(m.repos))
		for _, r := range m.repos {
			items = append(items, repoItem{name: r.Name, sshUrl: r.SSHUrl})
		}
		m.repoList.SetItems(items)
		m.repoList.Paginator.PerPage = m.pageSize
		m.repoList.Paginator.SetTotalPages(len(items))
		m.state = StateRepoList
	case errMsg:
		m.message = fmt.Sprintf("Error fetching repos: %v", msg.err)
		m.state = StateDone
	}
	return m, nil
}

// =============== REPO LIST ===============
func (m TuiModel) updateRepoList(msg tea.Msg) (TuiModel, tea.Cmd) {
	newList, listCmd := m.repoList.Update(msg)
	m.repoList = newList

	m.repoList.Paginator.PerPage = m.pageSize
	cnt := len(m.repoList.Items())
	m.repoList.Paginator.SetTotalPages(cnt)
	if m.repoList.Paginator.Page >= m.repoList.Paginator.TotalPages {
		m.repoList.Paginator.Page = max(0, m.repoList.Paginator.TotalPages-1)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			m.showHelp = !m.showHelp
			m.repoList.SetShowHelp(m.showHelp)
		case "enter":
			if sel, ok := m.repoList.SelectedItem().(repoItem); ok {
				go ghops.CloneRepo(sel.sshUrl, sel.name)
				m.message = fmt.Sprintf("Cloning %s...", sel.name)
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
			m.downloading = true
			m.done = false
			m.state = StateDownloading

			cmd := m.progress.SetPercent(0.0)
			cloneCmd := multiCloneAllCmd(m.downloadRepos)
			return m, tea.Batch(cmd, cloneCmd)
		}
	}
	return m, listCmd
}

// =============== DOWNLOADING ===============
func (m TuiModel) updateDownloading(msg tea.Msg) (TuiModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		newSpin, cmd := m.sp.Update(msg)
		m.sp = newSpin
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
		newProg, cmd := m.progress.Update(msg)
		if np, ok := newProg.(progress.Model); ok {
			m.progress = np
		}
		return m, cmd
	}
	return m, nil
}

// =============== DONE ===============
func (m TuiModel) updateDone(msg tea.Msg) (TuiModel, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.state = StateMenu
		m.message = ""
	}
	return m, nil
}

// =============== FETCH CMDS ===============
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
		var (
			repos []github.Repo
			err   error
		)
		if mode == "self" {
			repos, err = ghops.ListSelfRepos()
		} else {
			repos, err = ghops.ListPublicRepos(username)
		}
		if err != nil {
			out, _ := exec.Command("gh", "auth", "status").CombinedOutput()
			authMsg := strings.TrimSpace(string(out))
			return errMsg{fmt.Errorf("%w\nEnsure you're authenticated.\n%s", err, authMsg)}
		}
		return reposMsg(repos)
	}
}

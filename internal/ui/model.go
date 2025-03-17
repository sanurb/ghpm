package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	zone "github.com/lrstanley/bubblezone"

	"github.com/sanurb/ghpm/internal/github"
)

// States
const (
	StateMenu = iota
	StateOrgFetch
	StateOrgSelect
	StateRepoFetch
	StateRepoList
	StateDone
	StateInput
	StateDownloading
)

// Custom messages
type (
	reposMsg []github.Repo
	orgsMsg  []github.Org
	errMsg   struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

type repoItem struct {
	name   string
	sshUrl string
}

func (r repoItem) Title() string       { return r.name }
func (r repoItem) Description() string { return r.sshUrl }
func (r repoItem) FilterValue() string { return r.name }

type keyMap struct {
	Quit     key.Binding
	Help     key.Binding
	CloneAll key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Help, k.CloneAll}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit, k.Help, k.CloneAll}}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		CloneAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "clone all repos"),
		),
	}
}

type TuiModel struct {
	state int

	sp spinner.Model

	menuOptions []string
	menuCursor  int

	repoList list.Model
	repos    []github.Repo

	operation   string
	inputForm   tea.Model
	inputResult string
	command     string
	message     string

	orgs          []github.Org
	orgSelectForm *huh.Form
	selectedOrg   string

	helpModel help.Model
	keys      keyMap

	progress       progress.Model
	downloading    bool
	downloadIndex  int
	downloadTarget int
	downloadRepos  []string
	done           bool

	width    int
	height   int
	showHelp bool
	pageSize int
}

// NewTuiModel sets up spinner, help, and the list with per-page
func NewTuiModel(perPage int) TuiModel {
	sp := spinner.New()
	sp.Style = DownloadSpinnerStyle

	menu := []string{
		"Clone Own Repos",
		"Clone Public Repos",
		"Clone Repos from an Org",
		"Run Command in All Repos",
		"Set SSH Remote",
		"Exit",
	}

	repoList := list.New(nil, list.NewDefaultDelegate(), 50, 10)
	repoList.Title = "Repositories"
	repoList.SetFilteringEnabled(true)
	repoList.SetShowHelp(true)
	repoList.SetShowStatusBar(false)
	repoList.SetShowPagination(true)
	repoList.Paginator.PerPage = perPage

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return TuiModel{
		state: StateMenu,

		sp:          sp,
		menuOptions: menu,
		repoList:    repoList,
		helpModel:   help.New(),
		keys:        defaultKeyMap(),
		progress:    p,
		pageSize:    perPage,
	}
}

// Init starts the spinner
func (m TuiModel) Init() tea.Cmd {
	return tea.Batch(m.sp.Tick)
}

// Update delegates to specialized update methods
func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMenu:
		return m.updateMenu(msg)
	case StateOrgFetch:
		return m.updateOrgFetch(msg)
	case StateOrgSelect:
		return m.updateOrgSelect(msg)
	case StateRepoFetch:
		return m.updateRepoFetch(msg)
	case StateRepoList:
		return m.updateRepoList(msg)
	case StateDownloading:
		return m.updateDownloading(msg)
	case StateDone:
		return m.updateDone(msg)
	case StateInput:
		return m.updateInput(msg)
	}
	return m, nil
}

func (m TuiModel) View() string {
	var out string

	switch m.state {
	case StateMenu:
		out = m.renderWelcomeAndMenu()
	case StateOrgFetch:
		out = fmt.Sprintf("Fetching organizations... %s", m.sp.View())
	case StateOrgSelect:
		if m.orgSelectForm == nil {
			out = "No org selection form yet!"
		} else {
			out = m.orgSelectForm.View()
		}
	case StateRepoFetch:
		out = fmt.Sprintf("Fetching repositories... %s", m.sp.View())
	case StateRepoList:
		out = m.repoList.View()
	case StateDownloading:
		out = m.renderDownloading()
	case StateDone:
		out = m.message + "\nPress any key to return to menu."
	case StateInput:
		out = m.inputForm.View()
	default:
		out = "(unknown state)"
	}

	// Remove zone markers
	return zone.Scan(out)
}

func (m TuiModel) renderWelcomeAndMenu() string {
	welcome := "Welcome to GHPM!\n\nManage GitHub repositories, clone your org repos,\n" +
		"run commands across all repos, and configure SSH remotes.\n"

	menu := "Select an option:\n\n"
	for i, option := range m.menuOptions {
		cursor := "  "
		if i == m.menuCursor {
			cursor = "> "
		}
		lineID := fmt.Sprintf("menu-%d", i)
		marked := zone.Mark(lineID, fmt.Sprintf("%s%s", cursor, option))
		menu += marked + "\n"
	}

	return WelcomeBoxStyle.Render(welcome) + "\n\n" + menu
}

func (m TuiModel) renderDownloading() string {
	if m.done {
		return DoneStyle.Render(fmt.Sprintf("Done! Cloned %d repos.\n", m.downloadTarget))
	}
	spin := m.sp.View() + " "
	bar := m.progress.View()
	var repoName string
	if m.downloadIndex < len(m.downloadRepos) {
		repoName = m.downloadRepos[m.downloadIndex]
	} else {
		repoName = ""
	}
	info := "Cloning " + CurrentRepoStyle.Render(repoName)
	return fmt.Sprintf("%s\n\n%s\n\n%s", spin+info, bar, "Press q/esc to exit")
}

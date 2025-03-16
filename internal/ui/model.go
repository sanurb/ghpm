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
	"github.com/sanurb/ghpm/internal/github"
)

// --- Application States ---
const (
	StateWelcome = iota
	StateMenu
	StateOrgFetch
	StateOrgSelect
	StateRepoFetch
	StateRepoList
	StateDone
	StateInput
	StateDownloading // NEW: for the multi-download progress bar
)

// --- Messages ---
type (
	reposMsg []github.Repo
	orgsMsg  []github.Org
	errMsg   struct{ err error }
)

func (e errMsg) Error() string { return e.err.Error() }

// We itemize each repository in the list.
type repoItem struct {
	name   string
	sshUrl string
}

func (r repoItem) Title() string       { return r.name }
func (r repoItem) Description() string { return r.sshUrl }
func (r repoItem) FilterValue() string { return r.name }

// Key bindings
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

// TuiModel is our Bubble Tea model.
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

	// NEW: for multi-repo download progress
	progress       progress.Model
	downloading    bool     // are we in the middle of a multi-clone?
	downloadIndex  int      // how many done so far
	downloadTarget int      // how many total
	downloadRepos  []string // store the repo names for display
	done           bool
}

// NewTuiModel initializes our TUI, starting with a welcome screen (optional).
func NewTuiModel() TuiModel {
	sp := spinner.New()
	sp.Style = DownloadSpinnerStyle // from styles.go

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
	repoList.SetShowHelp(true)
	repoList.SetShowStatusBar(false)
	repoList.SetFilteringEnabled(true)
	repoList.SetShowPagination(true)
	repoList.Paginator.PerPage = 10

	// NEW: Setup a progress bar for multi download
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return TuiModel{
		state:       StateWelcome,
		sp:          sp,
		menuOptions: menu,
		repoList:    repoList,
		helpModel:   help.New(),
		keys:        defaultKeyMap(),

		progress: p,
	}
}

func (m TuiModel) Init() tea.Cmd {
	// spinner tick, or do multiple if needed
	return tea.Batch(m.sp.Tick)
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateWelcome:
		return m.updateWelcome(msg)
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
	case StateDownloading: // NEW
		return m.updateDownloading(msg)
	case StateDone:
		return m.updateDone(msg)
	case StateInput:
		return m.updateInput(msg)
	default:
		return m, nil
	}
}

func (m TuiModel) View() string {
	switch m.state {
	case StateWelcome:
		return m.renderWelcome()
	case StateMenu:
		return m.renderMenu()
	case StateOrgFetch:
		return fmt.Sprintf("Fetching organizations... %s", m.sp.View())
	case StateOrgSelect:
		if m.orgSelectForm == nil {
			return "No org selection form yet!"
		}
		return m.orgSelectForm.View()
	case StateRepoFetch:
		return fmt.Sprintf("Fetching repositories... %s", m.sp.View())
	case StateRepoList:
		return m.repoList.View()
	case StateDownloading: // show spinner + progress
		return m.renderDownloading()
	case StateDone:
		return m.message + "\nPress any key to return to menu."
	case StateInput:
		return "Input:\n" + m.inputForm.View()
	default:
		return "(unknown state)"
	}
}

// A simple welcome screen
func (m TuiModel) renderWelcome() string {
	boxContent := "Welcome to GHPM!\n\n" +
		"Manage GitHub repositories, clone your org repos,\n" +
		"run commands across all repos, and configure SSH remotes.\n\n" +
		"Press any key to begin."

	return WelcomeBoxStyle.Render(boxContent)
}

// A basic menu rendering
func (m TuiModel) renderMenu() string {
	view := "Select an option:\n\n"
	for i, option := range m.menuOptions {
		cursor := "  "
		if i == m.menuCursor {
			cursor = "> "
		}
		view += fmt.Sprintf("%s%s\n", cursor, option)
	}
	return view
}

// NEW: display spinner + progress + "Currently Cloning: X"
func (m TuiModel) renderDownloading() string {
	if m.done {
		return DoneStyle.Render(fmt.Sprintf("Done! Cloned %d repos.\n", m.downloadTarget))
	}
	spin := m.sp.View() + " "
	// percent = m.downloadIndex / float64(m.downloadTarget)
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

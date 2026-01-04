package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#FAFAFA")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51CF66"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type Config struct {
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type RepoStatus struct {
	Name   string
	Path   string
	Branch string
	Dirty  bool
	Ahead  int
	Behind int
	Error  string
}

type model struct {
	repos       []Repository
	statuses    []RepoStatus
	cursor      int
	selected    map[int]struct{}
	loading     bool
	message     string
	messageType string // "success", "error", or ""
	pendingOps  int    // Track pending operations
}

type statusUpdateMsg struct {
	index  int
	status RepoStatus
}

type pullCompleteMsg struct {
	index int
	err   error
}

type fetchCompleteMsg struct {
	index int
	err   error
}

type clearMessageMsg struct{}

func initialModel(configPath string) (model, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return model{}, err
	}

	statuses := make([]RepoStatus, len(config.Repositories))
	for i, repo := range config.Repositories {
		statuses[i] = RepoStatus{
			Name:   repo.Name,
			Path:   repo.Path,
			Branch: "loading...",
		}
	}

	return model{
		repos:      config.Repositories,
		statuses:   statuses,
		selected:   make(map[int]struct{}),
		loading:    true,
		pendingOps: len(config.Repositories),
	}, nil
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths
	for i := range config.Repositories {
		if !filepath.IsAbs(config.Repositories[i].Path) {
			config.Repositories[i].Path, err = filepath.Abs(config.Repositories[i].Path)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve path for %s: %w", config.Repositories[i].Name, err)
			}
		}
	}

	return &config, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshAllStatuses(),
		tea.EnterAltScreen,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}
			return m, nil
		case " ":
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			return m, nil
		case "r", "R":
			// Refresh all
			m.loading = true
			m.message = ""
			m.pendingOps = len(m.repos)
			return m, m.refreshAllStatuses()
		case "p", "P":
			// Pull all
			if len(m.selected) == 0 {
				m.message = "No repositories selected. Use spacebar to select."
				m.messageType = "error"
				return m, nil
			}
			m.loading = true
			m.message = "Pulling repositories..."
			m.messageType = ""
			m.pendingOps = len(m.selected)
			return m, m.pullSelected()
		case "f", "F":
			// Fetch all
			if len(m.selected) == 0 {
				m.message = "No repositories selected. Use spacebar to select."
				m.messageType = "error"
				return m, nil
			}
			m.loading = true
			m.message = "Fetching repositories..."
			m.messageType = ""
			m.pendingOps = len(m.selected)
			return m, m.fetchSelected()
		case "a", "A":
			// Select all
			for i := range m.repos {
				m.selected[i] = struct{}{}
			}
			return m, nil
		case "d", "D":
			// Deselect all
			m.selected = make(map[int]struct{})
			return m, nil
		}

	case statusUpdateMsg:
		m.statuses[msg.index] = msg.status
		m.pendingOps--
		if m.pendingOps == 0 && m.loading {
			m.loading = false
			m.message = "Status refreshed"
			m.messageType = "success"
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return clearMessageMsg{}
			})
		}
		return m, nil

	case fetchCompleteMsg:
		m.pendingOps--
		if msg.err != nil {
			m.message = fmt.Sprintf("Error fetching %s: %v", m.repos[msg.index].Name, msg.err)
			m.messageType = "error"
		} else {
			m.message = fmt.Sprintf("Fetched %s", m.repos[msg.index].Name)
			m.messageType = "success"
		}
		// Refresh status after fetch
		refreshCmd := m.refreshStatus(msg.index)
		if m.pendingOps == 0 {
			m.loading = false
		}
		return m, refreshCmd

	case pullCompleteMsg:
		m.pendingOps--
		if msg.err != nil {
			m.message = fmt.Sprintf("Error pulling %s: %v", m.repos[msg.index].Name, msg.err)
			m.messageType = "error"
		} else {
			m.message = fmt.Sprintf("Pulled %s", m.repos[msg.index].Name)
			m.messageType = "success"
		}
		// Refresh status after pull
		refreshCmd := m.refreshStatus(msg.index)
		if m.pendingOps == 0 {
			m.loading = false
		}
		return m, refreshCmd

	case clearMessageMsg:
		m.message = ""
		m.messageType = ""
		return m, nil

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(" Rhiza Manager "))
	b.WriteString("\n\n")

	// Repository list
	for i, status := range m.statuses {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		selected := " "
		if _, ok := m.selected[i]; ok {
			selected = "✓"
		}

		style := normalStyle
		if m.cursor == i {
			style = selectedStyle
		}

		// Status line
		statusText := fmt.Sprintf("%s · %s · ↑%d ↓%d",
			status.Branch,
			map[bool]string{true: "dirty", false: "clean"}[status.Dirty],
			status.Ahead,
			status.Behind,
		)

		if status.Error != "" {
			statusText = errorStyle.Render("error: " + status.Error)
		}

		line := fmt.Sprintf("%s %s %s  %s",
			cursor,
			selected,
			style.Render(status.Name),
			statusStyle.Render(statusText),
		)

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Message
	if m.message != "" {
		var msgStyle lipgloss.Style
		switch m.messageType {
		case "error":
			msgStyle = errorStyle
		case "success":
			msgStyle = successStyle
		default:
			msgStyle = normalStyle
		}
		b.WriteString(msgStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	// Help
	help := helpStyle.Render(
		"↑/↓: navigate  space: select  r: refresh  p: pull  f: fetch  a: select all  d: deselect all  q: quit",
	)
	b.WriteString(help)

	if m.loading {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Loading..."))
	}

	return b.String()
}

func (m model) refreshAllStatuses() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.repos))
	for i := range m.repos {
		cmds[i] = m.refreshStatus(i)
	}
	return tea.Batch(cmds...)
}

func (m model) refreshStatus(index int) tea.Cmd {
	return func() tea.Msg {
		status := getRepoStatus(m.repos[index].Path)
		return statusUpdateMsg{index: index, status: status}
	}
}

func (m model) pullSelected() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.selected {
		cmds = append(cmds, m.pullRepo(i))
	}
	return tea.Batch(cmds...)
}

func (m model) pullRepo(index int) tea.Cmd {
	return func() tea.Msg {
		err := runGitCommandNoOutput(m.repos[index].Path, "git pull")
		return pullCompleteMsg{index: index, err: err}
	}
}

func (m model) fetchSelected() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.selected {
		cmds = append(cmds, m.fetchRepo(i))
	}
	return tea.Batch(cmds...)
}

func (m model) fetchRepo(index int) tea.Cmd {
	return func() tea.Msg {
		err := runGitCommandNoOutput(m.repos[index].Path, "git fetch")
		return fetchCompleteMsg{index: index, err: err}
	}
}

func getRepoStatus(repoPath string) RepoStatus {
	status := RepoStatus{
		Path: repoPath,
		Name: filepath.Base(repoPath),
	}

	// Get branch
	branch, err := runGitCommand(repoPath, "git branch --show-current")
	if err != nil {
		status.Error = err.Error()
		status.Branch = "unknown"
		return status
	}
	status.Branch = strings.TrimSpace(branch)
	if status.Branch == "" {
		status.Branch = "detached"
	}

	// Check if dirty
	dirtyOutput, err := runGitCommand(repoPath, "git status --porcelain")
	if err != nil {
		status.Error = err.Error()
		return status
	}
	status.Dirty = strings.TrimSpace(dirtyOutput) != ""

	// Get ahead/behind
	revList, err := runGitCommand(repoPath, "git rev-list --left-right --count HEAD...@{upstream} 2>/dev/null")
	if err != nil {
		// Upstream may not exist, that's okay
		status.Ahead = 0
		status.Behind = 0
	} else {
		parts := strings.Fields(revList)
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &status.Behind)
			fmt.Sscanf(parts[1], "%d", &status.Ahead)
		}
	}

	return status
}

func runGitCommand(repoPath, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func runGitCommandNoOutput(repoPath, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	m, err := initialModel(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

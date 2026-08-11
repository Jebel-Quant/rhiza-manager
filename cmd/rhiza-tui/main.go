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
	"gopkg.in/yaml.v3"
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

type TemplateInfo struct {
	URL           string
	Branch        string
	Behind        int    // Commits behind template
	TemplateError string // Error message if template check failed
}

type RepoStatus struct {
	Name          string
	Path          string
	Branch        string
	Dirty         bool
	Ahead         int
	Behind        int
	Error         string
	HasTemplate   bool
	Template      *TemplateInfo
	TemplateError string
}

type model struct {
	repos        []Repository
	statuses     []RepoStatus
	cursor       int
	selected     map[int]struct{}
	loading      bool
	message      string
	messageType  string        // "success", "error", or ""
	pendingOps   int           // Track pending operations
	commitPrompt *CommitPrompt // Active commit prompt
}

type CommitPrompt struct {
	repoIndex  int
	repoName   string
	gitStatus  string
	showPrompt bool
	selected   string // "current" or "pr"
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

type syncCompleteMsg struct {
	index int
	err   error
}

type refreshAfterSyncMsg struct {
	index int
}

type commitPromptMsg struct {
	index     int
	gitStatus string
}

type commitCompleteMsg struct {
	index      int
	err        error
	branchName string
	mode       string
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
		case "s", "S":
			// Sync/Materialize templates
			if len(m.selected) == 0 {
				m.message = "No repositories selected. Use spacebar to select."
				m.messageType = "error"
				return m, nil
			}
			m.loading = true
			m.message = "Syncing templates..."
			m.messageType = ""
			m.pendingOps = len(m.selected)
			return m, m.syncSelected()
		case "1":
			// Select option 1: commit on current branch
			if m.commitPrompt != nil {
				m.commitPrompt.selected = "current"
			}
			return m, nil
		case "2":
			// Select option 2: create PR branch
			if m.commitPrompt != nil {
				m.commitPrompt.selected = "pr"
			}
			return m, nil
		case "enter":
			// Confirm selection
			if m.commitPrompt != nil {
				return m, m.commitChanges(m.commitPrompt.repoIndex, m.commitPrompt.selected)
			}
			return m, nil
		case "n", "N", "esc":
			// No/Escape - skip commit
			if m.commitPrompt != nil {
				repoIndex := m.commitPrompt.repoIndex
				m.commitPrompt = nil
				m.message = fmt.Sprintf("Skipped commit for %s", m.repos[repoIndex].Name)
				m.messageType = ""
				// Still refresh status
				return m, m.refreshStatus(repoIndex)
			}
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

	case syncCompleteMsg:
		m.pendingOps--
		if msg.err != nil {
			m.message = fmt.Sprintf("Error syncing %s: %v", m.repos[msg.index].Name, msg.err)
			m.messageType = "error"
			if m.pendingOps == 0 {
				m.loading = false
			}
			return m, nil
		} else {
			m.message = fmt.Sprintf("Synced %s", m.repos[msg.index].Name)
			m.messageType = "success"
		}
		// Check git status and prompt for commit
		if m.pendingOps == 0 {
			m.loading = false
		}
		// Get git status and show commit prompt
		return m, m.checkGitStatusForCommit(msg.index)

	case commitPromptMsg:
		// Show commit prompt with git status
		m.commitPrompt = &CommitPrompt{
			repoIndex:  msg.index,
			repoName:   m.repos[msg.index].Name,
			gitStatus:  msg.gitStatus,
			showPrompt: true,
		}
		return m, nil

	case commitCompleteMsg:
		// Commit completed
		if m.commitPrompt != nil && m.commitPrompt.repoIndex == msg.index {
			m.commitPrompt = nil
		}
		if msg.err != nil {
			m.message = fmt.Sprintf("Error committing %s: %v", m.repos[msg.index].Name, msg.err)
			m.messageType = "error"
		} else {
			if msg.mode == "pr" {
				m.message = fmt.Sprintf("Created branch %s and pushed %s. Create PR: gh pr create --title 'chore: rhiza manage sync' --body 'Sync with rhiza template'", msg.branchName, m.repos[msg.index].Name)
			} else {
				m.message = fmt.Sprintf("Committed and pushed %s", m.repos[msg.index].Name)
			}
			m.messageType = "success"
		}
		// Refresh status after commit
		return m, m.refreshStatus(msg.index)

	case refreshAfterSyncMsg:
		// Refresh status after sync completes
		return m, m.refreshStatus(msg.index)

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
		var statusParts []string
		statusParts = append(statusParts, status.Branch)
		statusParts = append(statusParts, map[bool]string{true: "dirty", false: "clean"}[status.Dirty])
		statusParts = append(statusParts, fmt.Sprintf("↑%d ↓%d", status.Ahead, status.Behind))

		// Add template status if available
		if status.HasTemplate && status.Template != nil {
			if status.Template.TemplateError != "" {
				statusParts = append(statusParts, fmt.Sprintf("template: error"))
			} else if status.Template.Behind > 0 {
				statusParts = append(statusParts, fmt.Sprintf("template: ↓%d", status.Template.Behind))
			} else {
				statusParts = append(statusParts, "template: up-to-date")
			}
		}

		statusText := strings.Join(statusParts, " · ")

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

	// Commit prompt
	if m.commitPrompt != nil && m.commitPrompt.showPrompt {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render(" Commit Changes? "))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render(fmt.Sprintf("Repository: %s", m.commitPrompt.repoName)))
		b.WriteString("\n\n")
		if m.commitPrompt.gitStatus != "" {
			b.WriteString(statusStyle.Render("Git status:"))
			b.WriteString("\n")
			// Show only first 20 lines of git status to avoid overwhelming the screen
			statusLines := strings.Split(m.commitPrompt.gitStatus, "\n")
			if len(statusLines) > 20 {
				statusLines = statusLines[:20]
				b.WriteString(strings.Join(statusLines, "\n"))
				b.WriteString("\n... (truncated)")
			} else {
				b.WriteString(m.commitPrompt.gitStatus)
			}
			b.WriteString("\n\n")
		}
		b.WriteString(normalStyle.Render("Commit message: "))
		b.WriteString(successStyle.Render("chore: rhiza manage sync"))
		b.WriteString("\n\n")

		// Show options
		option1Style := normalStyle
		option2Style := normalStyle
		if m.commitPrompt.selected == "current" {
			option1Style = selectedStyle
		} else {
			option2Style = selectedStyle
		}

		b.WriteString(option1Style.Render("  [1] Commit on current branch"))
		b.WriteString("\n")
		b.WriteString(option2Style.Render("  [2] Create PR from new branch"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 1/2 to select, Enter to confirm, n/esc to skip"))
		b.WriteString("\n\n")
	}

	// Message
	if m.message != "" && (m.commitPrompt == nil || !m.commitPrompt.showPrompt) {
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
		"↑/↓: navigate  space: select  r: refresh  p: pull  f: fetch  s: sync  a: select all  d: deselect all  q: quit",
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
		// If this repo has a template, fetch the template remote first to get accurate status
		templatePath := filepath.Join(m.repos[index].Path, ".rhiza", "template.yml")
		if _, err := os.Stat(templatePath); err == nil {
			// Template exists, fetch template remote for accurate status
			_ = runGitCommandNoOutput(m.repos[index].Path, "git fetch template 2>/dev/null || git fetch origin 2>/dev/null || true")
		}
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

func (m model) syncSelected() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.selected {
		cmds = append(cmds, m.syncRepo(i))
	}
	return tea.Batch(cmds...)
}

func (m model) syncRepo(index int) tea.Cmd {
	return func() tea.Msg {
		// Check if rhiza CLI is available
		rhizaCmd := "rhiza"
		if _, err := exec.LookPath("rhiza"); err != nil {
			// Try uvx as fallback
			if _, err := exec.LookPath("uvx"); err == nil {
				rhizaCmd = "uvx rhiza"
			} else {
				return syncCompleteMsg{index: index, err: fmt.Errorf("rhiza CLI not found. Install with: pip install rhiza or use uvx")}
			}
		}

		// Run rhiza materialize --force in the repository
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s materialize --force", rhizaCmd))
		cmd.Dir = m.repos[index].Path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return syncCompleteMsg{index: index, err: fmt.Errorf("%s", strings.TrimSpace(string(output)))}
		}
		return syncCompleteMsg{index: index, err: nil}
	}
}

func (m model) checkGitStatusForCommit(index int) tea.Cmd {
	return func() tea.Msg {
		// Wait a moment for git state to update
		time.Sleep(500 * time.Millisecond)

		// Get git status
		statusOutput, err := runGitCommand(m.repos[index].Path, "git status --short")
		if err != nil {
			// If git status fails, just refresh without prompting
			return refreshAfterSyncMsg{index: index}
		}

		statusStr := strings.TrimSpace(statusOutput)
		if statusStr == "" {
			// No changes, just refresh
			return refreshAfterSyncMsg{index: index}
		}

		// Get full git status for display
		fullStatus, _ := runGitCommand(m.repos[index].Path, "git status")

		return commitPromptMsg{index: index, gitStatus: fullStatus}
	}
}

func (m model) commitChanges(index int, mode string) tea.Cmd {
	return func() tea.Msg {
		repoPath := m.repos[index].Path

		// Get current branch name
		currentBranch, err := runGitCommand(repoPath, "git branch --show-current")
		if err != nil {
			return commitCompleteMsg{index: index, err: fmt.Errorf("failed to get current branch: %v", err), mode: mode}
		}
		currentBranch = strings.TrimSpace(currentBranch)

		branchName := currentBranch

		if mode == "pr" {
			// Create a new branch for PR
			// Use timestamp-based branch name: rhiza-sync-YYYYMMDD-HHMMSS
			timestamp := time.Now().Format("20060102-150405")
			branchName = fmt.Sprintf("rhiza-sync-%s", timestamp)

			// Create and checkout new branch
			if err := runGitCommandNoOutput(repoPath, fmt.Sprintf("git checkout -b %s", branchName)); err != nil {
				return commitCompleteMsg{index: index, err: fmt.Errorf("failed to create branch: %v", err), mode: mode}
			}
		}

		// Add all changes
		if err := runGitCommandNoOutput(repoPath, "git add ."); err != nil {
			return commitCompleteMsg{index: index, err: fmt.Errorf("git add failed: %v", err), mode: mode}
		}

		// Commit with message
		commitMsg := "chore: rhiza manage sync"
		if err := runGitCommandNoOutput(repoPath, fmt.Sprintf("git commit -m %q", commitMsg)); err != nil {
			return commitCompleteMsg{index: index, err: fmt.Errorf("git commit failed: %v", err), mode: mode}
		}

		// Push
		if mode == "pr" {
			// Push new branch and set upstream
			if err := runGitCommandNoOutput(repoPath, fmt.Sprintf("git push -u origin %s", branchName)); err != nil {
				return commitCompleteMsg{index: index, err: fmt.Errorf("git push failed: %v", err), mode: mode}
			}
		} else {
			// Push current branch
			if err := runGitCommandNoOutput(repoPath, fmt.Sprintf("git push origin %s", branchName)); err != nil {
				return commitCompleteMsg{index: index, err: fmt.Errorf("git push failed: %v", err), mode: mode}
			}
		}

		return commitCompleteMsg{index: index, err: nil, branchName: branchName, mode: mode}
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

	// Check for rhiza template
	templateInfo := checkRhizaTemplate(repoPath)
	status.HasTemplate = templateInfo != nil
	status.Template = templateInfo

	return status
}

type TemplateYAML struct {
	TemplateRepository string `yaml:"template-repository"`
	TemplateBranch     string `yaml:"template-branch"`
}

func checkRhizaTemplate(repoPath string) *TemplateInfo {
	templatePath := filepath.Join(repoPath, ".rhiza", "template.yml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return &TemplateInfo{URL: "error", TemplateError: err.Error()}
	}

	var templateYAML TemplateYAML
	if err := yaml.Unmarshal(data, &templateYAML); err != nil {
		return &TemplateInfo{URL: "error", TemplateError: err.Error()}
	}

	if templateYAML.TemplateRepository == "" {
		return nil
	}

	// Convert repository format (owner/repo) to full URL if needed
	templateURL := templateYAML.TemplateRepository
	if !strings.Contains(templateURL, "://") && !strings.Contains(templateURL, "@") {
		// Assume GitHub format, convert to HTTPS URL
		if !strings.Contains(templateURL, "/") {
			return &TemplateInfo{URL: "error", TemplateError: "invalid repository format"}
		}
		templateURL = fmt.Sprintf("https://github.com/%s.git", templateURL)
	}

	templateInfo := &TemplateInfo{
		URL:    templateURL,
		Branch: templateYAML.TemplateBranch,
	}

	// Default branch to main if not specified
	if templateInfo.Branch == "" {
		templateInfo.Branch = "main"
	}

	// Check how many commits behind the template
	// First, ensure we have the template remote fetched
	// Check if template remote exists
	remotes, err := runGitCommand(repoPath, "git remote")
	if err == nil {
		remoteList := strings.TrimSpace(remotes)
		hasTemplateRemote := strings.Contains(remoteList, "template") || strings.Contains(remoteList, "rhiza")

		if !hasTemplateRemote {
			// Try to add template as remote (but don't fail if it doesn't work)
			_ = runGitCommandNoOutput(repoPath, fmt.Sprintf("git remote add template %s 2>/dev/null", templateInfo.URL))
		}

		// Fetch template updates
		remoteName := "template"
		if !hasTemplateRemote {
			// Try to find existing remote that matches
			// Extract repo name from URL for comparison (e.g., "jebel-quant/rhiza" from "https://github.com/jebel-quant/rhiza.git")
			repoName := templateInfo.URL
			if strings.Contains(repoName, "github.com/") {
				parts := strings.Split(repoName, "github.com/")
				if len(parts) > 1 {
					repoName = strings.TrimSuffix(parts[1], ".git")
				}
			}

			for _, remote := range strings.Fields(remoteList) {
				remoteURL, _ := runGitCommand(repoPath, fmt.Sprintf("git remote get-url %s", remote))
				// Check if remote URL contains the repo name
				if strings.Contains(remoteURL, repoName) || strings.Contains(repoName, remoteURL) {
					remoteName = remote
					break
				}
			}
		}

		// Fetch the template
		_ = runGitCommandNoOutput(repoPath, fmt.Sprintf("git fetch %s %s 2>/dev/null", remoteName, templateInfo.Branch))

		// Check commits behind template
		// Compare current branch with template branch
		behindCmd := fmt.Sprintf("git rev-list --count HEAD..%s/%s 2>/dev/null", remoteName, templateInfo.Branch)
		behindOutput, err := runGitCommand(repoPath, behindCmd)
		if err == nil {
			var behind int
			if n, _ := fmt.Sscanf(strings.TrimSpace(behindOutput), "%d", &behind); n == 1 {
				templateInfo.Behind = behind
			}
		}
	}

	return templateInfo
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

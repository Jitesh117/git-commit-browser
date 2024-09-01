package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF7F50"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	titleStyle        = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#8f157b")).
				Padding(0, 1)
)

type commit struct {
	hash    string
	message string
}

func (c commit) Title() string       { return c.hash }
func (c commit) Description() string { return c.message }
func (c commit) FilterValue() string { return c.hash + " " + c.message }

type model struct {
	commits     list.Model
	searchInput textinput.Model
	err         error
	quitting    bool
	selectedMsg string
}

func initialModel() model {
	commits, err := getCommits()
	items := make([]list.Item, len(commits))
	for i, c := range commits {
		items[i] = c
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Git commit browser"
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	ti := textinput.New()
	ti.Placeholder = "Search commits..."
	ti.Focus()
	return model{
		commits:     l,
		searchInput: ti,
		err:         err,
	}
}

func getCommits() ([]commit, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%s")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	commits := make([]commit, len(lines))
	for i, line := range lines {
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			commits[i] = commit{hash: parts[0], message: parts[1]}
		}
	}
	return commits, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.searchInput.Value() != "" {
				return m, m.performSearch
			}
			if i, ok := m.commits.SelectedItem().(commit); ok {
				m.selectedMsg = fmt.Sprintf("Selected commit: %s", i.hash)
				return m, nil
			}
		case "C":
			if i, ok := m.commits.SelectedItem().(commit); ok {
				err := clipboard.WriteAll(i.hash)
				if err != nil {
					m.selectedMsg = fmt.Sprintf("Error copying to clipboard: %v", err)
				}
				//     else {
				// 	m.selectedMsg = fmt.Sprintf("Copied commit hash to clipboard: %s", i.hash)
				// }
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(0, 0).GetFrameSize()
		m.commits.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.commits, cmd = m.commits.Update(msg)
	return m, cmd
}

func (m model) performSearch() tea.Msg {
	query := strings.ToLower(m.searchInput.Value())
	var filteredItems []list.Item
	for _, item := range m.commits.Items() {
		if fuzzyMatch(item.(commit).FilterValue(), query) {
			filteredItems = append(filteredItems, item)
		}
	}
	m.commits.SetItems(filteredItems)
	if len(filteredItems) > 0 {
		m.commits.Select(0)
	}
	m.searchInput.SetValue("")
	return nil
}

func fuzzyMatch(s, query string) bool {
	s = strings.ToLower(s)
	queryRunes := []rune(query)
	queryIndex := 0
	for _, r := range s {
		if queryIndex >= len(queryRunes) {
			return true
		}
		if r == queryRunes[queryIndex] ||
			unicode.ToLower(r) == unicode.ToLower(queryRunes[queryIndex]) {
			queryIndex++
		}
	}
	return queryIndex >= len(queryRunes)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	// Create the base view with the title and commits list
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		m.searchInput.View(),
		m.commits.View(),
	)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#8f157b")).
		Padding(0, 1).
		Render("Git commit browser")

	return lipgloss.JoinVertical(lipgloss.Left, title, view)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

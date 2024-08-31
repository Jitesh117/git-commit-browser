package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#2196f3"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF7F50"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
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
	l.Title = "Git Commit Browser"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

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
		}
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.commits.SetSize(msg.Width-h, msg.Height-v-3)
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.commits, cmd = m.commits.Update(msg)
	return m, cmd
}

func (m *model) performSearch() tea.Msg {
	items := m.commits.Items()
	options := make([]string, len(items))
	for i, item := range items {
		options[i] = item.(commit).FilterValue()
	}

	pattern := algo.ParsePattern(m.searchInput.Value(), true, algo.CaseSmartCase, true)
	slab := util.MakeSlab(256)
	matched := algo.FuzzyMatchV2(false, false, true, options, pattern, true, algo.FuncMap{}, &slab)

	if len(matched) > 0 {
		m.commits.Select(matched[0].Index)
	}
	m.searchInput.SetValue("")
	return nil
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using Git Commit Browser!")
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.searchInput.View(),
		m.commits.View(),
		m.selectedMsg,
		"Press q to quit.",
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

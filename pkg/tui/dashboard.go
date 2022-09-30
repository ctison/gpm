package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardModel struct {
	tabs struct {
		activeIndex int
		titles      []string
	}
	views      []tea.Model
	windowSize tea.WindowSizeMsg
}

func NewDashboardModel() *DashboardModel {
	dm := &DashboardModel{}
	dm.AddTab("F1 Search", NewSearch())
	dm.AddTab("F2 Installed", NewSearch())
	return dm
}

func (this *DashboardModel) AddTab(title string, view tea.Model) {
	this.tabs.titles = append(this.tabs.titles, title)
	this.views = append(this.views, view)
}

func (this DashboardModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(this.views))
	for i := range this.views {
		cmds = append(cmds, this.views[i].Init())
	}
	return tea.Sequentially(tea.EnterAltScreen, tea.Batch(cmds...))
}
func (this DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return this, tea.Quit
		case tea.KeyF1:
			this.tabs.activeIndex = 0
			return this, nil
		case tea.KeyF2:
			this.tabs.activeIndex = 1
			return this, nil
		}
	case tea.WindowSizeMsg:
		this.windowSize = msg
		msg.Height -= lipgloss.Height(this.RenderTabs()) + 1
		for i := range this.views {
			this.views[i], _ = this.views[i].Update(msg)
		}
		return this, nil
	}
	cmds := make([]tea.Cmd, 0, len(this.views))
	for i := range this.views {
		var cmd tea.Cmd
		this.views[i], cmd = this.views[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return this, tea.Batch(cmds...)
}

var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	activeTab = tab.Copy().Border(activeTabBorder, true)

	tabGap = tab.Copy().
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
)

func (this DashboardModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, this.RenderTabs(), this.RenderView())
}

func (this DashboardModel) RenderTabs() string {
	s := make([]string, 0, len(this.tabs.titles))
	for i, tabText := range this.tabs.titles {
		if i == this.tabs.activeIndex {
			s = append(s, activeTab.Render(tabText))
		} else {
			s = append(s, tab.Render(tabText))
		}
	}
	row := lipgloss.JoinHorizontal(
		lipgloss.Left,
		s...,
	)
	gap := tabGap.Render(strings.Repeat(" ", max(0, this.windowSize.Width-lipgloss.Width(row)-2)))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
}

func (this DashboardModel) RenderView() string {
	return this.views[this.tabs.activeIndex].View()
}

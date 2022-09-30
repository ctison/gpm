package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v47/github"
)

type focus int

const (
	focusSearchQuery focus = iota
	focusContent
	focusUnset
)

type view int

const (
	viewNothing view = iota
	viewFetchingRepositories
	viewRepositories
	viewFetchingReleases
	viewReleases
	viewAssets
)

type SearchModel struct {
	windowSize tea.WindowSizeMsg
	focused    bool
	focus      focus
	view       view
	views      struct {
		searchQuery  textinput.Model
		spinner      spinner.Model
		repositories table.Model
		releases     table.Model
		assets       table.Model
	}
	data struct {
		repositories []*github.Repository
		releases     []*github.RepositoryRelease
	}
}

func NewSearch() *SearchModel {
	search := &SearchModel{
		focus: focusSearchQuery,
	}
	search.views.searchQuery = textinput.New()
	search.views.searchQuery.Focus()
	search.views.searchQuery.Placeholder = "Search Github Repositories"
	search.views.spinner = spinner.New(
		spinner.WithSpinner(spinner.Points),
	)
	search.views.repositories = table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "Stars", Width: 8},
			{Title: "Language", Width: 20},
		}),
		table.WithFocused(false),
	)
	search.views.releases = table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "Date RFC822", Width: 20},
		}),
		table.WithFocused(false),
	)
	search.views.assets = table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "Downloads", Width: 10},
			{Title: "Size", Width: 10},
			{Title: "ContentType", Width: 25},
		}),
	)
	return search
}

func (this SearchModel) Init() tea.Cmd {
	return tea.Batch(this.views.searchQuery.Focus(), this.views.spinner.Tick)
}

func (this SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.windowSize = msg
		this.ComputeViewsSize()
		return this, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if this.focus == focusSearchQuery && this.views.searchQuery.Value() != "" {
				cmd := this.SetView(focusContent, viewFetchingRepositories)
				return this, tea.Batch(
					queryRepositories(this.views.searchQuery.Value()),
					cmd,
				)
			}
			if this.focus == focusContent && this.view == viewRepositories && len(this.data.repositories) > 0 {
				cmd := this.SetView(focusContent, viewFetchingReleases)
				selectedRepository := this.data.repositories[this.views.repositories.Cursor()]
				return this, tea.Batch(
					queryReleases(selectedRepository.GetOwner().GetLogin(), selectedRepository.GetName()),
					cmd,
				)
			}
			if this.focus == focusContent && this.view == viewReleases && len(this.data.releases) > 0 {
				cmd := this.SetView(focusContent, viewAssets)
				selectedRelease := this.data.releases[this.views.repositories.Cursor()]
				rows := make([]table.Row, 0, len(selectedRelease.Assets))
				for _, asset := range selectedRelease.Assets {
					rows = append(rows, table.Row{
						asset.GetName(),
						fmt.Sprintf("%d", asset.GetDownloadCount()),
						ByteCountIEC(int64(asset.GetSize())),
						asset.GetContentType(),
					})
				}
				this.views.assets.SetRows(rows)
				return this, cmd
			}
		case tea.KeyEscape:
			var cmd tea.Cmd
			switch this.view {
			case viewRepositories, viewFetchingRepositories:
				cmd = this.SetFocus(focusSearchQuery)
			case viewFetchingReleases, viewReleases:
				cmd = this.SetView(focusContent, viewRepositories)
			case viewAssets:
				cmd = this.SetView(focusContent, viewReleases)
			}
			return this, cmd
		case tea.KeyTab:
			cmd := this.FocusNext()
			return this, cmd
		}
	case []*github.Repository:
		this.data.repositories = msg
		cmd := this.SetView(focusContent, viewRepositories)
		rows := make([]table.Row, 0, len(msg))
		for _, repository := range msg {
			rows = append(rows, table.Row{
				repository.GetFullName(),
				fmt.Sprintf("%d", repository.GetStargazersCount()),
				repository.GetLanguage(),
			})
		}
		this.views.repositories.SetRows(rows)
		this.views.repositories.GotoTop()
		return this, cmd
	case []*github.RepositoryRelease:
		this.data.releases = msg
		cmd := this.SetView(focusContent, viewReleases)
		rows := make([]table.Row, 0, len(msg))
		for _, release := range msg {
			rows = append(rows, table.Row{
				release.GetName(),
				release.GetPublishedAt().Local().Format(time.RFC822),
			})
		}
		this.views.releases.SetRows(rows)
		this.views.releases.GotoTop()
		return this, cmd
	}
	const viewsCount = 5
	cmds := make([]tea.Cmd, 0, viewsCount)
	var cmd tea.Cmd
	this.views.searchQuery, cmd = this.views.searchQuery.Update(msg)
	cmds = append(cmds, cmd)
	this.views.spinner, cmd = this.views.spinner.Update(msg)
	cmds = append(cmds, cmd)
	this.views.repositories, cmd = this.views.repositories.Update(msg)
	cmds = append(cmds, cmd)
	this.views.releases, cmd = this.views.releases.Update(msg)
	cmds = append(cmds, cmd)
	this.views.assets, cmd = this.views.assets.Update(msg)
	cmds = append(cmds, cmd)
	return this, tea.Batch(cmds...)
}

func (this *SearchModel) Focus() {
	this.focused = true
}

func (this *SearchModel) Blur() {
	this.focused = false
}

func (this *SearchModel) FocusNext() tea.Cmd {
	if this.focus+1 == focusContent && this.view == viewNothing {
		return nil
	}
	this.focus += 1
	if this.focus >= focusUnset {
		this.focus = 0
	}
	return this.SetFocus(this.focus)
}

func (this *SearchModel) SetFocus(newFocus focus) tea.Cmd {
	var cmds []tea.Cmd
	this.focus = newFocus
	if newFocus != focusSearchQuery {
		this.views.searchQuery.Blur()
	}
	if newFocus != focusContent {
		this.views.repositories.Blur()
		this.views.releases.Blur()
		this.views.assets.Blur()
	}
	switch newFocus {
	case focusSearchQuery:
		cmds = append(cmds, this.views.searchQuery.Focus())
	case focusContent:
		switch this.view {
		case viewRepositories:
			this.views.repositories.Focus()
		case viewReleases:
			this.views.releases.Focus()
		case viewAssets:
			this.views.assets.Focus()
		}
	}
	return tea.Batch(cmds...)
}

func (this *SearchModel) SetView(focus focus, view view) tea.Cmd {
	this.view = view
	return this.SetFocus(focus)
}

// https://pkg.go.dev/github.com/muesli/termenv#readme-color-chart
var (
	border = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	borderFocused = border.Copy().
			BorderForeground(lipgloss.Color("129"))
)

func (this *SearchModel) ComputeViewsSize() {
	borderWidth := lipgloss.Width(border.String())
	borderHeight := lipgloss.Height(border.String())
	searchInputHeight := lipgloss.Height(this.RenderSearchInput())
	contentHeight := this.windowSize.Height - borderHeight - searchInputHeight
	this.views.searchQuery.Width = this.windowSize.Width - borderWidth - lipgloss.Width(this.views.searchQuery.Prompt) - 1
	this.views.repositories.SetWidth(this.windowSize.Width - borderWidth)
	this.views.repositories.SetHeight(contentHeight)
	this.views.releases.SetWidth(this.windowSize.Width - borderWidth)
	this.views.releases.SetHeight(contentHeight)
	this.views.assets.SetWidth(this.windowSize.Width - borderWidth)
	this.views.assets.SetHeight(contentHeight)
}

func (this SearchModel) View() string {
	searchInput := this.RenderSearchInput()
	var main string
	switch this.view {
	case viewFetchingRepositories, viewFetchingReleases:
		main = this.views.spinner.View()
		if this.focus == focusContent {
			main = borderFocused.Copy().
				Align(lipgloss.Center, lipgloss.Center).
				Width(this.windowSize.Width - lipgloss.Width(main)).
				Height(this.windowSize.Height - lipgloss.Height(searchInput) - lipgloss.Height(main)).
				Render(main)
		} else {
			main = border.Copy().
				Align(lipgloss.Center, lipgloss.Center).
				Width(this.windowSize.Width - lipgloss.Width(main)).
				Height(this.windowSize.Height - lipgloss.Height(searchInput) - lipgloss.Height(main)).
				Render(main)
		}
	case viewRepositories:
		repositories := this.views.repositories.View()
		if this.focus == focusContent {
			main = borderFocused.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(repositories)).
				Render(repositories)
		} else {
			main = border.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(repositories)).
				Render(repositories)
		}
	case viewReleases:
		releases := this.views.releases.View()
		if this.focus == focusContent {
			main = borderFocused.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(releases)).
				Render(releases)
		} else {
			main = border.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(releases)).
				Render(releases)
		}
	case viewAssets:
		assets := this.views.assets.View()
		if this.focus == focusContent {
			main = borderFocused.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(assets)).
				Render(assets)
		} else {
			main = border.Copy().
				PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(assets)).
				Render(assets)
		}
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		searchInput,
		main,
	)
}

func (this SearchModel) RenderSearchInput() string {
	searchInput := this.views.searchQuery.View()
	if this.focus == focusSearchQuery {
		searchInput = borderFocused.Copy().
			PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(searchInput)).
			Render(searchInput)
	} else {
		searchInput = border.Copy().
			PaddingRight(this.windowSize.Width - 2 - lipgloss.Width(searchInput)).
			Render(searchInput)
	}
	return searchInput
}

func queryRepositories(query string) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient(nil)
		result, _, err := client.Search.Repositories(context.Background(), query, &github.SearchOptions{})
		if err != nil {
			return err
		}
		return result.Repositories
	}
}

func queryReleases(owner, repo string) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient(nil)
		result, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{})
		if err != nil {
			return err
		}
		return result
	}
}

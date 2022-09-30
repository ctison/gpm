package tui

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ctison/gpm/pkg/gpm"
)

type InstallModel struct {
	width, height int
	deps          []gpm.Dependency
	spinners      []spinner.Model
	progresses    []DownloadProgress
	done          []bool
	errors        []error
	totalDone     int
}

func NewInstallModel(gpm gpm.GPM, deps ...gpm.Dependency) InstallModel {
	im := InstallModel{
		deps:       deps,
		spinners:   make([]spinner.Model, 0, len(deps)),
		progresses: make([]DownloadProgress, 0, len(deps)),
		done:       make([]bool, len(deps)),
		errors:     make([]error, len(deps)),
	}
	for _, dep := range deps {
		im.spinners = append(im.spinners, spinner.New(
			spinner.WithSpinner(spinner.Dot),
		))
		im.progresses = append(im.progresses, NewDownloadProgress(gpm, dep))
	}
	return im
}

func (im InstallModel) Errored() bool {
	for _, err := range im.errors {
		if err != nil {
			return true
		}
	}
	return false
}

func (im InstallModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(im.spinners)+len(im.progresses))
	for _, spinner := range im.spinners {
		cmds = append(cmds, spinner.Tick)
	}
	for _, progress := range im.progresses {
		cmds = append(cmds, progress.Init())
	}
	return tea.Batch(cmds...)
}

func (im InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, len(im.spinners)+len(im.progresses))
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		im.width, im.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return im, tea.Quit
		}
	}
	for i := range im.spinners {
		var cmd tea.Cmd
		im.spinners[i], cmd = im.spinners[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	for i := range im.progresses {
		model, cmd := im.progresses[i].Update(msg)
		im.progresses[i] = model.(DownloadProgress)
		cmds = append(cmds, cmd)
		if im.progresses[i].Finished() {
			if err := im.progresses[i].Err(); err != nil {
				im.errors[i] = err
			}
			if !im.done[i] {
				im.totalDone++
			}
			im.done[i] = true
			if im.totalDone == len(im.done) {
				return im, tea.Quit
			}
		}
	}
	return im, tea.Batch(cmds...)
}

var (
	errorCross = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).SetString("x")
	checkMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).SetString("✓")
)

func (im InstallModel) View() string {
	var buf bytes.Buffer
	for i, dep := range im.deps {
		if im.errors[i] != nil {
			buf.WriteString(errorCross.String())
		} else if im.done[i] {
			buf.WriteString(checkMark.String())
		} else {
			buf.WriteString(im.spinners[i].View())
		}
		buf.WriteString(" " + dep.String())
		if !im.done[i] {
			buf.WriteString("   " + im.progresses[i].View())
		}
		buf.WriteString(fmt.Sprintln(""))
	}
	for _, progress := range im.progresses {
		if err := progress.Err(); err != nil {
			buf.WriteString(fmt.Sprintln("Error:", err.Error()))
		}
	}
	return buf.String()
}

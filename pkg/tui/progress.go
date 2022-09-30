package tui

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ctison/gpm/pkg/gpm"
)

// Internal ID management. Used during animating to ensure that frame messages
// are received only by spinner components that sent them.
var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

type DownloadProgress struct {
	gpm             gpm.GPM
	id              int
	reader          io.Reader
	Progress        progress.Model
	dep             gpm.Dependency
	c               chan ProgressMsg
	totalByteSize   int64
	currentByteSize int64
	err             error
	finished        bool
}

func NewDownloadProgress(gpm gpm.GPM, dep gpm.Dependency, opts ...progress.Option) DownloadProgress {
	return DownloadProgress{
		gpm:      gpm,
		id:       nextID(),
		dep:      dep,
		Progress: progress.New(opts...),
		c:        make(chan ProgressMsg),
	}
}

func (dp DownloadProgress) Init() tea.Cmd {
	go func(dp DownloadProgress) {
		if err := dp.gpm.InstallDependency(context.Background(), dp.dep, dp); err != nil {
			dp.c <- ProgressMsg{
				id:  dp.id,
				err: err,
			}
		} else {
			dp.c <- ProgressMsg{
				id:  dp.id,
				eof: true,
			}
		}
		close(dp.c)
	}(dp)
	return dp.ListenProgress
}

func (dp DownloadProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if prg, ok := msg.(ProgressMsg); ok {
		if dp.id != prg.id {
			return dp, nil
		}
		if prg.err != nil {
			dp.err = prg.err
			dp.finished = true
			return dp, nil
		}
		if prg.eof {
			if dp.currentByteSize != dp.totalByteSize {
				log.Printf("ERROR: %d != %d but eof == true\n", dp.currentByteSize, dp.totalByteSize)
			}
			dp.finished = true
			return dp, dp.Progress.SetPercent(100.)
		}
		if prg.currentSize != nil {
			dp.currentByteSize = *prg.currentSize
		}
		if prg.totalSize != nil {
			dp.totalByteSize = *prg.totalSize
		}
		if prg.readSize != nil {
			dp.currentByteSize += *prg.readSize
		}
		return dp, dp.ListenProgress
	}
	m, cmd := dp.Progress.Update(msg)
	dp.Progress = m.(progress.Model)
	return dp, cmd
}

func (dp DownloadProgress) View() string {
	if dp.err != nil {
		return fmt.Sprintf("Error: %s", dp.err.Error())
	}
	if dp.totalByteSize == 0 {
		return dp.Progress.ViewAs(0.)
	}
	return dp.Progress.ViewAs(float64(dp.currentByteSize) / (float64(dp.totalByteSize) / 100) / 100)
}

func (dp DownloadProgress) Err() error { return dp.err }

func (dp DownloadProgress) Finished() bool { return dp.finished }

type ProgressMsg struct {
	id                     int
	src                    *string
	currentSize, totalSize *int64
	readSize               *int64
	err                    error
	eof                    bool
}

func (dp DownloadProgress) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	log.Printf("TrackProgress: %s %d %d", src, currentSize, totalSize)
	dp.c <- ProgressMsg{
		id:          dp.id,
		src:         &src,
		currentSize: &currentSize,
		totalSize:   &totalSize,
	}
	dp.reader = stream
	return dp
}

func (dp DownloadProgress) ListenProgress() tea.Msg {
	prg, ok := <-dp.c
	if !ok {
		return nil
	}
	return prg
}

func (dp DownloadProgress) Read(p []byte) (n int, err error) {
	n, err = dp.reader.Read(p)
	nn := int64(n)
	dp.c <- ProgressMsg{
		id:       dp.id,
		readSize: &nn,
	}
	return n, err
}

func (dp DownloadProgress) Close() error {
	if closer, ok := dp.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

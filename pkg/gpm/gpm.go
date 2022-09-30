package gpm

import (
	"os"
	"path/filepath"
)

// GPM holds the common configurations to manage [Dependency].
type GPM struct {
	homePath  string
	storePath string
	binPath   string
}

func NewGPM(opts ...GPMOption) *GPM {
	gpm := &GPM{}

	// Apply all options to the program.
	for _, opt := range opts {
		opt(gpm)
	}

	return gpm
}

// GPMOption is used to set options on [GPM]. GPM can accept a variable number of options.
type GPMOption func(*GPM)

// WithHomePath sets the home path that is used when initializing default paths (eg. [GPM.GetBinPath]).
// Defaults to ~/.
func WithHomePath(homePath string) GPMOption {
	return func(gpm *GPM) {
		gpm.homePath = homePath
	}
}

// GetHomePath returns the user home directory. See [WithHomePath].
func (gpm GPM) GetHomePath() (string, error) {
	if gpm.homePath != "" {
		return gpm.homePath, nil
	}
	return os.UserHomeDir()
}

// WithStorePath sets the path prefix where the releases assets will be downloaded.
// Defaults to ~/.local/share/gpm.
func WithStorePath(storePath string) GPMOption {
	return func(gpm *GPM) {
		gpm.storePath = storePath
	}
}

// GetStorePath returns the store directory. See [WithStorePath].
func (gpm GPM) GetStorePath() (string, error) {
	if gpm.storePath != "" {
		return gpm.storePath, nil
	}
	homePath, err := gpm.GetHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(homePath, ".local", "share", "gpm"), nil
}

// WithBinPath sets the path where symlinks to executables will be created.
// Defaults to ~/.local/bin.
func WithBinPath(binPath string) GPMOption {
	return func(gpm *GPM) {
		gpm.binPath = binPath
	}
}

// GetBinPath returns bin directory. See [WithBinPath].
func (gpm GPM) GetBinPath() (string, error) {
	if gpm.binPath != "" {
		return gpm.binPath, nil
	}
	homePath, err := gpm.GetHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(homePath, ".local", "bin"), nil

}

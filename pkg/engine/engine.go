package engine

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cavaliercoder/grab"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

type Engine struct {
	log  *log.Logger
	path string
	gh   *github.Client
	grab *grab.Client
}

func New(log *log.Logger) (*Engine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find $HOME: %w", err)
	}
	gpmDir := filepath.Join(homeDir, "gpm")
	gpmDirBin := filepath.Join(gpmDir, "bin")
	if err := os.MkdirAll(gpmDirBin, 0755); err != nil {
		return nil, fmt.Errorf("failed to mkdir %q: %w", gpmDirBin, err)
	}
	return &Engine{
		log:  log,
		path: gpmDir,
		gh:   github.NewClient(nil),
		grab: grab.NewClient(),
	}, nil
}

// Stop engine.
func (ng *Engine) Stop() error {
	return nil
}

// List returns installed assets.
func (ng *Engine) List() ([]*Asset, error) {
	binPath := filepath.Join(ng.path, "bin")
	dir, err := os.Open(binPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", binPath, err)
	}
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %q: %w", binPath, err)
	}
	assets := make([]*Asset, 0, len(files))
	for _, f := range files {
		if f.Mode()&os.ModeSymlink != 0 {
			path, err := os.Readlink(filepath.Join(binPath, f.Name()))
			if err != nil {
				ng.log.Errorf("failed to read symlink %q: %v", f.Name(), err)
				continue
			}
			asset := NewAsset("", "", "", "", "")
			asset.Path = path
			path = strings.TrimPrefix(path, ng.path+string(filepath.Separator))
			a := strings.Split(path, string(filepath.Separator))
			if len(a) != 5 {
				ng.log.Errorf("failed to map symlink to asset: %q -> %v", asset.Path, a)
				continue
			}
			asset.Site = a[0]
			asset.Owner = a[1]
			asset.Repo = a[2]
			asset.Version = a[3]
			asset.Name = a[4]
			asset.LinkName = f.Name()
			assets = append(assets, asset)
		}
	}
	return assets, nil
}

// Install downloads, unpacks if needed, and links an asset.
func (ng *Engine) Install(asset *Asset) error {
	if err := ng.Download(asset); err != nil {
		return err
	}
	if err := ng.Unpack(asset); err != nil {
		return err
	}
	if err := ng.Link(asset, nil); err != nil {
		return err
	}
	return nil
}

func (ng *Engine) Download(asset *Asset) error {
	assetDir := filepath.Join(ng.path, "github.com", asset.Owner, asset.Repo, asset.Version)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return fmt.Errorf("failed to create dir %q: %w", assetDir, err)
	}
	ng.log.Debugf("mkdir %q", assetDir)
	assetPath := filepath.Join(assetDir, asset.Name)
	ghAsset, _, err := ng.gh.Repositories.GetReleaseAsset(context.Background(), asset.Owner, asset.Repo, asset.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch asset %q: %w", asset, err)
	}
	ng.log.Infof("Start download: %q", *ghAsset.BrowserDownloadURL)
	_, err = grab.Get(assetDir, *ghAsset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset %q: %w", asset, err)
	}
	ng.log.Infof("successfully downloaded %q", assetPath)
	asset.Path = assetPath
	return nil
}

func (ng *Engine) Unpack(asset *Asset) error {
	buffer := make([]byte, 512)
	f, err := os.Open(asset.Path)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", asset.Path, err)
	}
	if _, err = f.Read(buffer); err != nil {
		f.Close()
		return fmt.Errorf("failed to read %q: %w", asset.Path, err)
	}
	f.Close()
	contentType := http.DetectContentType(buffer)
	switch contentType {
	case "application/octet-stream":
		if err := os.Chmod(asset.Path, 0500); err != nil {
			return fmt.Errorf("failed to chmod %q: %w", asset.Path, err)
		}
	default:
		return fmt.Errorf("format not supported: %q: %q", contentType, asset.Path)
	}
	return nil
}

func (ng *Engine) Link(asset *Asset, name *string) error {
	if name == nil {
		name = &asset.Repo
	}
	symlinkPath := filepath.Join(ng.path, "bin", *name)
	_ = os.Remove(symlinkPath)
	if err := os.Symlink(asset.Path, symlinkPath); err != nil {
		return fmt.Errorf("failed to symlink %q to %q: %w", symlinkPath, asset.Path, err)
	}
	asset.LinkName = *name
	ng.log.Infof("symlinked %q -> %q", symlinkPath, asset.Path)
	return nil
}

func (ng *Engine) Prune() {}

func (ng *Engine) Upgrade() {}

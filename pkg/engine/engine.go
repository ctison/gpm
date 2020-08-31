package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/github"
)

type Engine struct {
	client *github.Client
}

type Release struct {
	// Local path where release is installed. Empty if it is not.
	Path string
	// Values of github.com/{{.User}}/{{.Repo}}/releases/{{.Version}}/
	User, Repo, Version string
	Assets              []Asset
}

func (release *Release) DownloadAsset(url string) (*Asset, error) {
	if url == "" {
		panic("asset must have an URL to be downloaded from")
	}
	if release.Path == "" {
		releasePath := fmt.Sprintf("/usr/local/gpm/github.com/%s/%s/%s", release.User, release.Repo, release.Version)
		if err := os.MkdirAll(releasePath, 0755); err != nil {
			return nil, err
		}
		release.Path = releasePath
	}
	asset := Asset{
		URL: url,
	}
	s := strings.Split(asset.URL, "/")
	asset.Name = s[len(s)-1]
	asset.Path = fmt.Sprintf("%s/%s", release.Path, asset.Name)
	if _, err := os.Stat(asset.Path); err == nil {
		log.Printf("%s already exists: skipping.", asset.Path)
		return &asset, nil
	}
	f, err := os.Create(asset.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("Start downloading %s", asset.Path)
	resp, err := http.Get(asset.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return nil, err
	}
	return &asset, nil
}

type Asset struct {
	Path string
	Name string
	URL  string
}

func (asset *Asset) Link(path string) error {
	os.Remove(path)
	if err := os.Symlink(asset.Path, path); err != nil {
		return err
	}
	if err := os.Chmod(asset.Path, 0555); err != nil {
		return err
	}
	return nil
}

func (release *Release) String() string {
	return fmt.Sprintf("%s/%s@%s", release.User, release.Repo, release.Version)
}

func NewRelease(location string) (*Release, error) {
	s := strings.Split(location, "@")
	if len(s) > 2 {
		return nil, fmt.Errorf("%q cannot contain more than one '@'", location)
	}
	var release Release
	if len(s) == 2 {
		release.Version = s[1]
	}
	s = strings.Split(s[0], "/")
	if len(s) > 2 {
		return nil, fmt.Errorf("%q cannot contain more than one '/'", location)
	}
	if len(s) == 1 {
		release.Repo = s[0]
	} else if len(s) == 2 {
		release.User = s[0]
		release.Repo = s[1]
	}
	return &release, nil
}

func NewEngine() (*Engine, error) {
	return &Engine{
		client: github.NewClient(nil),
	}, nil
}

func (engine *Engine) Install(release Release) (*Release, error) {
	var id int64
	if release.Version == "" || release.Version == "latest" {
		githubRelease, _, err := engine.client.Repositories.GetLatestRelease(context.Background(), release.User, release.Repo)
		if err != nil {
			return nil, err
		}
		id = githubRelease.GetID()
		release.Version = *githubRelease.TagName
	}
	if id != 0 {
		githubAssets, _, err := engine.client.Repositories.ListReleaseAssets(context.Background(), release.User, release.Repo, id, nil)
		if err != nil {
			return nil, err
		}
		re := regexp.MustCompile(fmt.Sprintf("(?i)%s", runtime.GOOS))
		for _, githubAsset := range githubAssets {
			if githubAsset.Name == nil || *githubAsset.Name == "" {
				continue
			}
			if !re.MatchString(*githubAsset.Name) {
				log.Printf("asset %q filtered out", *githubAsset.Name)
			} else {
				asset, err := release.DownloadAsset(*githubAsset.BrowserDownloadURL)
				if err != nil {
					return nil, err
				}
				if err := asset.Link(fmt.Sprintf("/usr/local/bin/%s", release.Repo)); err != nil {
					return nil, err
				}
				return &release, nil
			}
		}
	}
	return &release, nil
}

func (engine *Engine) List(pattern string) ([]Release, error) {
	panic("not implemented")
}

func (engine *Engine) Search(release Release) (local *Release, remote *Release, err error) {
	if release.Path == "" {
		local, err = engine.SearchLocal(release)
		if err != nil {
			return nil, nil, err
		}
	}
	remote, err = engine.SearchRemote(release)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (engine *Engine) SearchLocal(release Release) (*Release, error) {
	panic("not implemented")
}

func (engine *Engine) SearchRemote(release Release) (*Release, error) {
	panic("not implemented")
}

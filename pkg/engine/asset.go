package engine

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

type Asset struct {
	Name     string
	ID       int64
	Path     string
	LinkName string
	Release
}

func (a *Asset) String() string {
	var name string
	if a.Name != "" {
		name = ":" + a.Name
	}
	return a.Release.String() + name
}

type Release struct {
	Repository
	Version string
	ID      int64
}

func (r *Release) String() string {
	var version string
	if r.Version != "" {
		version = "@" + r.Version
	}
	return r.Repository.String() + version
}

type Repository struct {
	Site, Owner, Repo string
}

func (r *Repository) String() string {
	if r.Site == "" {
		r.Site = "github.com"
	}
	return fmt.Sprintf("%s/%s/%s", r.Site, r.Owner, r.Repo)
}

const Format = "[SITE://][OWNER/]REPO[@VERSION][:ASSET]"

var AssetReg = regexp.MustCompile(`^((.+)://)?(([^/]+)/)?([[:alnum:]]+)(@([^:]+))?(:(.+))?$`)
var ErrEmptyString = errors.New("string is empty")

func ParseAsset(s string) (*Asset, error) {
	if len(s) == 0 {
		return nil, ErrEmptyString
	}
	matches := AssetReg.FindAllStringSubmatch(s, -1)
	asset := Asset{
		Release: Release{
			Repository: Repository{
				Site:  matches[0][2],
				Owner: matches[0][4],
				Repo:  matches[0][5],
			},
			Version: matches[0][7],
		},
		Name: matches[0][9],
	}
	return &asset, nil
}

func NewAsset(site, owner, repo, version, name string) *Asset {
	return &Asset{
		Release: Release{
			Repository: Repository{
				Site:  site,
				Owner: owner,
				Repo:  repo,
			},
			Version: version,
		},
		Name: name,
	}
}

func (a *Asset) SetAssetFromGithubAsset(ghAsset *github.ReleaseAsset) error {
	if ghAsset == nil {
		return fmt.Errorf("github asset is nil from %q", a)
	}
	if ghAsset.Name == nil {
		return fmt.Errorf("unamed github asset from %q", a)
	}
	if ghAsset.ID == nil {
		return fmt.Errorf("no asset ID from %q", a)
	}
	a.Name = *ghAsset.Name
	a.ID = *ghAsset.ID
	return nil
}

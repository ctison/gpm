package engine

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/github"
)

// Resolve the owner of a repo.
func (ng *Engine) GuessOwner(asset *Asset) error {
	if asset.Owner != "" {
		return nil
	}
	result, _, err := ng.gh.Search.Repositories(context.Background(), asset.Repo, &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to guess owner for a repo named %q: %w", asset.Repo, err)
	}
	if len(result.Repositories) == 0 {
		return fmt.Errorf("no result for a repo named %q", asset.Repo)
	}
	asset.Owner = *result.Repositories[0].Owner.Login
	ng.log.Debugf(`guessed %q owner: "%s/%s"`, asset.Repo, asset.Owner, asset.Repo)
	return nil
}

// Guess the version of an asset.
func (ng *Engine) GuessVersion(asset *Asset) error {
	if asset.Release.ID != 0 {
		return nil
	}
	if asset.Version == "" || asset.Version == "highest" {
		releases, _, err := ng.gh.Repositories.ListReleases(context.Background(), asset.Owner, asset.Repo, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch releases from %q: %w", asset.Repository, err)
		}
		if len(releases) == 0 {
			return fmt.Errorf("no release found from %q", asset.Repository)
		}
		var highest *github.RepositoryRelease
		var highestVer semver.Version
		for j := range releases {
			if releases[j] != nil && releases[j].TagName != nil && releases[j].ID != nil {
				ver, err := semver.Make(strings.TrimPrefix(*releases[j].TagName, "v"))
				if err == nil {
					if highest == nil || ver.GT(highestVer) {
						highest = releases[j]
						highestVer = ver
					}
				} else {
					ng.log.Debugf("failed to parse release's version %q from %q: %v", *releases[j].TagName, asset.Repository, err)
				}
			}
		}
		if highest == nil {
			return fmt.Errorf("could not parse any versions from %q", asset.Repository)
		}
		asset.Version = *highest.TagName
		asset.Release.ID = *highest.ID
	} else if asset.Version == "latest" {
		release, _, err := ng.gh.Repositories.GetLatestRelease(context.Background(), asset.Owner, asset.Repo)
		if err != nil {
			return fmt.Errorf("failed to fetch latest release from %q: %w", asset.Repository, err)
		}
		if release != nil && release.ID != nil {
			asset.Release.ID = *release.ID
			if release.TagName != nil {
				asset.Version = *release.TagName
			}
		} else {
			return fmt.Errorf("failed to get release ID from %q", asset)
		}
	} else {
		release, _, err := ng.gh.Repositories.GetReleaseByTag(context.Background(), asset.Owner, asset.Repo, asset.Version)
		if err != nil {
			return fmt.Errorf("failed to fetch release %q: %w", asset, err)
		}
		if release != nil && release.ID != nil {
			asset.Release.ID = *release.ID
			if release.TagName != nil {
				asset.Version = *release.TagName
			}
		} else {
			return fmt.Errorf("failed to get release ID from %q", asset)
		}
	}
	return nil
}

// Guess the name of an asset for the current platform.
func (ng *Engine) GuessName(asset *Asset) error {
	if asset.Name != "" {
		return nil
	}
	assets, _, err := ng.gh.Repositories.ListReleaseAssets(context.Background(), asset.Owner, asset.Repo, asset.Release.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch release assets from %q: %w", asset, err)
	}
	if len(assets) == 0 {
		return fmt.Errorf("no asset found from %q", asset)
	}
	if len(assets) == 1 {
		return asset.SetAssetFromGithubAsset(assets[0])
	}
	filtered := make([]*github.ReleaseAsset, 0, len(assets))
	for i := range assets {
		if assets[i] == nil || assets[i].Name == nil {
			continue
		}
		name := *assets[i].Name
		if strings.Contains(name, runtime.GOOS) && strings.Contains(name, runtime.GOARCH) {
			filtered = append(filtered, assets[i])
		}
	}
	if len(filtered) == 0 {
		return fmt.Errorf("failed to resolve asset to install from %q", asset)
	}
	if len(filtered) == 1 {
		return asset.SetAssetFromGithubAsset(filtered[0])
	}
	ng.log.Debugf("filtered are %v", filtered)
	for i := range filtered {
		name := *filtered[i].Name
		if i := strings.LastIndex(name, "."); i < 0 || len(name)-i > 4 {
			return asset.SetAssetFromGithubAsset(filtered[i])
		}
	}
	return fmt.Errorf("failed to resolve asset to install from %q", asset)
}

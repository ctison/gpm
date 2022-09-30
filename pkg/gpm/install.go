package gpm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/v47/github"
	"github.com/hashicorp/go-getter/v2"
)

func (gpm GPM) InstallDependency(ctx context.Context, dep Dependency, progressTracker getter.ProgressTracker) error {
	gh := github.NewClient(nil)
	release, _, err := gh.Repositories.GetReleaseByTag(ctx, dep.Owner, dep.Repo, dep.ReleaseTag)
	if err != nil {
		return fmt.Errorf("failed to get release %s/%s@%s: %w", dep.Owner, dep.Repo, dep.ReleaseTag, err)
	}
	for _, asset := range release.Assets {
		if asset.GetName() == dep.AssetName {
			get := getter.Client{
				Getters: []getter.Getter{
					&getter.HttpGetter{
						DoNotCheckHeadFirst:   false,
						XTerraformGetDisabled: true,
					},
				},
				DisableSymlinks: true,
			}
			storePath, err := gpm.GetStorePath()
			if err != nil {
				return fmt.Errorf("failed to get gpm store path")
			}
			dst := filepath.Join(storePath, "github.com", dep.Owner, dep.Repo, dep.ReleaseTag, dep.AssetName)
			_, err = get.Get(ctx, &getter.Request{
				Src:              asset.GetBrowserDownloadURL(),
				Dst:              dst,
				Forced:           "http",
				GetMode:          getter.ModeAny,
				DisableSymlinks:  true,
				ProgressListener: progressTracker,
			})
			if err != nil {
				return fmt.Errorf("failed to download %q: %w", dep, err)
			}
			log.Printf("Asset downloaded to %q", dst)
			dirEntries, err := os.ReadDir(dst)
			if err != nil {
				return fmt.Errorf("failed to open %q as directory: %w", dst, err)
			}
			binPath, err := gpm.GetBinPath()
			if err != nil {
				return fmt.Errorf("failed to get bin path: %w", err)
			}
			r := regexp.MustCompile(`^[^.]*((\d+\.){2}\d+)?[^.]*$`)
			for _, dirEntry := range dirEntries {
				filePath := filepath.Join(dst, dirEntry.Name())
				fileInfo, err := dirEntry.Info()
				if err != nil {
					log.Printf("Failed to get file stat from %q: %s", filePath, err.Error())
					continue
				}
				if fileInfo.Mode().IsRegular() && r.MatchString(fileInfo.Name()) {
					if fileInfo.Mode()&0100 == 0 {
						if err := os.Chmod(filePath, fileInfo.Mode()|0100); err != nil {
							return fmt.Errorf("failed to chmod 500 %q: %w", filePath, err)
						}
					}
					symLinkPath := filepath.Join(binPath, dep.Repo)
					if _, err := os.Readlink(symLinkPath); err == nil {
						if err := os.Remove(symLinkPath); err != nil {
							return fmt.Errorf("failed to remove symlink %q: %w", symLinkPath, err)
						}
					}
					if err := os.Symlink(filePath, symLinkPath); err != nil {
						return fmt.Errorf("failed to symlink %q -> %q: %w", symLinkPath, filePath, err)
					}
					log.Printf("Symlinked %q -> %q", symLinkPath, filePath)
					return nil
				}
			}
			return nil
		}
	}
	return fmt.Errorf("asset named %q not found in release %s/%s@%s", dep.AssetName, dep.Owner, dep.Repo, dep.ReleaseTag)
}

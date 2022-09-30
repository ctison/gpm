package gpm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (gpm GPM) ListDownloadedDependencies(ctx context.Context) ([]Dependency, error) {
	storePath, err := gpm.GetStorePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get store path: %w", err)
	}

	dirEntries, err := os.ReadDir(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", storePath, err)
	}

	var deps []Dependency

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		domainName := dirEntry.Name()
		dirEntries, err := os.ReadDir(filepath.Join(storePath, domainName))
		if err != nil {
			continue
		}
		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() {
				continue
			}
			owner := dirEntry.Name()
			dirEntries, err := os.ReadDir(filepath.Join(storePath, domainName, owner))
			if err != nil {
				continue
			}
			for _, dirEntry := range dirEntries {
				if !dirEntry.IsDir() {
					continue
				}
				repo := dirEntry.Name()
				dirEntries, err := os.ReadDir(filepath.Join(filepath.Join(storePath, domainName, owner, repo)))
				if err != nil {
					continue
				}
				for _, dirEntry := range dirEntries {
					if !dirEntry.IsDir() {
						continue
					}
					releaseTag := dirEntry.Name()
					dirEntries, err := os.ReadDir(filepath.Join(storePath, domainName, owner, repo, releaseTag))
					if err != nil {
						continue
					}
					for _, dirEntry := range dirEntries {
						if !dirEntry.IsDir() {
							continue
						}
						assetName := dirEntry.Name()
						deps = append(deps, Dependency{
							Owner:      owner,
							Repo:       repo,
							ReleaseTag: releaseTag,
							AssetName:  assetName,
						})
					}
				}
			}
		}
	}

	return deps, nil
}

type LinkedDependencies struct {
	Src, Dst string
}

func (gpm GPM) ListLinkedDependencies(ctx context.Context) ([]LinkedDependencies, error) {
	binPath, err := gpm.GetBinPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get bin path: %w", err)
	}

	storePath, err := gpm.GetStorePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get store path: %w", err)
	}

	dirEntries, err := os.ReadDir(binPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %q: %w", binPath, err)
	}

	linkedDependencies := make([]LinkedDependencies, 0, len(dirEntries))

	for _, dirEntry := range dirEntries {
		if dirEntry.Type()&os.ModeSymlink == os.ModeSymlink {
			filePath := filepath.Join(binPath, dirEntry.Name())
			target, err := os.Readlink(filePath)
			if err != nil {
				log.Printf("failed to read link %q: %s", filePath, err.Error())
				continue
			}
			if strings.HasPrefix(target, storePath) {
				linkedDependencies = append(linkedDependencies, LinkedDependencies{
					Src: filePath,
					Dst: target,
				})
			}
		}
	}

	return linkedDependencies, nil
}

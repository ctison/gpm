package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/manifoldco/promptui"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

type search struct {
	interactive bool
}

// Function called by New to setup the search command.
func (cli *CLI) setupSearch() {
	cmd := newCommand()
	cmd.Use = "search QUERY"
	cmd.Short = "Search Github"
	cmd.Aliases = []string{"s"}
	cmd.RunE = cli.Search
	cmd.Flags().BoolVarP(&cli.search.interactive, "interactive", "i", false, "Interactive mode")
	cli.root.AddCommand(cmd)
}

func (cli *CLI) Search(cmd *cobra.Command, args []string) error {
	if cli.search.interactive {
		return cli.searchInteractive(cmd, args)
	}
	if len(args) != 1 {
		if err := cmd.Usage(); err != nil {
			log.Fatal("Error: ", err)
		}
		return nil
	}
	gh := github.NewClient(nil)
	result, _, err := gh.Search.Repositories(context.Background(), args[0], nil)
	if err != nil {
		log.Fatal("Error: search github failed: ", err)
	}
	for _, repo := range result.Repositories {
		fmt.Printf("%s %s %d⭐️\n", *repo.Name, *repo.FullName, repo.StargazersCount)
	}
	return nil
}

func (cli *CLI) searchInteractive(_ *cobra.Command, args []string) error {
	var query string
	var err error
	if len(args) == 0 {
		prompt := promptui.Prompt{
			Label: "Search",
		}
		query, err = prompt.Run()
		if err != nil {
			return err
		}
	} else {
		query = args[0]
	}
	gh := github.NewClient(nil)
	result, _, err := gh.Search.Repositories(context.Background(), query, &github.SearchOptions{
		Sort: "stars",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to search github: %w", err)
	}

	i, err := fuzzyfinder.Find(
		result.Repositories,
		func(i int) string {
			return *result.Repositories[i].FullName
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i > len(result.Repositories) {
				return ""
			}
			b, err := json.MarshalIndent(result.Repositories[i], "", "  ")
			if err != nil {
				return fmt.Sprint("Error: ", err)
			}
			return string(b)
		}),
	)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	return cli.searchReleaseInteractive(*result.Repositories[i].Owner.Login, *result.Repositories[i].Name)
}

func (cli *CLI) searchReleaseInteractive(owner, repo string) error {
	gh := github.NewClient(nil)
	results, _, err := gh.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{
		PerPage: 10,
	})
	if err != nil {
		return fmt.Errorf(`failed to fetch releases: "%s/%s": %w`, owner, repo, err)
	}
	type Release struct {
		Name string
		Size int
	}
	i, err := fuzzyfinder.Find(
		results,
		func(i int) string {
			return *results[i].Name
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			b, err := json.MarshalIndent(results[i], "", "  ")
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return string(b)
		}),
	)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	assets := results[i].Assets
	_, err = fuzzyfinder.Find(
		assets,
		func(j int) string {
			return *assets[i].Name
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			b, err := json.MarshalIndent(assets[i], "", "  ")
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return string(b)
		}),
	)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	return nil
}

package gpm

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Dependency represents a release asset from a Github repository.
type Dependency struct {
	Owner      string
	Repo       string
	ReleaseTag string
	AssetName  string
}

func (dep Dependency) String() string {
	var buf bytes.Buffer
	if dep.Owner != "" {
		buf.WriteString(dep.Owner + "/")
	}
	buf.WriteString(dep.Repo)
	if dep.ReleaseTag != "" {
		buf.WriteString("@" + dep.ReleaseTag)
	}
	if dep.AssetName != "" {
		buf.WriteString(":" + dep.AssetName)
	}
	return buf.String()
}

// RegexpDependency is used to parse dependencies from raw strings.
var RegexpDependency = regexp.MustCompile(`^((?P<owner>[^/]+)/)?(?P<repo>[a-zA-Z-_.]+)(@(?P<tag>[^:]*))?(:(?P<assets>.*))?$`)

// ConvertDependenciesStrings parses raw strings with [RegexpDependency].
func ConvertDependenciesStrings(s ...string) ([]Dependency, error) {
	// Accumulate dependencies in an array of size len(dependencies) but a dep string can target many concrete dependencies
	dependencies := make([]Dependency, 0, len(s))

	for _, dependency := range s {
		match := RegexpDependency.FindStringSubmatch(dependency)
		if match == nil {
			return nil, fmt.Errorf("'%s' not matching `%s`", dependency, RegexpDependency.String())
		}

		owner := match[RegexpDependency.SubexpIndex("owner")]
		repo := match[RegexpDependency.SubexpIndex("repo")]
		releaseTag := match[RegexpDependency.SubexpIndex("tag")]
		assetsNames := strings.Split(match[RegexpDependency.SubexpIndex("assets")], ",")

		for _, assetName := range assetsNames {
			dependencies = append(dependencies, Dependency{
				Owner:      owner,
				Repo:       repo,
				ReleaseTag: releaseTag,
				AssetName:  assetName,
			})
		}
	}

	return dependencies, nil
}

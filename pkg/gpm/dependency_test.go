package gpm

import "testing"

func TestDependency_String(t *testing.T) {
	type fields struct {
		Owner      string
		Repo       string
		ReleaseTag string
		AssetName  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Only Repo", fields{Repo: "gpm"}, "gpm"},
		{"Owner + Repo", fields{Owner: "gpm", Repo: "gpm"}, "gpm/gpm"},
		{"Repo" + "Version", fields{Repo: "gpm", ReleaseTag: "v42"}, "gpm@v42"},
		{"Repo + AssetName", fields{Repo: "gpm", AssetName: "gpm-linux-arm64"}, "gpm:gpm-linux-arm64"},
		{"Full", fields{Owner: "owner", Repo: "repo", ReleaseTag: "v1.0.0", AssetName: "asset-linux-arm64.tar.gz//exec"}, "owner/repo@v1.0.0:asset-linux-arm64.tar.gz//exec"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := Dependency{
				Owner:      tt.fields.Owner,
				Repo:       tt.fields.Repo,
				ReleaseTag: tt.fields.ReleaseTag,
				AssetName:  tt.fields.AssetName,
			}
			if got := dep.String(); got != tt.want {
				t.Errorf("Dependency.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

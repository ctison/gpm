package engine

import (
	"testing"
)

func TestParseAsset(t *testing.T) {
	tcs := []struct {
		S           string
		Asset       *Asset
		ShouldError bool
	}{
		{"xyz", NewAsset("", "", "xyz", "", ""), false},
		{"xxx/yyy", NewAsset("", "xxx", "yyy", "", ""), false},
		{"www://xxx/yyy", NewAsset("www", "xxx", "yyy", "", ""), false},
		{"www://yyy", NewAsset("www", "", "yyy", "", ""), false},
		{"xxx/yyy@zzz", NewAsset("", "xxx", "yyy", "zzz", ""), false},
		{"www://xxx/yyy@zzz:aaa", NewAsset("www", "xxx", "yyy", "zzz", "aaa"), false},
		{"www://xxx/yyy:aaa", NewAsset("www", "xxx", "yyy", "", "aaa"), false},
		{"", nil, true},
	}
	for _, tc := range tcs {
		asset, err := ParseAsset(tc.S)
		if tc.ShouldError {
			if err == nil {
				t.Errorf("%q should return error", tc.S)
			}
			continue
		}
		if *asset != *tc.Asset {
			t.Errorf("%#v differs from expected %#v", *asset, *tc.Asset)
		}
	}
}

package moddb

import "testing"

func TestNewestPrefersStableThenAnyWithAFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		releases []Release
		want     string
		ok       bool
	}{
		{"test build is the fallback", []Release{{ModVersion: "1.0.0"}, {ModVersion: "1.1.0-rc1", MainFile: "f.zip"}}, "1.1.0-rc1", true},
		{"stable wins over a newer test build", []Release{{ModVersion: "2.0.0-pre.1", MainFile: "b.zip"}, {ModVersion: "1.0.0", MainFile: "a.zip"}}, "1.0.0", true},
		{"no file anywhere", []Release{{ModVersion: "1.0.0"}, {ModVersion: "1.1.0-rc1"}}, "", false},
		{"no releases", nil, "", false},
	}
	for _, c := range cases {
		got, ok := Mod{Releases: c.releases}.Newest(false)
		if ok != c.ok || got.ModVersion != c.want {
			t.Errorf("%s: newest = %q, %v; want %q, %v", c.name, got.ModVersion, ok, c.want, c.ok)
		}
	}
}

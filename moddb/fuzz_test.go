package moddb

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// canned answers every request with body, so a fuzz target can feed any answer the site might send.
type canned string

func (c canned) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(string(c))),
		Header:     http.Header{"Content-Type": {"application/json"}},
		Request:    req,
	}, nil
}

// FuzzDecoders holds that no answer body from the Mod DB makes the listing, the page or the tags panic, and that a
// decoded listing's pages stay on the site.
func FuzzDecoders(f *testing.F) {
	f.Add([]byte(`{"mods":[{"modid":1,"assetid":2,"name":"A","tags":["x"]}]}`))
	f.Add([]byte(`{"statuscode":"200","mods":[
		{"modid":1,"assetid":11,"name":"A","tags":["Library"],"lastreleased":"2026-10-16 08:25:06"},
		{"modid":2,"assetid":22,"name":"B","urlalias":"bee","follows":7}]}`))
	f.Add([]byte(`{"mod":{"modid":1,"assetid":2,"releases":[{"releaseid":3,"modversion":"1.0.0","mainfile":"https://c/f.zip"}]}}`))
	f.Add([]byte(`{"mod":{"modid":890,"assetid":4405,"name":"Carry On","releases":[
		{"releaseid":3,"modversion":"2.0.0-rc.1","mainfile":"https://cdn/3.zip","filename":"3.zip"},
		{"releaseid":2,"modversion":"1.14.3","mainfile":"https://cdn/2.zip","filename":"2.zip"}]}}`))
	f.Add([]byte(`{"mod":null}`))
	f.Add([]byte(`{"tags":[{"tagid":"1","name":"Library"}]}`))
	f.Add([]byte(`{"tags":[{"tagid":186,"name":"Library"}]}`))
	f.Add([]byte(`{`))
	f.Fuzz(func(t *testing.T, body []byte) {
		c := New(Config{BaseURL: "http://moddb.test", HTTP: &http.Client{Transport: canned(body)}})
		mods, err := c.Mods(t.Context(), ListQuery{})
		if err == nil {
			if len(mods) > len(body)+1 {
				t.Fatalf("mods longer than body: %d > %d", len(mods), len(body))
			}
			for _, m := range mods {
				if !onSite(PageURL(m.URLAlias, m.AssetID)) {
					t.Fatalf("page escaped site: %q", PageURL(m.URLAlias, m.AssetID))
				}
			}
		}
		m, err := c.Mod(t.Context(), "1-2")
		if err == nil {
			_, _ = m.Newest(true)
		}
		tags, err := c.Tags(t.Context())
		if err == nil {
			for k := range tags {
				if k != strings.ToLower(k) {
					t.Fatalf("tag key %q not lower-case", k)
				}
			}
		}
	})
}

func onSite(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host == "mods.vintagestory.at"
}

// FuzzParseLink holds that an accepted install link names a mod id and a version of safe characters only.
func FuzzParseLink(f *testing.F) {
	f.Add("vintagestorymodinstall://carryon@1.14.3")
	f.Add("vintagestorymodinstall://carryon@0.13.0-pre.2/")
	f.Add("vintagestorymodinstall://carryon")
	f.Add("https://mods.vintagestory.at/carryon")
	f.Add("vintagestorymodinstall://../x@1")
	f.Fuzz(func(t *testing.T, raw string) {
		ref, err := ParseLink(raw)
		if err == nil && (!refModID.MatchString(ref.ModID) || !refVersion.MatchString(ref.Version)) {
			t.Fatalf("%q gave %+v", raw, ref)
		}
	})
}

// FuzzSplitID holds that an accepted id names a numeric mod, and that SplitID takes apart what ID joins.
func FuzzSplitID(f *testing.F) {
	f.Add("890-4405", 890, 4405)
	f.Add("123-456", 123, 456)
	f.Add("890", 890, 0)
	f.Add("123", 0, 0)
	f.Add("nope", 0, 0)
	f.Add("x-1", 7, 0)
	f.Fuzz(func(t *testing.T, id string, modID, assetID int) {
		if m, _, ok := SplitID(id); ok {
			if _, err := strconv.Atoi(m); err != nil {
				t.Fatalf("SplitID(%q) accepted mod id %q", id, m)
			}
		}
		if modID < 0 || assetID < 0 {
			return
		}
		m, a, ok := SplitID(ID(modID, assetID))
		if !ok || m != strconv.Itoa(modID) || a != strconv.Itoa(assetID) {
			t.Fatalf("SplitID(ID(%d, %d)) = %q, %q, %v", modID, assetID, m, a, ok)
		}
	})
}

// FuzzPageURL holds that an alias the site might send never leaves mods.vintagestory.at.
func FuzzPageURL(f *testing.F) {
	f.Add("carryon", 4405)
	f.Add("bee", 22)
	f.Add("", 4405)
	f.Add("../x", 1)
	f.Add("//evil.test", 1)
	f.Fuzz(func(t *testing.T, alias string, assetID int) {
		raw := PageURL(alias, assetID)
		if !onSite(raw) {
			t.Fatalf("PageURL(%q, %d) = %q escaped the site", alias, assetID, raw)
		}
	})
}

// FuzzPreRelease holds that only a hyphen in the version core, never build metadata, marks a test build.
func FuzzPreRelease(f *testing.F) {
	f.Add("1.14.3")
	f.Add("2.0.0-rc.1")
	f.Add("1.2-dev.3")
	f.Add("1.0.0+build-3")
	f.Add("1.0.0-rc.1+build")
	f.Fuzz(func(t *testing.T, version string) {
		core, _, _ := strings.Cut(version, "+")
		if PreRelease(version) != strings.Contains(core, "-") {
			t.Fatalf("PreRelease(%q) = %v", version, PreRelease(version))
		}
	})
}

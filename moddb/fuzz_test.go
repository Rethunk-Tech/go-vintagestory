package moddb

import (
	"io"
	"net/http"
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

// FuzzDecoders holds that no answer body from the Mod DB makes the listing, the page or the tags panic.
func FuzzDecoders(f *testing.F) {
	f.Add([]byte(`{"mods":[{"modid":1,"assetid":2,"name":"A","tags":["x"]}]}`))
	f.Add([]byte(`{"mod":{"modid":1,"assetid":2,"releases":[{"releaseid":3,"modversion":"1.0.0","mainfile":"https://c/f.zip"}]}}`))
	f.Add([]byte(`{"mod":null}`))
	f.Add([]byte(`{"tags":[{"tagid":"1","name":"Library"}]}`))
	f.Add([]byte(`{`))
	f.Fuzz(func(t *testing.T, body []byte) {
		c := New(Config{BaseURL: "http://moddb.test", HTTP: &http.Client{Transport: canned(body)}})
		mods, err := c.Mods(t.Context(), ListQuery{})
		if err == nil {
			for _, m := range mods {
				_ = PageURL(m.URLAlias, m.AssetID)
			}
		}
		m, err := c.Mod(t.Context(), "1-2")
		if err == nil {
			_, _ = m.Newest(true)
		}
		_, _ = c.Tags(t.Context())
	})
}

// FuzzParseLink holds that an accepted install link names a mod id and a version of safe characters only.
func FuzzParseLink(f *testing.F) {
	f.Add("vintagestorymodinstall://carryon@1.14.3")
	f.Add("vintagestorymodinstall://carryon@0.13.0-pre.2/")
	f.Add("vintagestorymodinstall://../x@1")
	f.Fuzz(func(t *testing.T, raw string) {
		ref, err := ParseLink(raw)
		if err == nil && (!refModID.MatchString(ref.ModID) || !refVersion.MatchString(ref.Version)) {
			t.Fatalf("%q gave %+v", raw, ref)
		}
	})
}

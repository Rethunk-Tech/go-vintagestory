package moddb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func fake(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mods":
			if r.URL.Query().Get("text") != "" && r.URL.Query().Get("orderby") != OrderDownloads {
				http.Error(w, "a search with text and no sort must order by downloads", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"statuscode":"200","mods":[
				{"modid":1,"assetid":11,"name":"A","tags":["Library"],"lastreleased":"2026-10-16 08:25:06"},
				{"modid":2,"assetid":22,"name":"B","urlalias":"bee","follows":7}]}`))
		case "/mod/890":
			// The site lists releases newest first, test builds included.
			_, _ = w.Write([]byte(`{"statuscode":"200","mod":{"modid":890,"assetid":4405,"name":"Carry On","releases":[
				{"releaseid":3,"modversion":"2.0.0-rc.1","mainfile":"https://cdn/3.zip","filename":"3.zip"},
				{"releaseid":2,"modversion":"1.14.3","mainfile":"https://cdn/2.zip","filename":"2.zip"}]}}`))
		case "/tags":
			_, _ = w.Write([]byte(`{"tags":[{"tagid":"1","name":"Library"},{"tagid":"2","name":"QoL"}]}`))
		default:
			_, _ = w.Write([]byte(`{"statuscode":"404"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL, HTTP: srv.Client()})
}

func TestModsListsInOneAnswerAndOrdersATextSearchByDownloads(t *testing.T) {
	c := fake(t)
	mods, err := c.Mods(t.Context(), ListQuery{Text: "bees", Tags: []string{"1"}})
	if err != nil || len(mods) != 2 || mods[0].LastReleased != "2026-10-16 08:25:06" || mods[1].URLAlias != "bee" {
		t.Fatalf("mods = %+v, %v", mods, err)
	}
}

func TestModReadsThePageAndItsNewestRelease(t *testing.T) {
	c := fake(t)
	m, err := c.Mod(t.Context(), "890")
	if err != nil || m.Name != "Carry On" || m.AssetID != 4405 {
		t.Fatalf("mod = %+v, %v", m, err)
	}
	if r, ok := m.Newest(false); !ok || r.ModVersion != "1.14.3" {
		t.Fatalf("newest stable = %+v, %v", r, ok)
	}
	if r, ok := m.Newest(true); !ok || r.ModVersion != "2.0.0-rc.1" {
		t.Fatalf("newest = %+v, %v", r, ok)
	}
	if _, err := c.Mod(t.Context(), "404"); err == nil {
		t.Fatal("a mod the site does not know read as one")
	}
}

func TestTagsAreKeyedByLowerCasedName(t *testing.T) {
	c := fake(t)
	tags, err := c.Tags(t.Context())
	if err != nil || tags["library"] != "1" || tags["qol"] != "2" {
		t.Fatalf("tags = %v, %v", tags, err)
	}
}

func TestIDsAndPagesAgreeWithTheSite(t *testing.T) {
	for _, tc := range []struct {
		id, modID, assetID string
		ok                 bool
	}{{"890-4405", "890", "4405", true}, {"890", "890", "", true}, {"nope", "", "", false}} {
		modID, assetID, ok := SplitID(tc.id)
		if modID != tc.modID || assetID != tc.assetID || ok != tc.ok {
			t.Errorf("SplitID(%q) = %q, %q, %v", tc.id, modID, assetID, ok)
		}
	}
	if ID(890, 4405) != "890-4405" {
		t.Errorf("ID = %q", ID(890, 4405))
	}
	if got := PageURL("carryon", 4405); got != SiteURL+"/carryon" {
		t.Errorf("alias page = %q", got)
	}
	if got := PageURL("", 4405); got != SiteURL+"/show/mod/4405" {
		t.Errorf("asset page = %q", got)
	}
}

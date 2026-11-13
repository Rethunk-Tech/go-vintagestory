// Package moddb is a client for the Vintage Story Mod DB's public API, https://mods.vintagestory.at/api. A mod is
// addressed two ways: its modid, which its modinfo.json names, and its assetid, which its page uses; neither answers to
// the other.
package moddb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	// BaseURL is the API root.
	BaseURL = "https://mods.vintagestory.at/api"
	// SiteURL is the site the API serves, where a mod's page lives.
	SiteURL = "https://mods.vintagestory.at"

	requestTimeout = 20 * time.Second
	// maxBody caps one API answer.
	maxBody = 16 << 20
)

// The site's orderby values.
const (
	OrderDownloads    = "downloads"
	OrderFollows      = "follows"
	OrderLastReleased = "lastreleased"
	OrderCreated      = "asset.created"
)

// Config is what a Client is built from.
type Config struct {
	// BaseURL replaces the API root. Empty means BaseURL.
	BaseURL string
	// HTTP sends every request; nil means http.DefaultClient. Redirects to another origin are refused.
	HTTP *http.Client
	// UserAgent is sent with every request when set.
	UserAgent string
}

// Client calls the Mod DB. It holds no state beyond its configuration, so it is safe for concurrent use.
type Client struct {
	base, agent string
	hc          *http.Client
}

// New builds a Client.
func New(cfg Config) *Client {
	return &Client{
		base:  strings.TrimRight(orDefault(cfg.BaseURL, BaseURL), "/"),
		agent: cfg.UserAgent,
		hc:    httpkit.WithRedirect(cfg.HTTP, httpkit.SameOrigin),
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// get sends one GET and decodes the 200 answer into out. A 429 is a *httpkit.RateLimitError and any other status a
// *httpkit.StatusError, so a caller can tell a missing mod from a failing site.
func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("vintage story mod db: %w", err)
	}
	if c.agent != "" {
		req.Header.Set("User-Agent", c.agent)
	}
	req.Header.Set("Accept", "application/json")
	raw, err := httpkit.DoCapped(c.hc, req, maxBody, nil)
	if err != nil {
		return fmt.Errorf("vintage story mod db: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("vintage story mod db: %w", err)
	}
	return nil
}

// ListQuery is one mods listing. Zero fields are not sent.
type ListQuery struct {
	// Text matches names, summaries and descriptions.
	Text string
	// OrderBy is one of the Order constants; empty keeps the site's order, except for a text search, which the site
	// orders by downloads.
	OrderBy string
	// Tags keeps the mods carrying any of these tag ids, as Tags hands them out.
	Tags []string
}

// ListedMod is one mod of a listing.
type ListedMod struct {
	ModID        int      `json:"modid"`
	AssetID      int      `json:"assetid"`
	Downloads    int      `json:"downloads"`
	Follows      int      `json:"follows"`
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	Author       string   `json:"author"`
	URLAlias     string   `json:"urlalias"`
	Logo         string   `json:"logo"`
	Tags         []string `json:"tags"`
	LastReleased string   `json:"lastreleased"` // the site's own "2006-01-02 15:04:05" UTC stamp
}

// Mods lists the site's mods in one answer; the API has no offset and returns the whole match at once, so paging is the
// caller's.
func (c *Client) Mods(ctx context.Context, q ListQuery) ([]ListedMod, error) {
	params := url.Values{}
	if q.Text != "" {
		params.Set("text", q.Text)
	}
	by := q.OrderBy
	if by == "" && q.Text != "" {
		// The site has no relevance order, and its own order puts a word found in any description ahead of the mod
		// named by the search; the most downloaded matches come first instead.
		by = OrderDownloads
	}
	if by != "" {
		params.Set("orderby", by)
		params.Set("orderdirection", "desc")
	}
	for _, t := range q.Tags {
		params.Add("tagids[]", t)
	}
	var parsed struct {
		Mods []ListedMod `json:"mods"`
	}
	if err := c.get(ctx, "/mods", params, &parsed); err != nil {
		return nil, err
	}
	if parsed.Mods == nil {
		parsed.Mods = []ListedMod{}
	}
	return parsed.Mods, nil
}

// Release is one uploaded build of a mod.
type Release struct {
	ReleaseID  int      `json:"releaseid"`
	MainFile   string   `json:"mainfile"`
	FileName   string   `json:"filename"`
	ModVersion string   `json:"modversion"`
	ModIDStr   string   `json:"modidstr"`
	Tags       []string `json:"tags"`
	Created    string   `json:"created"`
	Changelog  string   `json:"changelog"`
}

// Mod is one mod's page listing.
type Mod struct {
	ModID    int       `json:"modid"`
	AssetID  int       `json:"assetid"`
	Name     string    `json:"name"`
	Author   string    `json:"author"`
	URLAlias string    `json:"urlalias"`
	Logo     string    `json:"logofile"`
	Text     string    `json:"text"`
	Tags     []string  `json:"tags"`
	Releases []Release `json:"releases"`
}

// Newest is the newest release that can be downloaded. Without includePre it is the newest stable one, falling back to
// the newest test build when the mod has no stable release. The site lists releases newest first.
func (m Mod) Newest(includePre bool) (Release, bool) {
	if !includePre {
		for _, r := range m.Releases {
			if !PreRelease(r.ModVersion) && r.MainFile != "" {
				return r, true
			}
		}
	}
	for _, r := range m.Releases {
		if r.MainFile != "" {
			return r, true
		}
	}
	return Release{}, false
}

// Mod reads one mod by its modid, or by the mod id its modinfo.json names.
func (c *Client) Mod(ctx context.Context, ref string) (Mod, error) {
	var parsed struct {
		Mod *Mod `json:"mod"`
	}
	if err := c.get(ctx, "/mod/"+url.PathEscape(ref), nil, &parsed); err != nil {
		return Mod{}, err
	}
	if parsed.Mod == nil {
		return Mod{}, fmt.Errorf("vintage story mod db: %w", &httpkit.StatusError{Code: http.StatusNotFound, Status: "404 Not Found"})
	}
	return *parsed.Mod, nil
}

// Tags is the site's tags by their lower-cased name, each its id as the listing takes it.
func (c *Client) Tags(ctx context.Context) (map[string]string, error) {
	var parsed struct {
		Tags []struct {
			ID   json.Number `json:"tagid"`
			Name string      `json:"name"`
		} `json:"tags"`
	}
	if err := c.get(ctx, "/tags", nil, &parsed); err != nil {
		return nil, err
	}
	ids := make(map[string]string, len(parsed.Tags))
	for _, t := range parsed.Tags {
		ids[strings.ToLower(t.Name)] = t.ID.String()
	}
	return ids, nil
}

// ID joins a mod's modid and assetid into the id the site's own install links and pages agree on.
func ID(modID, assetID int) string { return strconv.Itoa(modID) + "-" + strconv.Itoa(assetID) }

// SplitID is the modid and assetid of id; an id without an assetid still names its mod, its assetid then empty.
func SplitID(id string) (modID, assetID string, ok bool) {
	modID, assetID, _ = strings.Cut(id, "-")
	if _, err := strconv.Atoi(modID); err != nil {
		return "", "", false
	}
	if _, err := strconv.Atoi(assetID); err != nil {
		assetID = ""
	}
	return modID, assetID, true
}

// PageURL is the mod's page: the author's alias when there is one, the address the site itself links, else the asset's
// own path.
func PageURL(alias string, assetID int) string {
	if alias != "" {
		return SiteURL + "/" + url.PathEscape(alias)
	}
	return SiteURL + "/show/mod/" + strconv.Itoa(assetID)
}

// PreRelease reports a version its author marks as a test build (1.0.0-rc.1, 2.0.0-pre.8, 1.2-dev.3).
func PreRelease(version string) bool {
	core, _, _ := strings.Cut(version, "+")
	return strings.Contains(core, "-")
}

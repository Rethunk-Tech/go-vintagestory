package moddb

import (
	"errors"
	"regexp"
	"strings"
)

// Scheme is the scheme of the site's 1-click install button.
const Scheme = "vintagestorymodinstall"

// Ref names one release in an install link: the mod id its modinfo.json gives and the release's version.
type Ref struct{ ModID, Version string }

var (
	refModID   = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	refVersion = regexp.MustCompile(`^[0-9A-Za-z.+-]{1,40}$`)
	errBadLink = errors.New("not a vintagestorymodinstall link")
)

// ParseLink reads vintagestorymodinstall://<modid>@<version>, as the Mod DB's install button writes it.
func ParseLink(raw string) (Ref, error) {
	rest, ok := strings.CutPrefix(raw, Scheme+"://")
	if !ok {
		return Ref{}, errBadLink
	}
	modID, ver, ok := strings.Cut(strings.TrimSuffix(rest, "/"), "@")
	if !ok || !refModID.MatchString(modID) || !refVersion.MatchString(ver) {
		return Ref{}, errBadLink
	}
	return Ref{ModID: modID, Version: ver}, nil
}

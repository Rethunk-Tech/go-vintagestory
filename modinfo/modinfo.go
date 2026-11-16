// Package modinfo reads a Vintage Story mod's modinfo.json, the manifest at the root of every mod folder.
package modinfo

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/Rethunk-Tech/go-lenientjson"
)

// ModInfo is what a mod's modinfo.json says of its mod.
type ModInfo struct {
	Name, ModID, Version string
	Authors              []string
	// Dependencies are the mod ids the file requires, lower-cased and sorted; the values (versions) are not read.
	Dependencies []string
	// GameVersion is the game version the mod requires, its "game" dependency, when it names one.
	GameVersion string
}

// Parse reads a modinfo.json as leniently as the game reads it with Newtonsoft: comments, trailing commas and any key
// case. Newtonsoft sets a property once per occurrence, so of several spellings of one key the last in the file wins.
func Parse(b []byte) (ModInfo, error) {
	clean := lenientjson.Clean(b)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(clean, &obj); err != nil {
		return ModInfo{}, err
	}
	if obj == nil {
		return ModInfo{}, errors.New("modinfo is not a JSON object")
	}
	raw, _ := members(clean)
	info := ModInfo{Name: text(raw, "name"), ModID: text(raw, "modid"), Version: text(raw, "version")}
	var authors []string
	_ = json.Unmarshal(field(raw, "authors"), &authors)
	for _, a := range authors {
		if a = strings.TrimSpace(a); a != "" {
			info.Authors = append(info.Authors, a)
		}
	}
	if deps, ok := members(field(raw, "dependencies")); ok {
		for _, d := range deps {
			info.Dependencies = append(info.Dependencies, strings.ToLower(d.key))
			var ver string
			var n json.Number
			if json.Unmarshal(d.val, &n) == nil && n != "" {
				ver = n.String()
			} else if json.Unmarshal(d.val, &ver) != nil {
				continue
			}
			if ver = strings.TrimSpace(ver); strings.EqualFold(d.key, "game") && ver != "*" {
				info.GameVersion = strings.TrimPrefix(ver, "v")
			}
		}
		slices.Sort(info.Dependencies)
		info.Dependencies = slices.Compact(info.Dependencies)
	}
	return info, nil
}

type member struct {
	key string
	val json.RawMessage
}

// members lists an object's entries in document order, which a map would lose.
func members(b json.RawMessage) ([]member, bool) {
	dec := json.NewDecoder(bytes.NewReader(b))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	var out []member
	for dec.More() {
		k, err := dec.Token()
		key, isKey := k.(string)
		var v json.RawMessage
		if err != nil || !isKey || dec.Decode(&v) != nil {
			return out, false
		}
		out = append(out, member{key, v})
	}
	return out, true
}

func field(raw []member, name string) json.RawMessage {
	for _, m := range slices.Backward(raw) {
		if strings.EqualFold(m.key, name) {
			return m.val
		}
	}
	return nil
}

func text(raw []member, name string) string {
	f := field(raw, name)
	var n json.Number // Newtonsoft reads a bare number as the string it spells
	if json.Unmarshal(f, &n) == nil && n != "" {
		return n.String()
	}
	var s string
	if json.Unmarshal(f, &s) != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

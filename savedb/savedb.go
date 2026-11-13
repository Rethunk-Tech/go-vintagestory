// Package savedb reads a Vintage Story world's .vcdbs: a SQLite database whose single gamedata row is a protobuf
// SaveGame.
package savedb

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"math"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SaveGame fields of a world's gamedata row, from the protobuf schema the game is built from
// (schema-1.21.0.rc.3.proto).
const (
	vsWorldName          = 13
	vsTotalSecondsPlayed = 14
	vsLastSavedVersion   = 21
	vsPlayStyle          = 29
	vsSavegameIdentifier = 34
	vsModData            = 11
)

// baseDomains are the game's own asset domains; every other domain of a block or item code is the id of the mod that
// added it.
var baseDomains = []string{"game", "creative", "survival"}

var errBadProto = errors.New("the world's game data is not a protobuf message")

// World is what a .vcdbs records of itself.
type World struct {
	// Name is the world's title and PlayStyle the code of its play style (surviveandbuild, creativebuilding, ...).
	Name, PlayStyle string
	// GameVersion is the game version that last saved the world.
	GameVersion string
	// Played is the total play time the world records.
	Played time.Duration
	// WorldID is the world's SavegameIdentifier, a GUID that names its files outside the saves folder.
	WorldID string
	// Domains are the asset domains of the mods that added blocks or items to the world, lower-cased and sorted. A mod
	// of code alone leaves no trace.
	Domains []string
}

// Read opens the world at path and reads its gamedata row. The database is opened read-only and immutable, so a world
// the game holds open is readable and a read never writes to it.
func Read(ctx context.Context, path string) (World, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return World{}, err
	}
	conn, err := sql.Open("sqlite", readOnlyURI(path))
	if err != nil {
		return World{}, err
	}
	defer func() { _ = conn.Close() }()
	var data []byte
	if err := conn.QueryRowContext(ctx, "SELECT data FROM gamedata LIMIT 1").Scan(&data); err != nil {
		return World{}, err
	}
	var w World
	used := map[string]struct{}{}
	err = eachField(data, func(num, v uint64, b []byte) {
		switch num {
		case vsWorldName:
			w.Name = string(b)
		case vsTotalSecondsPlayed:
			// An int32 in the game; a negative one is a broken count, not a play time.
			if v <= math.MaxInt32 {
				w.Played = time.Duration(v) * time.Second
			}
		case vsLastSavedVersion:
			w.GameVersion = string(b)
		case vsPlayStyle:
			w.PlayStyle = string(b)
		case vsSavegameIdentifier:
			w.WorldID = string(b)
		case vsModData:
			addCodeDomains(b, used)
		}
	})
	if err != nil {
		return World{}, err
	}
	w.Domains = w.Domains[:0]
	for d := range used {
		w.Domains = append(w.Domains, d)
	}
	slices.Sort(w.Domains)
	return w, nil
}

// readOnlyURI is the SQLite URI for db, opened read-only and immutable. A Windows drive letter has to keep its colon,
// which no URL path may carry, so it gains a leading slash.
func readOnlyURI(db string) string {
	p := filepath.ToSlash(db)
	if len(p) > 1 && p[1] == ':' {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: p, RawQuery: "mode=ro&immutable=1"}).String()
}

// addCodeDomains adds to used the domain of each code in one ModData entry when it is the world's block or item id map
// (map<int32,string> of codes, "domain:path" for a mod's content).
func addCodeDomains(entry []byte, used map[string]struct{}) {
	var key string
	var value []byte
	_ = eachField(entry, func(num, _ uint64, b []byte) {
		switch num {
		case 1:
			key = string(b)
		case 2:
			value = b
		}
	})
	if key != "BlockIDs" && key != "ItemIDs" {
		return
	}
	_ = eachField(value, func(num, _ uint64, pair []byte) {
		if num != 1 {
			return
		}
		_ = eachField(pair, func(n, _ uint64, code []byte) {
			if domain, _, ok := strings.Cut(string(code), ":"); n == 2 && ok && domain != "" && !slices.Contains(baseDomains, domain) {
				used[strings.ToLower(domain)] = struct{}{}
			}
		})
	})
}

// eachField calls f with each top-level field of a protobuf message: a varint's value, or a length-delimited field's
// bytes. Fixed-width fields and groups (protobuf-net writes some nested objects as groups) are skipped.
func eachField(msg []byte, f func(num, v uint64, b []byte)) error {
	depth := 0
	for len(msg) > 0 {
		key, n := binary.Uvarint(msg)
		if n <= 0 {
			return errBadProto
		}
		msg = msg[n:]
		num := key >> 3
		switch key & 7 {
		case 0:
			v, n := binary.Uvarint(msg)
			if n <= 0 {
				return errBadProto
			}
			msg = msg[n:]
			if depth == 0 {
				f(num, v, nil)
			}
		case 1, 5:
			size := 8
			if key&7 == 5 {
				size = 4
			}
			if len(msg) < size {
				return errBadProto
			}
			msg = msg[size:]
		case 2:
			l, n := binary.Uvarint(msg)
			if n <= 0 || l > math.MaxInt32 {
				return errBadProto
			}
			size, rest := int(l), msg[n:]
			if size > len(rest) {
				return errBadProto
			}
			if depth == 0 {
				f(num, 0, rest[:size])
			}
			msg = rest[size:]
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return errBadProto
			}
			depth--
		default:
			return errBadProto
		}
	}
	return nil
}

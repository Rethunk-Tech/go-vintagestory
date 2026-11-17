package savedb

import (
	"database/sql"
	"encoding/binary"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// FuzzEachField holds that no bytes make the gamedata walk panic or read past its buffer, and that a walked field's
// bytes never reach outside the message they came from.
func FuzzEachField(f *testing.F) {
	f.Add([]byte{0x6a, 0x03, 0x0a, 0x01, 'x'})              // field 13, a name
	f.Add([]byte{0x0a, 0x02, 0x08, 0x01})                   // field 1, a nested message
	f.Add([]byte{0x8a, 0x01, 0xff, 0xff, 0xff, 0xff, 0x0f}) // a length wider than the message
	f.Add([]byte{0x0b, 0x6a, 0x01, 'x', 0x0c})              // a group wrapping a name
	f.Add([]byte{0x6a})                                     // a key with no value
	f.Fuzz(func(t *testing.T, msg []byte) {
		_ = eachField(msg, func(_, _ uint64, b []byte) {
			if len(b) > len(msg) {
				t.Fatalf("field of %d bytes from a %d-byte message", len(b), len(msg))
			}
		})
	})
}

// FuzzAddCodeDomains holds that a ModData entry yields only domains a mod could own: non-empty, lower-case, free of the
// separator, and never a base-game domain.
func FuzzAddCodeDomains(f *testing.F) {
	for _, key := range []string{"BlockIDs", "ItemIDs", "Other"} {
		pair := protoString(nil, 2, "MyMod:plank")
		ids := binary.AppendUvarint(nil, 1<<3|2)
		ids = binary.AppendUvarint(ids, uint64(len(pair)))
		ids = append(ids, pair...)
		entry := protoString(nil, 1, key)
		entry = binary.AppendUvarint(entry, 2<<3|2)
		entry = binary.AppendUvarint(entry, uint64(len(ids)))
		f.Add(append(entry, ids...))
	}
	f.Fuzz(func(t *testing.T, entry []byte) {
		used := map[string]struct{}{}
		_ = addCodeDomains(entry, used)
		for d := range used {
			if d == "" || d != strings.ToLower(d) || strings.Contains(d, ":") || slices.Contains(baseDomains, d) {
				t.Fatalf("domain %q from %x", d, entry)
			}
		}
	})
}

// FuzzRead holds that a .vcdbs of any bytes fails rather than panics, and that a world it accepts never names a field
// longer than the file or a domain the game itself owns.
func FuzzRead(f *testing.F) {
	f.Add([]byte("not sqlite"))
	f.Add([]byte{})
	dir := f.TempDir()
	path := filepath.Join(dir, "seed.vcdbs")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		f.Fatal(err)
	}
	var data []byte
	data = protoString(data, vsWorldName, "Bjornheim")
	data = protoString(data, vsLastSavedVersion, "1.21.4")
	data = protoString(data, vsPlayStyle, "surviveandbuild")
	data = protoString(data, vsSavegameIdentifier, "9f8e7d6c-5b4a-3210-9876-543210fedcba")
	var ids []byte
	for _, code := range []string{"air", "creative:lamp", "primitivesurvival:fishbasket"} {
		ids = binary.AppendUvarint(ids, 1<<3|2)
		ids = binary.AppendUvarint(ids, uint64(len(protoString(nil, 2, code))))
		ids = append(ids, protoString(nil, 2, code)...)
	}
	entry := protoString(nil, 1, "BlockIDs")
	entry = binary.AppendUvarint(entry, 2<<3|2)
	entry = binary.AppendUvarint(entry, uint64(len(ids)))
	entry = append(entry, ids...)
	data = binary.AppendUvarint(data, vsModData<<3|2)
	data = binary.AppendUvarint(data, uint64(len(entry)))
	data = append(data, entry...)
	if _, err := conn.Exec("CREATE TABLE gamedata (savegameid integer PRIMARY KEY, data BLOB)"); err != nil {
		f.Fatal(err)
	}
	if _, err := conn.Exec("INSERT INTO gamedata VALUES (1, ?)", data); err != nil {
		f.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		f.Fatal(err)
	}
	seed, err := os.ReadFile(path)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Fuzz(func(t *testing.T, file []byte) {
		p := filepath.Join(t.TempDir(), "w.vcdbs")
		if err := os.WriteFile(p, file, 0o600); err != nil {
			t.Fatal(err)
		}
		w, err := Read(t.Context(), p)
		if err != nil {
			return
		}
		n := len(file)
		if len(w.Name) > n || len(w.GameVersion) > n || len(w.PlayStyle) > n || len(w.WorldID) > n || len(w.Domains) > n {
			t.Fatalf("field longer than file")
		}
		for _, d := range w.Domains {
			if d == "" || d != strings.ToLower(d) || strings.Contains(d, ":") || slices.Contains(baseDomains, d) {
				t.Fatalf("domain %q", d)
			}
		}
	})
}

// FuzzReadOnlyURI holds that a world path always opens as a file URI pinned read-only and immutable.
func FuzzReadOnlyURI(f *testing.F) {
	f.Add("C:/saves/Bjornheim.vcdbs")
	f.Add("/saves/Bjornheim.vcdbs")
	f.Fuzz(func(t *testing.T, db string) {
		raw := readOnlyURI(db)
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("readOnlyURI(%q) = %q: %v", db, raw, err)
		}
		if u.Scheme != "file" {
			t.Fatalf("scheme %q from %q", u.Scheme, db)
		}
		q := u.Query()
		if q.Get("mode") != "ro" || q.Get("immutable") != "1" {
			t.Fatalf("query = %q", u.RawQuery)
		}
	})
}

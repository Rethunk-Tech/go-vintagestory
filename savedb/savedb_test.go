package savedb

import (
	"database/sql"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func protoString(b []byte, num uint64, s string) []byte {
	b = binary.AppendUvarint(b, num<<3|2)
	b = binary.AppendUvarint(b, uint64(len(s)))
	return append(b, s...)
}

func TestReadReadsAWorldsGamedata(t *testing.T) {
	dir := t.TempDir()
	var data []byte
	data = binary.AppendUvarint(data, 7<<3|5) // a fixed32 the reader skips
	data = append(data, 1, 2, 3, 4)
	data = protoString(data, vsWorldName, "Bjornheim")
	data = binary.AppendUvarint(data, 40<<3|3) // a group whose own name field must not win
	data = protoString(data, vsWorldName, "inner")
	data = binary.AppendUvarint(data, 40<<3|4)
	data = binary.AppendUvarint(data, vsTotalSecondsPlayed<<3)
	data = binary.AppendUvarint(data, 7200)
	data = protoString(data, vsLastSavedVersion, "1.21.4")
	data = protoString(data, vsPlayStyle, "surviveandbuild")
	data = protoString(data, vsSavegameIdentifier, "9f8e7d6c-5b4a-3210-9876-543210fedcba")
	// ModData["BlockIDs"]: a map of block ids to codes, a mod's codes in its own domain.
	var ids []byte
	for i, code := range []string{"air", "creative:lamp", "primitivesurvival:fishbasket", "PrimitiveSurvival:weir"} {
		var pair []byte
		pair = binary.AppendUvarint(pair, 1<<3)
		pair = binary.AppendUvarint(pair, uint64(i))
		pair = protoString(pair, 2, code)
		ids = binary.AppendUvarint(ids, 1<<3|2)
		ids = binary.AppendUvarint(ids, uint64(len(pair)))
		ids = append(ids, pair...)
	}
	entry := protoString(nil, 1, "BlockIDs")
	entry = binary.AppendUvarint(entry, 2<<3|2)
	entry = binary.AppendUvarint(entry, uint64(len(ids)))
	entry = append(entry, ids...)
	data = binary.AppendUvarint(data, vsModData<<3|2)
	data = binary.AppendUvarint(data, uint64(len(entry)))
	data = append(data, entry...)
	path := filepath.Join(dir, "Bjornheim.vcdbs")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"PRAGMA journal_mode=WAL", "CREATE TABLE gamedata (savegameid integer PRIMARY KEY, data BLOB)"} {
		if _, err := conn.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := conn.ExecContext(t.Context(), "INSERT INTO gamedata VALUES (1, ?)", data); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	w, err := Read(t.Context(), path)
	if err != nil || w.Name != "Bjornheim" || w.GameVersion != "1.21.4" || w.PlayStyle != "surviveandbuild" ||
		w.Played != 2*time.Hour || w.WorldID != "9f8e7d6c-5b4a-3210-9876-543210fedcba" ||
		!slices.Equal(w.Domains, []string{"primitivesurvival"}) {
		t.Fatalf("world = %+v, %v", w, err)
	}
}

func TestReadOfAFileThatIsNotSQLiteFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Broken.vcdbs")
	if err := os.WriteFile(path, []byte("not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(t.Context(), path); err == nil {
		t.Fatal("a file that is not a database read as one")
	}
}

func TestReadOpensARelativeWorldPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "world.vcdbs")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(t.Context(), "CREATE TABLE gamedata (savegameid integer PRIMARY KEY, data BLOB)"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(t.Context(), "INSERT INTO gamedata VALUES (1, ?)", []byte{}); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if _, err := Read(t.Context(), "world.vcdbs"); err != nil {
		t.Fatal(err)
	}
}

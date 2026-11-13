package savedb

import (
	"encoding/binary"
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
		addCodeDomains(entry, used)
		for d := range used {
			if d == "" || d != strings.ToLower(d) || strings.Contains(d, ":") || slices.Contains(baseDomains, d) {
				t.Fatalf("domain %q from %x", d, entry)
			}
		}
	})
}

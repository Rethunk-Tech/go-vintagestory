package savedb

import "testing"

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

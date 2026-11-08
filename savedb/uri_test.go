package savedb

import "testing"

func TestReadOnlyURIPinsADriveLetter(t *testing.T) {
	// A drive letter's colon would otherwise read as a URL scheme, so the path gains a leading slash.
	if got := readOnlyURI("C:/saves/Bjornheim.vcdbs"); got != "file:///C:/saves/Bjornheim.vcdbs?mode=ro&immutable=1" {
		t.Fatalf("drive uri = %q", got)
	}
	if got := readOnlyURI("/saves/Bjornheim.vcdbs"); got != "file:///saves/Bjornheim.vcdbs?mode=ro&immutable=1" {
		t.Fatalf("posix uri = %q", got)
	}
}

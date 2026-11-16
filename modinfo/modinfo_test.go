package modinfo

import (
	"slices"
	"testing"
)

func TestParseReadsAnyKeyCaseAndLenientJSON(t *testing.T) {
	info, err := Parse([]byte(`{
		// The game reads this with Newtonsoft, which takes comments and trailing commas.
		"Type": "code", "ModID": "carryon", "Name": "Carry On", "Version": "1.14.3",
		"authors": ["copygirl", "Nikea"],
		"Dependencies": { "game": "1.22.0", "VSImGui": "" },
	}`))
	if err != nil || info.Name != "Carry On" || info.ModID != "carryon" || info.Version != "1.14.3" ||
		!slices.Equal(info.Authors, []string{"copygirl", "Nikea"}) || !slices.Equal(info.Dependencies, []string{"game", "vsimgui"}) || info.GameVersion != "1.22.0" {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

func TestParseReadsANumberAsItsText(t *testing.T) {
	info, err := Parse([]byte(`{"modid": "a", "version": 1.2}`))
	if err != nil || info.Version != "1.2" {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

func TestParseReadsANumericGameVersion(t *testing.T) {
	info, err := Parse([]byte(`{"dependencies":{"game":1.21}}`))
	if err != nil || info.GameVersion != "1.21" {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

func TestParseTrimsWildcardGameVersion(t *testing.T) {
	info, err := Parse([]byte(`{"dependencies":{"game":" * "}}`))
	if err != nil || info.GameVersion != "" {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

func TestParseDuplicateModIDCaseIsDeterministic(t *testing.T) {
	in := []byte(`{"modid":"lower","ModID":"UPPER"}`)
	for range 32 {
		info, err := Parse(in)
		if err != nil || info.ModID != "UPPER" {
			t.Fatalf("info = %+v, %v", info, err)
		}
	}
}

func TestParseDuplicateGameDepCase(t *testing.T) {
	info, err := Parse([]byte(`{"dependencies":{"game":"1.20","Game":"1.21"}}`))
	if err != nil || !slices.Equal(info.Dependencies, []string{"game"}) || info.GameVersion != "1.21" {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

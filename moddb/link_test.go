package moddb

import "testing"

func TestParseLinkReadsTheInstallButton(t *testing.T) {
	for _, raw := range []string{"vintagestorymodinstall://carryon@1.14.3", "vintagestorymodinstall://carryon@0.13.0-pre.2/"} {
		ref, err := ParseLink(raw)
		want := Ref{ModID: "carryon", Version: "1.14.3"}
		if raw != "vintagestorymodinstall://carryon@1.14.3" {
			want.Version = "0.13.0-pre.2"
		}
		if err != nil || ref != want {
			t.Errorf("ParseLink(%q) = %+v, %v; want %+v", raw, ref, err, want)
		}
	}
	for _, raw := range []string{"vintagestorymodinstall://carryon", "https://mods.vintagestory.at/carryon", "vintagestorymodinstall://../x@1"} {
		if ref, err := ParseLink(raw); err == nil {
			t.Errorf("ParseLink(%q) = %+v, want an error", raw, ref)
		}
	}
}

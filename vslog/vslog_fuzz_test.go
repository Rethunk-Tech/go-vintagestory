package vslog

import (
	"strings"
	"testing"
)

func FuzzParseLine(f *testing.F) {
	f.Add("30.11.2025 18:29:40 [Notification] Game Version: v1.21.5 (Stable)\r")
	f.Add("1.1.2024 12:00:00 [Notification] hello")
	f.Add("12.12.2024 00:00:00 [Error] [modid] msg")
	f.Add("1.1.2024 12:00:00 [Fatal] x")
	f.Add("1.1.2024 12:00:00 [Notification] ")
	f.Add("   at ImmersiveLight.ImmersiveLightModSystem.Start()")
	f.Add("Game Version: v1.21.5 (Stable)")
	f.Add("")
	f.Add("not a log line")
	f.Fuzz(func(t *testing.T, line string) {
		l, ok := ParseLine(line)
		if !ok {
			return
		}
		if len(l.Time) > len(line) || len(l.Level) > len(line) || len(l.Message) > len(line) {
			t.Fatalf("field longer than input")
		}
	})
}

func FuzzGameVersion(f *testing.F) {
	f.Add("Game Version: v1.21.5 (Stable)")
	f.Add("Game Version: v1.22.0-rc.4 (Unstable)")
	f.Add("Game Version: v1.20.0")
	f.Add("")
	f.Add("no version")
	f.Fuzz(func(t *testing.T, message string) {
		v := GameVersion(message)
		if v != "" && !strings.Contains(message, v) {
			t.Fatalf("version %q not in input", v)
		}
	})
}

func FuzzFirstException(f *testing.F) {
	f.Add("\nSystem.NullReferenceException: Object reference not set to an instance of an object.\n   at ImmersiveLight.ImmersiveLightModSystem.Start()")
	f.Add("System.Exception: boom\n  at Foo")
	f.Add("\n\n  x  \n")
	f.Add("")
	f.Fuzz(func(t *testing.T, blob string) {
		got := FirstException(strings.Split(blob, "\n"))
		if got == "" {
			t.Fatal("empty")
		}
	})
}

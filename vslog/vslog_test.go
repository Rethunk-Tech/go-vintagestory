package vslog

import "testing"

// The lines are the shapes a client-main.log writes, dates and all.
func TestParseLineReadsTheHeaderAndLeavesTheLevelWordAlone(t *testing.T) {
	got, ok := ParseLine("30.11.2025 18:29:40 [Notification] Game Version: v1.21.5 (Stable)\r")
	if !ok || got != (Line{Time: "18:29:40", Level: "Notification", Message: "Game Version: v1.21.5 (Stable)"}) {
		t.Fatalf("line = %+v, %v", got, ok)
	}
	if _, ok := ParseLine("   at ImmersiveLight.ImmersiveLightModSystem.Start()"); ok {
		t.Fatal("a stack trace line parsed as a log line")
	}
}

func TestGameVersionReadsTheVersionTheGameLogsAtStart(t *testing.T) {
	line, ok := ParseLine("30.11.2025 18:29:40 [Notification] Game Version: v1.21.5 (Stable)")
	if !ok || GameVersion(line.Message) != "1.21.5" {
		t.Fatalf("line %+v: version %q", line, GameVersion(line.Message))
	}
	if GameVersion("Client logger started.") != "" {
		t.Fatal("a line naming no version gave one")
	}
}

func TestFirstExceptionSkipsToTheThrowingLine(t *testing.T) {
	lines := []string{
		"", "System.NullReferenceException: Object reference not set to an instance of an object.",
		"   at ImmersiveLight.ImmersiveLightModSystem.Start()",
	}
	if got := FirstException(lines); got != "System.NullReferenceException: Object reference not set to an instance of an object." {
		t.Fatalf("exception = %q", got)
	}
	if got := FirstException(nil); got != "threw an exception" {
		t.Fatalf("empty log = %q", got)
	}
}

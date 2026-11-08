// Package vslog reads the lines Vintage Story writes to its client-main.log and server-main.log, and the failures its
// mod loader records there.
package vslog

import (
	"regexp"
	"strings"
)

// Line is one log line the game wrote: the time of day, the game's own level word, and the message.
type Line struct {
	Time, Level, Message string
}

// The game writes `d.M.yyyy HH:mm:ss [Level] message`, with the level word one of Fatal, Error, Warning, Notification,
// Event, Chat, Audit, Build, StoryEvent, Worldgen, Debug or VerboseDebug.
var header = regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{4} (\d\d:\d\d:\d\d) \[(\w+)\] ?(.*)$`)

// ParseLine splits one log line. It reports false for a line in another shape, such as the continuation of a stack
// trace. Level is the game's word, unmapped: a caller decides which words mean what in its own vocabulary.
func ParseLine(line string) (Line, bool) {
	m := header.FindStringSubmatch(strings.TrimRight(line, "\r"))
	if m == nil {
		return Line{}, false
	}
	return Line{Time: m[1], Level: m[2], Message: m[3]}, true
}

// versionLine is the line the game logs at start: "Game Version: v1.21.5 (Stable)".
var versionLine = regexp.MustCompile(`^Game Version: v(\d+(?:\.\d+)+)`)

// GameVersion is the game version a log message names, without its "v"; empty when it names none.
func GameVersion(message string) string {
	if m := versionLine.FindStringSubmatch(message); m != nil {
		return m[1]
	}
	return ""
}

var (
	// FailureLine matches an error or fatal line the game attributes to one mod: the level, the mod id in brackets,
	// and the message. Its groups are the level, the mod id and the message.
	FailureLine = regexp.MustCompile(`^\d{1,2}\.\d{1,2}\.\d{4} \d\d:\d\d:\d\d \[(Error|Fatal)\] \[([^\]\s]+)\] ?(.*)$`)
	// DependencyLine matches one indented dependency line under a failure: the dependency's id, the version asked for,
	// and the site's or loader's reason, all separated by " - ".
	DependencyLine = regexp.MustCompile(`^\s+(\S+?)@(\S+) - (.*)$`)
)

// Messages the game writes at error level under the id of the mod that failed.
const (
	// Unresolved opens the block of dependency lines of a mod whose dependencies were not met.
	Unresolved = "Could not resolve some dependencies:"
	// StartFailed and LoadFailed open the stack trace of a mod that threw while starting.
	StartFailed = "An exception was thrown when trying to start the mod:"
	LoadFailed  = "An exception was thrown when trying to load assembly:"
	// NoAssembly is part of the message of a mod declared as a code mod that ships no .dll.
	NoAssembly = "declared as code mod, but there are no .dll files"
)

// FirstException is the first non-blank line after a failure header, which names what threw; a fallback when the log
// holds none.
func FirstException(lines []string) string {
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return "threw an exception"
}

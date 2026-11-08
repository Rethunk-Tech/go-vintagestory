# AGENTS.md

`vintagestory` holds the on-disk and over-the-wire formats of Vintage Story, one package per format:
mod manifests, game logs, world databases and the Mod DB API. Any Go program that reads a mod folder,
a log or a world, or that talks to the Mod DB, imports it; it imports nothing from them, so it never
knows what a caller does with a mod or a save.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH.

## Shape

- `savedb.Read` is the one path into a `.vcdbs`, and it opens the database read-only and immutable.
  A second opener would let a read take a write lock on a world the game has open, or write a WAL into
  a player's saves folder.
- The protobuf walker (`eachField`) and the domain scan (`addCodeDomains`) stay unexported: they are
  an implementation of one row's decode, not a general protobuf reader, and a caller that reached for
  them would depend on field numbers the game may renumber.
- Domains are plain strings, not a caller's mod-id type. The format holds a lower-case asset domain,
  and only a caller knows whether that domain names a mod it manages.
- `moddb` holds no cache. The site's listing and tag answers are cheap but its caching window is a
  policy decision — how stale a caller tolerates — and policy belongs to the caller, not to a format
  client. `New` keeps only configuration, which is what makes it safe for concurrent use.
- Every `moddb` request goes through `httpkit.DoCapped` with `maxBody`, and the client wraps the given
  `*http.Client` with `httpkit.WithRedirect(..., httpkit.SameOrigin)`: a redirect to another origin
  would send the request somewhere the caller never named, and an unbounded body would let a hostile
  answer exhaust memory.
- `vslog` exports patterns and the raw level word instead of a level type. The game's vocabulary
  (Notification, Build, Audit, StoryEvent, ...) is wider than any one consumer maps, and only the
  caller knows which of its own levels those words should become.

## Files

| File | Holds |
| --- | --- |
| `doc.go` | The module's package sentence. |
| `modinfo/modinfo.go` | `ModInfo` and `Parse`: a mod's modinfo.json, read as leniently as the game reads it. |
| `vslog/vslog.go` | `Line`, `ParseLine`, `GameVersion`, `FirstException`, and the patterns and messages of the game's loader failures. |
| `savedb/savedb.go` | `World` and `Read`: a `.vcdbs` world, its SQLite URI rule, and the protobuf walk over its gamedata row. |
| `moddb/moddb.go` | `Config`, `New`, `Client` and its `Mods`/`Mod`/`Tags`, the release selection, and the id and page helpers. |
| `moddb/link.go` | `ParseLink` and `Ref`: the site's 1-click install scheme. |

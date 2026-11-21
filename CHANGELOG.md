# Changelog

Notable changes to `go-vintagestory`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `modinfo`: the last spelling of a key wins, as Newtonsoft does, duplicate-case fields and game dependencies are deterministic, a JSON number is accepted as the game version, a bare number is read as its text, and the game version is trimmed before the wildcard check.
- `savedb`: relative world paths are resolved before opening and kept in the file URI path, a world path with a NUL byte is rejected, protobuf with unclosed groups is rejected, truncated block and item id maps are surfaced, and base domains are skipped in any case.
- `vslog`: `ParseLine` strips trailing newlines, and `GameVersion` keeps the prerelease token.
- `moddb`: build metadata is ignored when judging a prerelease, a missing mod is an httpkit not-found, and live `Tags` answers whose `tagid` is a number decode.

## [0.1.0] - 2026-11-08

### Added

- `modinfo`: `Parse` reads a modinfo.json as leniently as the game does — any key case, comments and trailing commas — into `ModInfo`.
- `vslog`: `ParseLine` splits a client-main.log or server-main.log line, `GameVersion` reads the version line, and `FirstException` plus the exported patterns locate a mod's loader failure.
- `savedb`: `Read` opens a world's `.vcdbs` read-only and returns its name, play style, game version, play time, world id and the mod domains its blocks and items record.
- `moddb`: `Client` lists, reads and tags against the Vintage Story Mod DB, with release selection (`Newest`, `PreRelease`), id and page helpers, and `ParseLink` for `vintagestorymodinstall://` links.

[Unreleased]: https://github.com/Rethunk-Tech/go-vintagestory/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Rethunk-Tech/go-vintagestory/releases/tag/v0.1.0

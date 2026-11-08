<h1 align="center">go-vintagestory</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

Vintage Story formats: modinfo manifests, client logs, save databases and the VS Mod DB API. Package `vintagestory`, module `github.com/Rethunk-Tech/go-vintagestory`.

Four packages, one per format, each an independent import: nothing here talks to a game install, a mod
folder layout or a filesystem policy, and nothing caches. A caller hands in bytes, a database path or an
HTTP client, and gets back plain values. Parsing is as lenient as the game itself, so a file the game
loads also parses here; a file it would reject returns an error rather than a partial value.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-vintagestory
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- `modinfo.Parse` reads a mod folder's modinfo.json the way the game does: any key case, comments, trailing commas, with dependencies and the required game version.
- `vslog.ParseLine` splits a main log's lines into time, level word and message; `vslog.FirstException` and the exported patterns find what a mod threw at start.
- `savedb.Read` opens a world's `.vcdbs` read-only and immutable, and returns its name, play style, game version, play time and the mod domains that added blocks or items.
- `moddb.Client` lists the Mod DB, reads one mod with its releases, and maps the site's tag names to ids, over a byte-capped response with cross-origin redirects refused.
- `moddb` also answers what an install needs to choose a release: `Mod.Newest`, `PreRelease`, `ID`, `SplitID`, `PageURL` and `ParseLink` for `vintagestorymodinstall://` links.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.

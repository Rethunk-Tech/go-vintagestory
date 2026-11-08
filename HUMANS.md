# HUMANS.md

Run and use `go-vintagestory`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-vintagestory
```

```go
package main

import (
 "context"
 "fmt"
 "log"

 "github.com/Rethunk-Tech/go-vintagestory/modinfo"
 "github.com/Rethunk-Tech/go-vintagestory/moddb"
 "github.com/Rethunk-Tech/go-vintagestory/savedb"
)

func main() {
 raw := []byte(`{"modid": "carryon", "name": "Carry On", "version": "1.14.3"}`)
 info, err := modinfo.Parse(raw)
 if err != nil {
  log.Fatal(err)
 }
 fmt.Println(info.ModID, info.Version)

 world, err := savedb.Read(context.Background(), "/saves/Bjornheim.vcdbs")
 if err != nil {
  log.Fatal(err)
 }
 fmt.Println(world.Name, world.Domains)

 client := moddb.New(moddb.Config{UserAgent: "example/1.0"})
 mod, err := client.Mod(context.Background(), info.ModID)
 if err != nil {
  log.Fatal(err)
 }
 if release, ok := mod.Newest(false); ok {
  fmt.Println(moddb.PageURL(mod.URLAlias, mod.AssetID), release.MainFile)
 }
}
```

## Contracts

| Entry point | What a caller can rely on |
| --- | --- |
| `modinfo.Parse` | Reads a manifest as leniently as the game: any key case, comments, trailing commas. Dependencies come back lower-cased and sorted; a manifest that is not a JSON object is an error, never a zero value. |
| `vslog.ParseLine` | Returns the time of day, the game's own level word untouched, and the message. Reports false for any line in another shape, such as a stack trace's. |
| `vslog.GameVersion` | The version a log message names, its "v" stripped; empty when the message names none. |
| `vslog.FirstException` | The first non-blank line after a failure header, or a fixed fallback; never empty. |
| `savedb.Read` | Opens the database read-only and immutable, so a world the game holds open is readable and nothing is written. Fields the row does not hold stay empty; a row that is not a protobuf message is an error. Domains are lower-cased, sorted and free of the game's own. |
| `moddb.New` | Keeps the client's configuration only, so it is safe for concurrent use; redirects to another origin are refused. |
| `moddb.Client.Mods` | Returns the whole match in one answer, since the API has no offset; no result is an empty slice, not nil. A text search without an order is ordered by downloads. |
| `moddb.Client.Mod` | A reference the site does not know is an error; a 404 is a `*httpkit.StatusError` matching `httpkit.ErrNotFound`. |
| `moddb.Mod.Newest` | Without `includePre`, the newest release that is not a test build and has a file; the newest test build only when no stable release has one. |
| `moddb.ParseLink` | Accepts `vintagestorymodinstall://<modid>@<version>` with an optional trailing slash, and nothing else: the mod id and version may hold only the characters the site writes. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```

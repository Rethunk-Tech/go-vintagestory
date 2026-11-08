# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

- `modinfo.Parse` takes an untrusted modinfo.json. A crafted manifest must not panic, and a value of
  the wrong JSON type is read as absent rather than as an error.
- `vslog.ParseLine`, `vslog.GameVersion` and `vslog.FirstException` take untrusted log text. No line,
  however long or malformed, may panic or hang: no pattern backtracks, and every scan is linear.
- `savedb.Read` takes an untrusted `.vcdbs`. A crafted database must not panic, and a crafted gamedata
  row must not make the protobuf walk read past its buffer or allocate beyond the row: every length is
  checked against the bytes that remain. The database is opened read-only and immutable, so a read
  never changes a world.
- `moddb` takes untrusted answers from the network. A crafted body must not panic any decoder, and must
  not exceed `maxBody`; a redirect to another origin is refused rather than followed.
- `moddb.ParseLink` takes an untrusted URI. A link it accepts names only characters that are safe in a
  file name and a path segment.

## What is out of scope

- The path a caller passes to `savedb.Read` is trusted: that the world is where the caller says, and
  that reading it is allowed, is the caller's decision. Only the database's own contents are untrusted.
- A caller's `*http.Client` is trusted. Its transport, proxy and timeouts are the caller's policy; the
  library adds only the same-origin redirect rule.
- An answer that parses into fewer fields than expected is intended: the site adds fields over time, and
  an unknown field is ignored rather than refused.
- Verifying what a downloaded release contains is out of scope. `moddb` names a release's address and
  version; checking the file is the caller's.

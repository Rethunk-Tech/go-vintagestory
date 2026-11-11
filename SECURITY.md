# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting on this repository: <https://github.com/Rethunk-Tech/go-vintagestory/security/advisories/new>, or the **Security** tab, then **Report a vulnerability**. The report stays private between you and us until a fix ships.

If private reporting is unavailable, email [oss@rethunk.tech](mailto:oss@rethunk.tech) with **[SECURITY]** in the subject.

Please do not open a public issue, pull request or discussion for a suspected vulnerability.

A useful report has the affected repository and version (or commit), steps to reproduce, what an attacker gets and under which conditions, and whether anyone else knows.

## What counts as a vulnerability

A vulnerability is a flaw that lets someone cross a boundary they should not: run code, read or change data they have no right to, escape a sandbox or path restriction, or leak a credential. A bug is anything else that misbehaves, such as a crash on your own input, wrong output or a slow path; open an ordinary issue for those. If you are unsure which one you have, report it privately and we will say.

## Coordinated disclosure

We handle the report inside the private advisory, so you see each status change and can comment on the fix. The advisory goes public when the fix ships, or 90 days after your report, whichever comes first. If a fix is close, either of us can ask for more time in the thread. You are credited in the advisory unless you would rather not be.

## Response timelines

- We acknowledge your report within 3 business days.
- We give you an assessment (accepted, needs more information or declined, with reasons) within 10 business days.
- We aim to ship a fix within 90 days of the report.

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

## Safe harbour

We will not pursue legal action against good-faith research that avoids privacy violations, data destruction and disruption of services, stops at a proof of concept, and keeps the details private until the advisory is published. Rethunk.Tech, LLC maintains this repository.

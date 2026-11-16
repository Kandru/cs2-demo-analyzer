# AGENTS.md — internal/report

Canonical report document shared by the HTML page and the JSON sidecar.

## Files

- `report.go` — `Header`, `Document`, JSON encoding/decoding, HTML/JSON header
  probes, freshness comparison, demo fingerprint, report options.
- `report_test.go` — header layout, round trip, probing, freshness, escaping.

## Invariants

- `Header` is always the first JSON member. `Document` embeds `*model.Match` so
  the JSON sidecar stays flat and backwards compatible.
- HTML artifacts start with `<!doctype html>` followed by exactly one
  `<!--cs2-demo-analyzer:{...}-->` line. The comment is produced by
  `HeaderComment`, never by a template.
- Only controlled or JSON-escaped values may enter the header. Map/server names
  are untrusted; never add player names, paths, or free-form match text.
- `Header.Matches` compares only cache-affecting fields. Presentation/index
  metadata (`Map`, `ServerName`, `Score`, `PlayedAt`, `GeneratedAt`) must not
  invalidate a report.
- Bump `HeaderFormat` when the header encoding changes in a way that would let
  an old artifact pass the freshness check. Bump `model.SchemaVersion` for
  incompatible match data changes.
- Header probing reads a bounded prefix for HTML and only the first JSON member;
  it must never scan a whole report.
- Do not add format-specific rendering here. HTML lives in `html`; new formats
  consume `Document` and embed the same header.

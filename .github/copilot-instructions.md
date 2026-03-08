# Copilot Instructions

## Project Overview

Go implementation of [JSON Pointer (RFC 6901)](https://datatracker.ietf.org/doc/html/rfc6901) for navigating
and mutating JSON documents represented as Go values. Works with `map[string]any`, slices, and Go structs
(resolved via `json` struct tags and reflection).

## Package Layout (single package)

| File | Contents |
|------|----------|
| `pointer.go` | Core types (`Pointer`, `JSONPointable`, `JSONSetable`), `New`, `Get`, `Set`, `Offset`, `Escape`/`Unescape` |
| `errors.go` | Sentinel errors: `ErrPointer`, `ErrInvalidStart`, `ErrUnsupportedValueType` |

## Key API

- `New(string) (Pointer, error)` — parse a JSON pointer string (e.g. `"/foo/0/bar"`)
- `Pointer.Get(document any) (any, reflect.Kind, error)` — retrieve a value
- `Pointer.Set(document, value any) (any, error)` — set a value (document must be pointer/map/slice)
- `Pointer.Offset(jsonString string) (int64, error)` — byte offset of token in raw JSON
- `GetForToken` / `SetForToken` — single-level convenience helpers
- `Escape` / `Unescape` — RFC 6901 token escaping (`~0` ↔ `~`, `~1` ↔ `/`)

Custom types can implement `JSONPointable` (for Get) or `JSONSetable` (for Set) to bypass reflection.

## Design Decisions

- Struct fields **must** have a `json` tag to be reachable; untagged fields are ignored.
- Anonymous embedded struct fields are traversed only if tagged.
- The RFC 6901 `"-"` array suffix (append) is **not** implemented.

## Dependencies

- `github.com/go-openapi/swag/jsonname` — struct tag to JSON field name resolution
- `github.com/go-openapi/testify/v2` — test-only assertions (zero-dep fork of `stretchr/testify`)

## Conventions

- All `.go` files must have SPDX license headers (Apache-2.0).
- Commits require DCO sign-off (`git commit -s`).
- Linting: `golangci-lint run` — config in `.golangci.yml` (posture: `default: all` with explicit disables).
- Every `//nolint` directive **must** have an inline comment explaining why.
- Tests: `go test ./...` with `-race`. CI runs on `{ubuntu, macos, windows} x {stable, oldstable}`.
- Test framework: `github.com/go-openapi/testify/v2` (not `stretchr/testify`).

See `.github/copilot/` (symlinked to `.claude/rules/`) for detailed rules on Go conventions, linting, and testing.

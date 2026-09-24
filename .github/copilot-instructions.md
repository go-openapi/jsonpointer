# Copilot Instructions

## Project Overview

Go implementation of [JSON Pointer (RFC 6901)](https://datatracker.ietf.org/doc/html/rfc6901) for navigating
and mutating JSON documents represented as Go values. Works with `map[string]any`, slices, and Go structs
(resolved via `json` struct tags and reflection).

## Package Layout

Two packages in one module (`github.com/go-openapi/jsonpointer`):

| File | Contents |
|------|----------|
| `pointer.go` | `Pointer`, `New`, `Get`, `Set`, `Offset`, `GetForToken`, `SetForToken`, `Escape`/`Unescape`, and the reflection helpers |
| `ifaces.go` | `JSONPointable`, `JSONSetable`, `NameProvider` interfaces |
| `options.go` | `Option`, `WithNameProvider`, `SetDefaultNameProvider`, `UseGoNameProvider`, `DefaultNameProvider` |
| `errors.go` | Sentinel errors and their constructors |
| `jsonname/name_provider.go` | `NameProvider` — the default, tag-only resolver (`DefaultJSONNameProvider`) |
| `jsonname/go_name_provider.go` | `GoNameProvider` — the alternative resolver, follows `encoding/json` rules |

## Key API

- `New(string) (Pointer, error)` — parse a JSON pointer string (e.g. `"/foo/0/bar"`)
- `Pointer.Get(document any, opts ...Option) (any, reflect.Kind, error)` — retrieve a value
- `Pointer.Set(document, value any, opts ...Option) (any, error)` — set a value (document must be a pointer, map or slice)
- `Pointer.Offset(jsonString string) (int64, error)` — byte offset of token in raw JSON
- `GetForToken` / `SetForToken` — single-level convenience helpers
- `Escape` / `Unescape` — RFC 6901 token escaping (`~0` ↔ `~`, `~1` ↔ `/`)

Custom types can implement `JSONPointable` (for Get) or `JSONSetable` (for Set) to bypass reflection.

Sentinel errors: `ErrPointer`, `ErrInvalidStart`, `ErrUnsupportedValueType`, `ErrDashToken`.
Errors raised by the package wrap `ErrPointer`. Errors returned by a caller's `JSONLookup` or
`JSONSet` are passed through unchanged, so `errors.Is(err, ErrPointer)` is false for those.

## Name providers

Two implementations of the `NameProvider` interface resolve go field names to json names. The
difference drives most "why is my field unreachable" questions:

- `jsonname.NameProvider` (the default, `jsonname.DefaultJSONNameProvider`): only fields carrying a
  non-empty `json` tag are reachable.
- `jsonname.GoNameProvider` (opt in with `UseGoNameProvider()` or `WithNameProvider`): follows
  `encoding/json`, so untagged exported fields are reachable under their go name.

## Design Decisions

- Under the default provider, struct fields **must** have a `json` tag to be reachable; untagged
  fields are ignored. Use `UseGoNameProvider()` for `encoding/json` semantics.
- An anonymous (embedded) field is walked for the fields it promotes, and its own `json` tag is
  ignored — tagged or not, the tag never becomes a reachable key.
- Embedding is traversed through a pointer (`*Base`). Resolving a promoted field on a value whose
  embedded pointer is nil reports an error; it does not allocate the intermediate struct.
- An anonymous field that is not a struct (e.g. an embedded named slice) promotes nothing and is
  unreachable.
- Maps must be keyed by a string type. A named string type is converted; `map[int]V` and friends
  cannot be addressed by a pointer and report an error.
- `Set` with a nil value (what `encoding/json` decodes a JSON null into) writes the zero value when
  the target can hold nil — interface, pointer, map, slice, chan, func. On a target that cannot
  represent null (a `string` field, an element of `[]int`) it reports an error rather than
  substituting a zero value. Setting a map member to nil keeps the member; it does not delete it.
- The RFC 6901 `"-"` array suffix **is** supported on `Pointer.Set` as an append operation
  (RFC 6902 convention). On `Pointer.Get` and `Pointer.Offset` it is always an error per
  RFC 6901 §4. See `ErrDashToken`.

## Dependencies

- `github.com/go-openapi/testify/v2` — test-only assertions (zero-dep fork of `stretchr/testify`)

The module has no non-test dependency. `jsonname` is a subpackage of this repository, not the
similarly named `github.com/go-openapi/swag/jsonname`.

## Conventions

- All `.go` files must have SPDX license headers (Apache-2.0).
- Commits require DCO sign-off (`git commit -s`).
- Linting: `golangci-lint run` — config in `.golangci.yml` (posture: `default: all` with explicit
  disables). Format with `golangci-lint fmt`, and check a change with
  `golangci-lint run --new-from-rev master`.
- Every `//nolint` directive **must** have an inline comment explaining why.
- Tests: `go test ./...` with `-race`. CI runs `{ubuntu, macos, windows} x {stable, oldstable}`,
  via the shared `go-openapi/ci-workflows/.github/workflows/go-test.yml`.
- Test framework: `github.com/go-openapi/testify/v2` (not `stretchr/testify`).

See `.github/copilot/` (symlinked to `.claude/rules/`) for detailed rules on Go conventions, linting, testing, and contributions.

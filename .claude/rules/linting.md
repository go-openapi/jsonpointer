---
paths:
  - "**/*.go"
---

# Linting conventions (go-openapi)

```sh
golangci-lint run
```

Config: `.golangci.yml` — posture is `default: all` with explicit disables.
Each disable is listed there under `linters.disable`.

Format with `golangci-lint fmt` (not `gofmt` or `gofumpt` directly — the config drives which
formatters run).

Check a change against the base branch before pushing, so pre-existing reports in untouched code
do not drown out your own:

```sh
golangci-lint run --new-from-rev master
```

Key rules:
- Every `//nolint` directive **must** have an inline comment explaining why.
- Prefer disabling a linter over scattering `//nolint` across the codebase.

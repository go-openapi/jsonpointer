# Coding style at `go-openapi`

**DRAFT**

> **TL;DR**
>
> We've never been super-strict on code style etc.
> But now go-openapi and go-swagger make a large codebase to maintain and keep afloat.
>
> Code quality and the harmonization of rules have thus become something that we need now.

## Meta-linter

Universally formatted go code promotes ease of writing, reading, and maintenance.

You should run `golangci-lint run` before committing your changes.

Many editors have plugins that do that automatically.

> We use the `golangci-lint` meta-linter. The configuration lies in `.golangci-lint.yml`.
> You may read <https://golangci-lint.run/docs/linters/configuration/>  for additional reference.

## Linting rules posture

Thanks to go's original design, we developers don't have to waste much time arguing about code figures of style.

We enable all linters published by `golangci-lint` by default, then disable a few ones.

Here are the reasons why they are disabled:

```yaml
  disable:
    - depguard              # we don't want to configure rules to constrain import. That's the reviewer's job
    - exhaustruct           # we don't want to configure regexp's to check type name. That's the reviewer's job
    - funlen                # we accept cognitive complexity as a meaningful metric, but function length is relevant
    - godox                 # we don't see any value in forbidding TODO's etc in code
    - nlreturn              # we usually apply this "blank line" rule to make code less compact. We just don't want to enforce it.
    - nonamedreturns        # we don't see any valid reason why we couldn't used named returns.
    - noinlineerr           # there is no value added forbidding inlined err
    - paralleltest          # we like parallel tests. We just don't want this to be enforced everywhere.
    - recvcheck             # we like the idea of having pointer and non-pointer receivers
    - testpackage           # we like test packages. We just don't want it to be enforced everywhere.
    - tparallel             # see paralleltest
    - varnamelen            # sometimes, we like short variables
    - whitespace            # no added value
    - wrapcheck             # although there is some sense with this linter's general idea, it produces too much noise
    - wsl                   # no added value. Noise.
    - wsl_v5                # no added value. Noise.
```

Enabled linters with relaxed constraints:

```yaml
  settings:
    dupl:
      threshold: 200        # in a older code base such as ours, we have to be tolerant with a little redundancy
    goconst:
      min-len: 2
      min-occurrences: 3
    cyclop:
      max-complexity: 20    # the default is too low for most of our functions. 20 is a nicer trade-off
    gocyclo:
      min-complexity: 20
    exhaustive:             # when using default in switch, this should be good enough
      default-signifies-exhaustive: true
      default-case-required: true
    lll:
      line-length: 180      # we just want to avoid extremely long lines. It is no big deal if a line or two don't fit on your terminal.
```

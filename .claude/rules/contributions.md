---
paths:
  - "**/*"
---

# Contribution rules (go-openapi)

Read the go-openapi contribution guide before opening a pull request, plus the sibling rules in
this directory: go conventions, linting, testing, github workflows and technical writing.

## Commit hygiene

- Every commit **must** be DCO signed-off (`git commit -s`) with a real email address.
  PGP-signed commits are appreciated but not required.
- Agents may be listed as co-authors (`Co-Authored-By:`) but the commit **author must be the human sponsor**.
  We do not accept commits solely authored by bots or agents.
- Squash commits into logical units of work before requesting review (`git rebase -i`).

## Linting

Before pushing, verify your changes pass linting against the base branch:

```sh
golangci-lint run --new-from-rev master
```

Install the latest version if you don't have it:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

## Problem statement

- Clearly describe the problem the PR solves, or reference an existing issue.
- PR descriptions must not be vague ("fix bug", "improve code") — explain *what* was wrong and *why* the change is correct.

## Tests are mandatory

- Every bug fix or feature **must** include tests that demonstrate the problem and verify the fix.
- The only exceptions are documentation changes and typo fixes.
- Aim for at least 80% coverage of your patch.
- Run the full test suite before submitting:

```sh
go test ./...
```

In a mono-repo, run it once per module (a Go workspace does not give `go test` a whole-workspace
target).

Run the race detector too, since CI does:

```sh
go test -race ./...
```

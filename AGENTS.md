# Agent instructions

Feed is a Go learning project developed through homework assignments.

Read [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) for architecture, data flows, and
implementation details. Use [README.md](README.md) for setup and usage. Check the
relevant source before making changes; update the context when behavior changes.

## Working rules

- Implement the current assignment with small, readable changes. Prefer the
  existing standard library and direct SQL; explain any necessary new dependency.
- Inspect `git status --short` and preserve unrelated edits. Existing TODOs and
  documented limitations do not expand the task.
- Preserve database records and keep schema changes compatible with existing
  tables. Do not reset data or delete volumes without an explicit user request.
- Keep secrets and local `.env` contents out of commits and responses.
- Preserve parameterized SQL, template escaping, password hashing, and session
  synchronization. Update both locale catalogs when adding interface text.
- Format changed Go files with `gofmt`. Keep documentation in English without
  decorative emoji, generated-by banners, or agent activity logs.
- Update affected documentation and CHANGELOG for notable changes. Explicit user
  instructions take precedence over these conventions.

## Validation

For code changes, run from the repository root:

```sh
go test ./...
go vet ./...
go build -o bin/feed ./cmd/app
git diff --check
```

For documentation-only changes, check claims, paths, links, and
`git diff --check`. Add focused regression tests for nontrivial behavior changes.
There are currently no automated tests; compilation alone does not verify the
application. Use the relevant manual checks in PROJECT_CONTEXT with a disposable
database. Report checks actually run and any failures or gaps.

# AGENTS.md for Trongo (Go project)

## Build/Lint/Test Commands
- `make help`: List targets
- `make check`: fmt + lint + test (pre-commit)
- `make test`: All tests
- `make test/coverage`: Tests w/ ≥75% coverage (coverage.html)
- `make go/fmt`: gofmt
- `make go/vet`: govet
- `make go/lint`: golangci-lint
- `make go/build`: Build binary

Single test: `go test -v ./pkg/name/... -run TestFunc`

## Code Style Guidelines
- Go 1.21+, Testify testing
- Tests: Table-driven, `*_test.go`, `TestFunc(t *testing.T)`, httptest, mocks, ≥75% cov
- Docs: Full sentences for exports (godoc)
- Files: Hyphens (no _), secrets/ dir
- Commits: Conventional (feat(scope): desc)
- Patterns: `wantErr` in tables, consumer interfaces, `make check` pre-commit
- Formatting: gofmt/goimports
- Errors: if err != nil
- Imports: std/third/local groups

See docs/warp.md (workflow) & docs/warp-go.md (Go details). No Cursor/Copilot rules.
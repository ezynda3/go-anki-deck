# Agent Guidelines for go-anki-deck

## Build/Test Commands
- Run all tests: `make test` or `go test -v -race ./...`
- Run single test: `go test -v -run TestName ./...`
- Run benchmarks: `make bench` or `go test -bench=. -benchmem ./...`
- Lint code: `make lint` (requires golangci-lint)
- Format code: `make fmt` or `go fmt ./...`
- Build: `make build` or `go build -v ./...`
- Coverage: `make coverage`

## Code Style Guidelines
- **Language**: Go 1.23.10
- **Package**: All code in `package anki` except examples
- **Imports**: Group stdlib, blank line, then external deps (e.g., `github.com/mattn/go-sqlite3`)
- **Error Handling**: Return errors as last value, check immediately after function calls
- **Naming**: Use camelCase for exported functions/types, lowercase for internal
- **Testing**: Test files end with `_test.go`, use `t.Fatalf` for setup failures, `t.Errorf` for assertions
- **Comments**: Add godoc comments for all exported types and functions
- **Dependencies**: Run `go mod tidy` after adding new imports
- **SQL**: Use prepared statements, always close database connections with defer
# gmailctl Development Guide

## Project Overview

gmailctl is a Go CLI tool for declarative Gmail filter management. It allows users to write filters in Jsonnet, validates them, generates diffs, and applies changes via Gmail API without manual XML imports.

**Key Components:**
- **Config Layer** (`internal/engine/config/`): Jsonnet parsing → v1alpha3.Config struct
- **Parser Layer** (`internal/engine/parser/`): Config → AST → Simplified criteria (query optimization)
- **Filter Engine** (`internal/engine/filter/`): AST → Gmail API filter format with diff computation
- **Apply Layer** (`internal/engine/apply/`): Orchestrates config parsing, diff generation, and API updates
- **Commands** (`cmd/gmailctl/cmd/`): CLI interface using cobra (init, apply, diff, download, test, etc.)

## Architecture Patterns

### Config Version System
- Current version: `v1alpha3` (in `internal/engine/config/v1alpha3/config.go`)
- Jsonnet files must specify `version: 'v1alpha3'` in the root object
- **Migration path**: Use `cmd/gmailctl-config-migrate/` to convert between versions
- YAML configs are explicitly **not supported** (returns error with help message)

### AST Simplification & Query Optimization
The parser performs multi-pass query simplification (`internal/engine/parser/ast.go`):
- Flattens nested AND/OR operations (e.g., `and[and[a,b],c]` → `and[a,b,c]`)
- Groups same-function queries (e.g., `from:a OR from:b` → `from:{a b}`)
- Gmail has a 1500-char limit per filter, so simplification is critical

Example simplification:
```go
// Before: {or: [{from: "a"}, {from: "b"}]}
// After: Single leaf with Function=FunctionFrom, Args=["a", "b"]
```

### Dependency Injection via APIProvider
`cmd/gmailctl/cmd/api_provider.go` defines the `GmailAPIProvider` interface. The main binary injects `localcred.Provider{}` for OAuth2 flows. Tests inject `fakegmail` (see Testing section).

### Standard Library (gmailctl.libsonnet)
Located at `internal/data/gmailctl.libsonnet`, embedded via `//go:embed`:
- `chainFilters(fs)`: Creates if-elsif chains by ANDing negations of previous filters
- `directlyTo(recipient)`: Matches TO field only (excludes CC/BCC)
- Import in configs: `local lib = import 'gmailctl.libsonnet';`

## Development Workflows

### Build & Test
```bash
# Build the binary
go build ./cmd/gmailctl

# Run unit tests
go test ./...

# Run integration tests with golden file updates
go test -v -update ./...

# Generate coverage report
./hack/coverage.sh  # Creates HTML coverage in temp file
```

### Integration Tests
`integration_test.go` uses:
- **Golden files** in `testdata/valid/`: `.jsonnet` (input), `.json` (expected config), `.xml` (expected Gmail export), `.diff` (expected diff)
- **fakegmail** (`internal/fakegmail/`): In-memory Gmail API server for hermetic testing
- **Update flag**: `go test -update` regenerates golden files (review changes carefully!)

Example test flow:
1. Parse `.jsonnet` config
2. Apply to fake Gmail server
3. Export to XML, compare with `.xml` golden
4. Import back, compare with `.json` golden
5. Diff against upstream, compare with `.diff` golden

### Testing Philosophy
- Use `testify/require` for assertions (fails fast)
- Test files use `require.Nil(t, err)` pattern
- Parser tests (`internal/engine/parser/ast_test.go`) verify simplification logic
- Use `reporting.Prettify(obj, colorize)` for debug output

## Critical Conventions

### Error Handling
Use `internal/errors` package for wrapped errors:
```go
errors.WithCause(err, ErrNotFound)  // Wraps with sentinel
errors.WithDetails(err, "help text")  // Adds user-facing details
```

### Color Output
- Commands support `--color=always|auto|never` flag
- `shouldUseColorDiff()` checks flag, TERM env, and TTY status
- Use `reporting.ColorizeDiff()` for diff output (fatih/color package)
- Recent change: Color diff feature was added to the `color-diffs` branch

### Filter Criteria Structure
`FilterNode` (config) → `CriteriaAST` (parser) → Gmail query string (filter)
- **Leaf nodes**: `{from: "x"}` → Leaf with Function=FunctionFrom, Args=["x"]
- **Operation nodes**: `{and: [...]}` → Node with Operation=OperationAnd, Children=[...]
- **IsEscaped flag**: When true, skip quoting/escaping (used for downloaded filters)

### Labels Management
- Labels are opt-in: Only managed if `labels` field is present in config
- Supports nested labels with `/` separator (e.g., `work/important`)
- Parent labels are auto-created during apply
- Deletion removes labels from all messages (irreversible warning shown)

## File Organization

```
cmd/gmailctl/          # Main binary
  cmd/                 # Cobra commands (apply, diff, etc.)
  localcred/           # OAuth2 provider implementation
internal/engine/       # Core logic (no imports from cmd/)
  config/              # Jsonnet → Config parsing
  parser/              # Config → AST transformation
  filter/              # AST → Gmail API filters + diffing
  apply/               # Orchestration layer
  api/                 # Gmail API client wrapper
  label/               # Label management
  export/              # XML export for manual import
  rimport/             # Import from Gmail (download command)
internal/data/         # Embedded Jsonnet library
testdata/valid/        # Integration test golden files
```

## Common Tasks

### Adding a New Filter Criterion
1. Add field to `FilterNode` in `config/v1alpha3/config.go`
2. Add `FunctionType` constant in `parser/ast.go`
3. Implement parsing in `parser/parser.go` (`parseFunction`)
4. Add query string generation in `filter/filter.go`
5. Update tests and documentation

### Debugging Filter Generation
Use `gmailctl debug -f config.jsonnet` to see:
- Parsed config structure
- Simplified AST
- Generated Gmail query strings

### Working with Jsonnet
- Jsonnet imports resolved relative to config file directory
- VM is created fresh for each parse (no persistent state)
- Standard library is always available via `import 'gmailctl.libsonnet'`

## Dependencies of Note

- `google.golang.org/api/gmail/v1`: Gmail API client
- `github.com/google/go-jsonnet`: Jsonnet interpreter
- `github.com/spf13/cobra`: CLI framework
- `github.com/fatih/color`: Terminal color output
- `github.com/pmezard/go-difflib`: Diff computation
- `internal/graph/`: Minimal fork of gosl graph package for munkres algorithm (label matching)

## Config Test System

Users can add `tests` array to config for validation:
- Tests run via `gmailctl test` command
- Limitations: Cannot test `isEscaped` expressions or raw queries
- See `cmd/gmailctl/cmd/test_cmd.go` and `internal/engine/cfgtest/`

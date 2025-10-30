# Zed IDE Configuration for KYC-DSL

This directory contains Zed IDE configuration for debugging and development tasks.

## Quick Start

### Running Tasks

Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Linux/Windows) and type "Tasks: Spawn" to see available tasks.

### Available Tasks

#### Development Tasks
- **Run KYC DSL (sample_case.dsl)** - Run the application with the sample DSL file
- **Build KYC DSL** - Build the kycctl binary
- **Format Code** - Format all Go code with gofmt
- **Vet Code** - Run go vet static analysis

#### Testing Tasks
- **Run Tests** - Run all tests with verbose output
- **Run Tests (Parser)** - Run only parser tests
- **Debug Tests (Parser)** - Debug parser tests with Delve

#### Debugging Tasks
- **Debug KYC DSL with Delve** - Launch interactive debugger for main application

#### Linting Tasks
- **Lint (golangci-lint)** - Run all linters

#### Database Tasks
- **Database: View All Cases** - Show all cases in the database
- **Database: View Latest Cases** - Show only the latest version of each case
- **Database: Clear All Cases** - Delete all cases (use with caution!)

#### Cleanup Tasks
- **Clean Build Cache** - Clear Go build cache

## Debugging with Delve

### Interactive Debugging

1. Run the "Debug KYC DSL with Delve" task
2. Use these commands in the Delve console:

```
b main.main                    # Set breakpoint at main function
b parser.ParseFile             # Set breakpoint in parser
b internal/engine/engine.go:23 # Set breakpoint at specific line
c                              # Continue execution
n                              # Next line
s                              # Step into function
p variableName                 # Print variable
locals                         # Show all local variables
goroutines                     # List all goroutines
exit                           # Exit debugger
```

### Common Breakpoint Locations

Based on the debug logging already in the code:

- `cmd/kycctl/main.go:23` - Main entry point
- `internal/parser/parser.go:43` - ParseFile start
- `internal/parser/parser.go:52` - Parse function start
- `internal/engine/engine.go:29` - RunCase start
- `internal/storage/postgres.go:71` - InsertCase start

### Example Debugging Session

```bash
$ dlv debug ./cmd/kycctl -- sample_case.dsl
Type 'help' for list of commands.
(dlv) b main.main
Breakpoint 1 set at 0x... for main.main()
(dlv) c
> main.main() ./cmd/kycctl/main.go:12
(dlv) n
(dlv) p file
"sample_case.dsl"
(dlv) c
```

## Settings

The `settings.json` file configures:

- **Go Language Settings** - Tab size, hard tabs, format on save
- **LSP (gopls)** - Language server configuration with static analysis
- **Debug Configurations** - Pre-configured debug sessions
- **Terminal Environment** - PostgreSQL environment variables

## Debug Logging

The code includes comprehensive debug logging that can be toggled:

Set `DEBUG = false` in these files to disable debug output:
- `cmd/kycctl/main.go`
- `internal/parser/parser.go`
- `internal/engine/engine.go`
- `internal/storage/postgres.go`

## Database Configuration

Default PostgreSQL settings (can be overridden with environment variables):

```bash
PGHOST=localhost
PGPORT=5432
PGUSER=adamtc007
PGDATABASE=kyc_dsl
```

## Tips

1. **Quick Run**: Use `Cmd+Shift+P` → "Tasks: Spawn" → "Run KYC DSL"
2. **Quick Debug**: Use `Cmd+Shift+P` → "Tasks: Spawn" → "Debug KYC DSL with Delve"
3. **View Results**: Use "Database: View All Cases" task to see inserted records
4. **Clean Up**: Use "Database: Clear All Cases" to reset the database

## Troubleshooting

### Delve Not Found
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Database Connection Failed
Check that PostgreSQL is running:
```bash
psql -d kyc_dsl -c "SELECT 1;"
```

### gopls Not Working
Install or update gopls:
```bash
go install golang.org/x/tools/gopls@latest
```

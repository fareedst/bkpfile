# BkpFile

A command-line application for creating and managing file backups.

## Installation

```bash
go install bkpfile@latest
```

## Configuration

Create a `.bkpfile.yml` file in the directory containing the file you want to backup:

```yaml
backup_dir_path: "../.bkpfile"  # Where backups are stored
use_current_dir_name: true      # Whether to include current directory name in backup path
```

If no configuration file is present, default values will be used.

## Usage

### Create a Backup

```bash
bkpfile backup [FILE_PATH] [NOTE]
```

Example:
```bash
bkpfile backup important.txt "monthly_backup"
```

### List Backups

```bash
bkpfile list [FILE_PATH]
```

Example:
```bash
bkpfile list important.txt
```

### Global Options

- `--dry-run`: Show what would be done without creating backups

Example:
```bash
bkpfile backup important.txt --dry-run
```

## Backup Naming

Backups are named using the following format:
- `[PREFIX-]FILENAME-[YYYY-MM-DD-hh-mm][=NOTE]`

Where:
- `PREFIX` is the source directory name (if `use_current_dir_name` is true)
- `FILENAME` is the name of the original file
- `YYYY-MM-DD-hh-mm` is the timestamp
- `NOTE` is an optional note provided by the user

Example: `docs-main.go-2024-03-20-15-30=monthly_backup`

## Development

### Building

Use the provided Makefile for building:

```bash
# Build for local development
make build-local

# Build for all platforms
make build-all

# Build for specific platforms
make build-macos
make build-ubuntu
```

### Testing

```bash
make test
```

### Linting

This project uses [revive](https://github.com/mgechev/revive) for Go code linting.

#### Running the linter

```bash
# Run revive linter
make lint

# Run comprehensive linting (fmt + vet + revive)
make lint-fix
```

#### Linting Rules

The project uses a custom `revive.toml` configuration with the following enabled rules:

- **exported**: Checks for exported functions/types documentation (with stuttering check disabled)
- **package-comments**: Ensures packages have proper comments
- **var-naming**: Validates variable naming conventions
- **var-declaration**: Checks variable declarations
- **struct-tag**: Validates struct tags
- **receiver-naming**: Ensures consistent receiver naming
- **error-strings**: Validates error string formatting
- **error-naming**: Checks error variable naming
- **dot-imports**: Prevents dot imports
- **blank-imports**: Validates blank imports
- **context-as-argument**: Ensures context is passed as first argument
- **context-keys-type**: Validates context key types
- **range**: Checks range loop usage
- **range-val-in-closure**: Prevents range value capture issues
- **range-val-address**: Prevents taking address of range values
- **unexported-return**: Checks unexported return types
- **time-naming**: Validates time-related naming
- **string-of-int**: Prevents string(int) conversions
- **string-format**: Validates string formatting
- **early-return**: Encourages early returns
- **unhandled-error**: Flags unhandled errors
- **defer**: Validates defer usage
- **unexported-naming**: Checks unexported naming
- **errorf**: Validates fmt.Errorf usage

#### Installing revive

If you need to install revive separately:

```bash
go install github.com/mgechev/revive@latest
```

## Project Structure

```
bkpfile/
├── cmd/bkpfile/          # Main application entry point
├── internal/bkpfile/     # Internal packages
├── revive.toml          # Revive linter configuration
├── .golangci.yml        # golangci-lint configuration
├── Makefile             # Build and development tasks
└── go.mod               # Go module definition
``` 
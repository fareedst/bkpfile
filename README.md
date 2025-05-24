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
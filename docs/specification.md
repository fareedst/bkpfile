# BkpFile: Single File Backup CLI Application

## Overview
BkpFile is a command-line application for macOS and Linux that creates backups of individual files. It supports customizable naming patterns and maintains a history of file backups.

> **Important**: This document describes the user-facing features and behaviors. For immutable specifications that cannot be changed without a major version bump, see [Immutable Specifications](immutable.md).

## Documentation Navigation

### For Users
- Start with this [Specification](specification.md) document
- Refer to [Immutable Specifications](immutable.md) for core behaviors that cannot change

### For Developers
- Begin with [Architecture](architecture.md) for system design
- Follow [Requirements](requirements.md) for implementation details
- Use [Testing](testing.md) for test coverage requirements

### For Contributors
- Review [Immutable Specifications](immutable.md) first to understand constraints
- Follow [Testing](testing.md) requirements for all changes

### Document Maintenance
- Keep [Specification](specification.md) and [Immutable Specifications](immutable.md) in sync
- Update [Requirements](requirements.md) with new features
- Maintain test coverage as per [Testing](testing.md)
- All changes must preserve existing functionality per [Immutable Specifications](immutable.md)

## Configuration Discovery
- Configuration files are discovered using a configurable search path
- The search path is controlled by the `BKPFILE_CONFIG` environment variable
- If `BKPFILE_CONFIG` is not set, the default search path is: `./.bkpfile.yml:~/.bkpfile.yml`
- Configuration files are processed in order, with values from earlier files taking precedence
- If multiple configuration files exist, settings in earlier files override settings in later files

### Environment Variable: BKPFILE_CONFIG
- Specifies a colon-separated list of configuration file paths to search
- Example: `BKPFILE_CONFIG="/etc/bkpfile.yml:~/.config/bkpfile.yml:./.bkpfile.yml"`
- Paths can be absolute or relative
- Relative paths are resolved from the current working directory
- Home directory (`~`) expansion is supported

## Configuration File
- Configuration is stored in YAML files with names specified by the configuration discovery system
- If no configuration files are found, default values are used (see [Immutable Specifications](immutable.md#configuration-defaults))
- Configuration files use the `.yml` extension by convention

### Configuration Options
1. **Configuration Search Path**
   - Controls which configuration files are searched
   - Default: `./.bkpfile.yml:~/.bkpfile.yml`
   - YAML key: `config`
   - Example: `config: ./.bkpfile.yml:~/.bkpfile.yml`

2. **Backup Directory Path**
   - Specifies where backups are stored
   - Default: `../.bkpfile` relative to current directory
   - YAML key: `backup_dir_path`
   - Backups maintain the source file's directory structure in the backup path

3. **Use Current Directory Name**
   - Controls whether to include current directory name in the backup path
   - Default: `true`
   - YAML key: `use_current_dir_name`
   - Example: With file 'cmd/bkpfile/main.go', backup path becomes '../.bkpfile/cmd/bkpfile/main.go-2025-05-12-13-49'

## Commands

### 1. List Backups
- Displays all backups associated with the specified file
- Usage: `bkpfile --list [FILE_PATH]`
- Shows each backup with its path (relative to current directory) and creation time in the format:
  ```
  .bkpfile/path/to/file.txt-2024-03-21-15-30=note (created: 2024-03-21 15:30:00)
  ```
- Backups are sorted by creation time (most recent first)
- Backups are organized by their source file paths

### 2. Create Backup
- Creates a copy of the specified file
- Usage: `bkpfile [FILE_PATH] [NOTE]`
- Before creating a backup:
  - Compares the file with its most recent backup using byte comparison
  - If the file is identical to the most recent backup:
    - Reports the existing backup path (relative to current directory)
    - Exits normally without creating a new backup
- Backup naming format: `SOURCE_FILENAME-YYYY-MM-DD-hh-mm[=NOTE]`
  - SOURCE_FILENAME is the base name of the original file
  - YYYY-MM-DD-hh-mm is the timestamp of the backup
  - NOTE is an optional note appended with an equals sign
- The backup maintains the original file's directory structure in the backup path
- NOTE is an optional positional argument provided by the user

## Global Options
- **Dry-Run Mode**: When enabled with `--dry-run` flag:
  - Shows the backup filename that would be created
  - No actual backup is created

## Implementation Details
For detailed implementation requirements and constraints, see:
- [Immutable Specifications](immutable.md) for core behaviors that cannot be changed
- [Architecture](architecture.md) for system design and implementation details
- [Requirements](requirements.md) for technical requirements and test coverage

## Platform Compatibility
- Works on macOS and Linux systems
- Uses platform-independent path handling
- Preserves file permissions and ownership where applicable
- Handles file system differences between platforms

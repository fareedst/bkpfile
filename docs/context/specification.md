# BkpFile: Single File Backup CLI Application

## Overview
BkpFile is a command-line application for macOS and Linux that creates backups of individual files. It supports customizable naming patterns, maintains a history of file backups, and provides robust error handling with automatic resource cleanup.

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
- Ensure all code passes linting requirements before submission

### Document Maintenance
- Keep [Specification](specification.md) and [Immutable Specifications](immutable.md) in sync
- Update [Requirements](requirements.md) with new features
- Maintain test coverage as per [Testing](testing.md)
- All changes must preserve existing functionality per [Immutable Specifications](immutable.md)

## Quality Assurance and Code Standards

### Linting Requirements
- All Go code must pass `revive` linter checks before commit
- Linting configuration is maintained in `.revive.toml`
- Run linting with `make lint` command
- Code must follow Go best practices and naming conventions
- All errors must be properly handled (no unhandled return values)

### Error Handling Standards
- All backup operations return structured errors with status codes
- Enhanced disk space detection for various storage conditions
- Panic recovery mechanisms prevent application crashes
- Context support for operation cancellation and timeouts
- Comprehensive error logging without exposing sensitive information

### Resource Management
- Automatic cleanup of temporary files and directories
- Thread-safe resource tracking for concurrent operations
- Atomic file operations to prevent data corruption
- No resource leaks in any error scenario
- Comprehensive cleanup testing and verification

## Configuration Discovery
- Configuration files are discovered using a configurable search path
- The search path is controlled by the `BKPFILE_CONFIG` environment variable
- If `BKPFILE_CONFIG` is not set, the default search path is hard-coded as: `./.bkpfile.yml:~/.bkpfile.yml`
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
1. **Backup Directory Path**
   - Specifies where backups are stored
   - Default: `../.bkpfile` relative to current directory
   - YAML key: `backup_dir_path`
   - Backups maintain the source file's directory structure in the backup path

2. **Use Current Directory Name**
   - Controls whether to include current directory name in the backup path
   - Default: `true`
   - YAML key: `use_current_dir_name`
   - Example: With file 'cmd/bkpfile/main.go', backup path becomes '../.bkpfile/cmd/bkpfile/main.go-2025-05-12-13-49'

3. **Status Code Configuration**
   - Configures exit status codes returned for different application conditions
   - Status codes have specific defaults if not specified (see [Immutable Specifications](immutable.md#configuration-defaults))
   - YAML keys for status codes:
     - `status_created_backup`: Exit code when a new backup is successfully created (default: 0)
     - `status_failed_to_create_backup_directory`: Exit code when backup directory creation fails (default: 31)
     - `status_file_is_identical_to_existing_backup`: Exit code when file is identical to most recent backup (default: 0)
     - `status_file_not_found`: Exit code when source file does not exist (default: 20)
     - `status_invalid_file_type`: Exit code when source file is not a regular file (default: 21)
     - `status_permission_denied`: Exit code when file access is denied (default: 22)
     - `status_disk_full`: Exit code when disk space is insufficient (default: 30)
     - `status_config_error`: Exit code when configuration is invalid (default: 10)
   - Example configuration:
     ```yaml
     status_created_backup: 0
     status_failed_to_create_backup_directory: 31
     status_file_is_identical_to_existing_backup: 0
     status_file_not_found: 20
     status_invalid_file_type: 21
     status_permission_denied: 22
     status_disk_full: 30
     status_config_error: 10
     ```

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
- Handles errors gracefully with appropriate status codes

### 2. Display Configuration
- Displays computed configuration values after processing configuration files
- Usage: `bkpfile --config`
- Shows each configuration value with its name, computed value (including defaults), and source file
- Example output format:
  ```
  backup_dir_path: ../.bkpfile (source: default)
  use_current_dir_name: true (source: ~/.bkpfile.yml)
  config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)
  ```
- The application exits after displaying the configuration values
- Configuration files are processed from the `BKPFILE_CONFIG` environment variable path list
- If `BKPFILE_CONFIG` is not set, uses the default search path

### 3. Create Backup
- Creates a copy of the specified file with robust error handling and resource cleanup
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

#### Enhanced Backup Features
- **Atomic Operations**: Uses temporary files to ensure backup integrity
- **Resource Cleanup**: Automatically cleans up temporary files on success or failure
- **Context Support**: Supports operation cancellation and timeouts
- **Enhanced Error Detection**: Detects various disk space and permission conditions
- **Panic Recovery**: Recovers from unexpected errors without leaving temporary files
- **Thread Safety**: Safe for concurrent operations

## Global Options
- **Dry-Run Mode**: When enabled with `--dry-run` flag:
  - Shows the backup filename that would be created
  - No actual backup is created
  - Includes resource cleanup verification in dry-run mode

## Error Handling and Recovery

### Structured Error Reporting
- All operations return structured errors with specific status codes
- Human-readable error messages for common scenarios
- Technical details logged to stderr when appropriate
- No sensitive information exposed in error messages

### Enhanced Error Detection
- **Disk Space**: Detects various disk full conditions including:
  - "no space left on device"
  - "disk full"
  - "not enough space"
  - "insufficient disk space"
  - "device full"
  - "quota exceeded"
  - "file too large"
- **Permission Errors**: Proper handling of file access permissions
- **File Type Validation**: Ensures only regular files are backed up
- **Path Resolution**: Handles both absolute and relative paths securely

### Panic Recovery
- Critical operations include panic recovery mechanisms
- Panics are logged to stderr without exposing internal details
- Resource cleanup still occurs even when panics happen
- Application doesn't crash on unexpected errors

### Context and Cancellation Support
- Long-running operations support cancellation via context
- Timeout handling for operations that might hang
- Graceful shutdown with proper resource cleanup
- Context cancellation checked at multiple operation points

## Resource Management

### Automatic Cleanup
- All temporary files and directories are automatically cleaned up
- Cleanup occurs on success, failure, and cancellation
- Thread-safe resource tracking for concurrent operations
- Error-resilient cleanup continues even if individual operations fail

### Atomic Operations
- File operations use temporary files to ensure atomicity
- Atomic rename operations prevent data corruption
- Temporary files are registered for automatic cleanup
- Successful operations remove files from cleanup lists

### Leak Prevention
- Comprehensive testing verifies no resource leaks
- All temporary resources are tracked and cleaned
- No orphaned files remain in any error scenario
- Memory usage is properly managed

## Build and Development Requirements

### Code Quality Standards
- All code must pass `revive` linter before commit
- Comprehensive test coverage required for all features
- All errors must be properly handled
- Documentation required for all public functions
- Backward compatibility must be maintained

### Build System
- `make lint`: Run code linting
- `make test`: Run all tests with verbose output
- `make build`: Build application (depends on lint and test passing)
- `make clean`: Remove build artifacts

### Testing Requirements
- Unit tests for all core functions
- Integration tests for complete workflows
- Resource cleanup verification in all test scenarios
- Context cancellation and timeout testing
- Performance benchmarks for critical operations
- Stress testing for concurrent operations

## Implementation Details
For detailed implementation requirements and constraints, see:
- [Immutable Specifications](immutable.md) for core behaviors that cannot be changed
- [Architecture](architecture.md) for system design and implementation details
- [Requirements](requirements.md) for technical requirements and test coverage
- [Resource Cleanup Documentation](../RESOURCE_CLEANUP.md) for detailed cleanup functionality

## Platform Compatibility
- Works on macOS and Linux systems
- Uses platform-independent path handling
- Preserves file permissions and ownership where applicable
- Handles file system differences between platforms
- Thread-safe operations for concurrent access
- Efficient resource management across platforms

## Performance Characteristics
- Minimal overhead for resource tracking
- Efficient file comparison with length checks first
- Optimized cleanup operations
- Low memory footprint
- Fast atomic file operations
- Scalable for large files and many backups

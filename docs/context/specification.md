# BkpFile: Single File Backup CLI Application

## Overview
BkpFile is a command-line application for macOS and Linux that creates backups of individual files. It supports customizable naming patterns, maintains a history of file backups, and provides robust error handling with automatic resource cleanup. It also features configurable printf-style output formatting for enhanced user experience.

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

4. **Output Format Configuration**
   - Configures printf-style format strings for all standard output
   - Configures template-based formatting with named placeholders and regex patterns
   - Format strings support text highlighting and structure formatting
   - Templates support both Go text/template syntax ({{.name}}) and placeholder syntax (%{name})
   - All user-facing text is extracted from code into configuration data
   - Format strings have specific defaults if not specified (see [Immutable Specifications](immutable.md#configuration-defaults))
   - YAML keys for printf-style format strings:
     - `format_created_backup`: Format for successful backup creation messages (default: "Created backup: %s\n")
     - `format_identical_backup`: Format for identical file messages (default: "File is identical to existing backup: %s\n")
     - `format_list_backup`: Format for backup listing entries (default: "%s (created: %s)\n")
     - `format_config_value`: Format for configuration value display (default: "%s: %s (source: %s)\n")
     - `format_dry_run_backup`: Format for dry-run backup messages (default: "Would create backup: %s\n")
     - `format_error`: Format for error messages (default: "Error: %s\n")
   - YAML keys for template-based format strings:
     - `template_created_backup`: Template for successful backup creation messages (default: "Created backup: %{path}\n")
     - `template_identical_backup`: Template for identical file messages (default: "File is identical to existing backup: %{path}\n")
     - `template_list_backup`: Template for backup listing entries (default: "%{path} (created: %{creation_time})\n")
     - `template_config_value`: Template for configuration value display (default: "%{name}: %{value} (source: %{source})\n")
     - `template_dry_run_backup`: Template for dry-run backup messages (default: "Would create backup: %{path}\n")
     - `template_error`: Template for error messages (default: "Error: %{message}\n")
   - YAML keys for regex patterns:
     - `pattern_backup_filename`: Named regex for parsing backup filenames (default: `(?P<filename>[^/]+)-(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})-(?P<hour>\d{2})-(?P<minute>\d{2})(?:=(?P<note>.+))?`)
     - `pattern_config_line`: Named regex for parsing configuration display lines (default: `(?P<name>[^:]+):\s*(?P<value>[^(]+)\s*\(source:\s*(?P<source>[^)]+)\)`)
     - `pattern_timestamp`: Named regex for parsing timestamps (default: `(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})\s+(?P<hour>\d{2}):(?P<minute>\d{2}):(?P<second>\d{2})`)
   - Format strings support ANSI color codes and text formatting for enhanced readability
   - Template strings support conditional formatting and advanced text processing
   - Example configuration:
     ```yaml
     # Printf-style formatting
     format_created_backup: "\033[32m✓ Created backup:\033[0m %s\n"
     format_identical_backup: "\033[33m≡ File is identical to existing backup:\033[0m %s\n"
     format_list_backup: "\033[36m%s\033[0m (created: \033[90m%s\033[0m)\n"
     format_config_value: "\033[1m%s:\033[0m %s \033[90m(source: %s)\033[0m\n"
     format_dry_run_backup: "\033[35m⚠ Would create backup:\033[0m %s\n"
     format_error: "\033[31m✗ Error:\033[0m %s\n"
     
     # Template-based formatting with named placeholders
     template_created_backup: "\033[32m✓ Created backup:\033[0m %{path}\n"
     template_identical_backup: "\033[33m≡ File %{filename} is identical to backup from %{year}-%{month}-%{day}:\033[0m %{path}\n"
     template_list_backup: "\033[36m%{path}\033[0m (created: \033[90m%{creation_time}\033[0m) %{note}\n"
     template_config_value: "\033[1m%{name}:\033[0m %{value} \033[90m(from %{source})\033[0m\n"
     template_dry_run_backup: "\033[35m⚠ Would create backup for %{filename} on %{year}-%{month}-%{day}:\033[0m %{path}\n"
     template_error: "\033[31m✗ Error in %{operation}:\033[0m %{message}\n"
     
     # Named regex patterns for data extraction
     pattern_backup_filename: "(?P<filename>[^/]+)-(?P<year>\\d{4})-(?P<month>\\d{2})-(?P<day>\\d{2})-(?P<hour>\\d{2})-(?P<minute>\\d{2})(?:=(?P<note>.+))?"
     pattern_timestamp: "(?P<year>\\d{4})-(?P<month>\\d{2})-(?P<day>\\d{2})\\s+(?P<hour>\\d{2}):(?P<minute>\\d{2}):(?P<second>\\d{2})"
     ```

## Commands

### 1. List Backups
- Displays all backups associated with the specified file
- Usage: `bkpfile --list [FILE_PATH]`
- Shows each backup with its path (relative to current directory) and creation time using configurable format string or template
- Default format: `.bkpfile/path/to/file.txt-2024-03-21-15-30=note (created: 2024-03-21 15:30:00)`
- Output formatting uses `format_list_backup` configuration setting with printf-style specifications
- Alternative template formatting uses `template_list_backup` with named placeholders and `pattern_backup_filename` for data extraction
- Supports text highlighting and color formatting through ANSI escape codes in format strings and templates
- Template-based formatting allows rich data extraction from backup filenames using named regex groups
- Backups are sorted by creation time (most recent first)
- Backups are organized by their source file paths
- Handles errors gracefully with appropriate status codes using `format_error` or `template_error` configuration

### 2. Display Configuration
- Displays computed configuration values after processing configuration files
- Usage: `bkpfile --config`
- Shows each configuration value with its name, computed value (including defaults), and source file
- Output formatting uses `format_config_value` configuration setting with printf-style specifications
- Alternative template formatting uses `template_config_value` with named placeholders for enhanced display
- Default format: `backup_dir_path: ../.bkpfile (source: default)`
- Supports text highlighting and color formatting for enhanced readability
- Template-based formatting allows conditional formatting based on configuration source and value types
- The application exits after displaying the configuration values
- Configuration files are processed from the `BKPFILE_CONFIG` environment variable path list
- If `BKPFILE_CONFIG` is not set, uses the default search path
- Includes display of all format string configurations, template configurations, and regex patterns with their current values and sources

### 3. Create Backup
- Creates a copy of the specified file with robust error handling and resource cleanup
- Usage: `bkpfile [FILE_PATH] [NOTE]`
- Before creating a backup:
  - Compares the file with its most recent backup using byte comparison
  - If the file is identical to the most recent backup:
    - Reports the existing backup path using `format_identical_backup` or `template_identical_backup` configuration
    - Template formatting can extract and display rich information from backup filename using `pattern_backup_filename`
    - Default format: "File is identical to existing backup: [PATH]"
    - Template format can show: "File [filename] is identical to backup from [year]-[month]-[day]: [path]"
    - Exits normally without creating a new backup
- When a new backup is created:
  - Reports success using `format_created_backup` or `template_created_backup` configuration
  - Default format: "Created backup: [PATH]"
  - Template format can include extracted filename and timestamp information
- Backup naming format: `SOURCE_FILENAME-YYYY-MM-DD-hh-mm[=NOTE]`
  - SOURCE_FILENAME is the base name of the original file
  - YYYY-MM-DD-hh-mm is the timestamp of the backup
  - NOTE is an optional note appended with an equals sign
- The backup maintains the original file's directory structure in the backup path
- NOTE is an optional positional argument provided by the user
- All output uses configurable printf-style format strings or template-based formatting for consistency and customization
- Error messages use `format_error` or `template_error` configuration setting with optional operation context

#### Enhanced Backup Features
- **Atomic Operations**: Uses temporary files to ensure backup integrity
- **Resource Cleanup**: Automatically cleans up temporary files on success or failure
- **Context Support**: Supports operation cancellation and timeouts
- **Enhanced Error Detection**: Detects various disk space and permission conditions
- **Panic Recovery**: Recovers from unexpected errors without leaving temporary files
- **Thread Safety**: Safe for concurrent operations
- **Configurable Output**: All user-facing messages use printf-style format strings

## Global Options
- **Dry-Run Mode**: When enabled with `--dry-run` flag:
  - Shows the backup filename that would be created using `format_dry_run_backup` or `template_dry_run_backup` configuration
  - Default format: "Would create backup: [PATH]"
  - Template format can show: "Would create backup for [filename] on [year]-[month]-[day]: [path]"
  - Template-based formatting can extract and display rich information about the planned backup
  - No actual backup is created
  - Includes resource cleanup verification in dry-run mode
  - All output uses configurable printf-style format strings or template-based formatting

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

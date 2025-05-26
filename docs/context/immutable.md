# Immutable Specifications

This document contains specifications that MUST NOT be changed without a major version bump. These are core behaviors that users and other systems depend on.

## Backup Naming Convention
- Format: `SOURCE_FILENAME-YYYY-MM-DD-hh-mm[=NOTE]`
- Maintain original file's directory structure in backup path
- Optional note can be appended with equals sign
- This naming convention is fixed and must not be modified

## File Operations
- Use platform-independent path handling
- Preserve file permissions and modification time in backups
- Handle both absolute and relative file paths
- Source files must be regular files (not directories or special files)
- Create backup directories automatically if they don't exist
- Display all paths relative to current directory
- **Atomic Operations**: All backup operations must be atomic to prevent corruption
- **Resource Cleanup**: All temporary files must be cleaned up automatically
- These file operation rules are fundamental and must not be altered

## Error Handling Requirements
- **Structured Errors**: All backup operations must return structured errors with status codes
- **No Resource Leaks**: No temporary files or directories may remain after any operation
- **Panic Recovery**: Application must recover from panics without leaving temporary resources
- **Context Support**: Long-running operations must support cancellation via context
- **Enhanced Detection**: Must detect various disk space and permission error conditions
- These error handling requirements are mandatory and must be preserved

## Code Quality Standards
- **Linting**: All Go code must pass `revive` linter checks
- **Error Handling**: All errors must be properly handled (no unhandled return values)
- **Testing**: All code must have comprehensive test coverage
- **Documentation**: All public functions must be documented
- **Backward Compatibility**: New features must not break existing functionality
- These quality standards are immutable and must be maintained

## Commands
1. List Backups:
   - Command: `bkpfile --list [FILE_PATH]`
   - Sort by creation time (most recent first)
   - Display format: `.bkpfile/path/to/file.txt-2024-03-21-15-30=note (created: 2024-03-21 15:30:00)`
   - This command structure and output format must be preserved

2. Display Configuration:
   - Command: `bkpfile --config`
   - Display computed configuration values with name, value, and source
   - Process configuration files from `BKPFILE_CONFIG` environment variable
   - Exit after displaying values
   - This command behavior must remain unchanged once implemented

3. Create Backup:
   - Command: `bkpfile [FILE_PATH] [NOTE]`
   - Compare with most recent backup before creating
   - Skip if identical to most recent backup
   - **Must use atomic operations with automatic cleanup**
   - **Must support context cancellation**
   - This backup creation logic must remain unchanged

## Configuration Defaults
- Configuration discovery uses `BKPFILE_CONFIG` environment variable to specify search path
- Default configuration search path is hard-coded as `./.bkpfile.yml:~/.bkpfile.yml` (if `BKPFILE_CONFIG` not set)
- Configuration files are processed in order with earlier files taking precedence
- Default backup directory: `../.bkpfile` relative to current directory
- Default use_current_dir_name: true
- Default status codes: All status codes default to `0` (success) if not specified
  - `status_config_error`: 10
  - `status_created_backup`: 0
  - `status_disk_full`: 30
  - `status_failed_to_create_backup_directory`: 31
  - `status_file_is_identical_to_existing_backup`: 0
  - `status_file_not_found`: 20
  - `status_invalid_file_type`: 21
  - `status_permission_denied`: 22
- These configuration defaults must never be changed without explicit user override

## Platform Compatibility
- Support macOS and Linux systems
- Handle platform-specific file system differences
- Preserve file permissions and ownership where applicable
- **Thread-safe operations for concurrent access**
- **Efficient resource management across platforms**
- Platform support must never be reduced or modified

## Global Options
- Support `--dry-run` flag for previewing backup operations
- **Dry-run must include resource cleanup verification**
- Existing flag behavior must be maintained

## Build System Requirements
- **Linting**: `make lint` must pass before any code commit
- **Testing**: `make test` must pass with comprehensive coverage
- **Building**: `make build` must depend on successful linting and testing
- **Cleaning**: `make clean` must remove all build artifacts
- These build requirements are immutable and must be enforced

## Resource Management Requirements
- **Automatic Cleanup**: All temporary resources must be cleaned up automatically
- **Thread Safety**: Resource management must be thread-safe
- **Atomic Operations**: File operations must use temporary files for atomicity
- **Leak Prevention**: No resource leaks allowed in any scenario
- **Error Resilience**: Cleanup must continue even if individual operations fail
- These resource management requirements are mandatory and cannot be relaxed

## Performance Requirements
- **Minimal Overhead**: Resource tracking must have minimal performance impact
- **Efficient Operations**: File comparison must check length before byte comparison
- **Scalability**: Must handle large files and many backups efficiently
- **Memory Management**: Must maintain low memory footprint
- These performance characteristics must be preserved

## Feature Preservation Rules
1. New Features:
   - Must not interfere with existing functionality
   - Must maintain all current behaviors
   - Must be optional and not affect existing workflows
   - **Must include automatic resource cleanup**
   - **Must support context cancellation where appropriate**
   - **Must pass all linting and testing requirements**

2. Modifications:
   - Must preserve all existing command-line interfaces
   - Must maintain current file handling behaviors
   - Must keep existing configuration options
   - Must not change established backup naming patterns
   - **Must not introduce resource leaks**
   - **Must maintain error handling standards**
   - **Must preserve atomic operation guarantees**

3. Testing Requirements:
   - All new code must include tests for existing functionality
   - Regression tests must verify no existing features are broken
   - Platform compatibility tests must be maintained
   - **Resource cleanup must be verified in all test scenarios**
   - **Context cancellation and timeout handling must be tested**
   - **Performance benchmarks must not regress**
   - **All code must pass linting before commit**

4. Quality Assurance:
   - **Code must pass revive linter with zero warnings**
   - **All errors must be properly handled**
   - **All public functions must be documented**
   - **Test coverage must meet minimum thresholds**
   - **No temporary files may remain after any operation**
   - **Memory leaks are strictly prohibited** 
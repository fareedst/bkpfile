# Architecture

This document describes the system architecture and design of the BkpFile application.

> **Note**: For testing architecture and requirements, see [Testing Requirements](testing.md).

## System Architecture Overview

The BkpFile application follows a layered architecture with clear separation of concerns:

1. **CLI Layer**: Command-line interface and user interaction
2. **Business Logic Layer**: Core backup functionality and workflows
3. **Infrastructure Layer**: File system operations, configuration, and resource management
4. **Quality Assurance Layer**: Linting, testing, and code quality enforcement

## Data Objects

1. **Config**
   - `Config`: string - Colon-separated list of configuration file paths to search
   - `BackupDirPath`: string - Path where backups are stored
   - `UseCurrentDirName`: bool - Whether to use current directory name in backup path
   - `StatusCreatedBackup`: int - Exit code when a new backup is successfully created
   - `StatusFailedToCreateBackupDirectory`: int - Exit code when backup directory creation fails
   - `StatusFileIsIdenticalToExistingBackup`: int - Exit code when file is identical to most recent backup
   - `StatusFileNotFound`: int - Exit code when source file does not exist
   - `StatusInvalidFileType`: int - Exit code when source file is not a regular file
   - `StatusPermissionDenied`: int - Exit code when file access is denied
   - `StatusDiskFull`: int - Exit code when disk space is insufficient
   - `StatusConfigError`: int - Exit code when configuration is invalid

2. **ConfigValue**
   - `Name`: string - Configuration parameter name
   - `Value`: string - Computed configuration value including defaults
   - `Source`: string - Source file path or "default" for default values

3. **Backup**
   - `Name`: string - Name of the backup file
   - `Path`: string - Full path to the backup file
   - `CreationTime`: time.Time - When the backup was created
   - `SourceFile`: string - Path to the original file
   - `Note`: string - Optional note for the backup

4. **BackupError**
   - `Message`: string - Human-readable error description
   - `StatusCode`: int - Numeric exit code for application
   - Implements Go's `error` interface for structured error handling
   - Provides consistent error reporting across all backup operations

5. **ResourceManager**
   - `tempFiles`: []string - List of temporary files to clean up
   - `tempDirs`: []string - List of temporary directories to clean up
   - `mutex`: sync.Mutex - Mutex for thread-safe access
   - Provides automatic resource cleanup and leak prevention
   - Thread-safe resource tracking for concurrent operations

## Core Functions

1. **Configuration Management**
   - `DefaultConfig() *Config`: Creates default configuration
   - `LoadConfig(root string) (*Config, error)`: Loads config from YAML files using discovery path or uses defaults
     - Reads `BKPFILE_CONFIG` environment variable for configuration search path
     - Processes multiple configuration files with precedence rules
     - Supports home directory expansion and path resolution
   - `GetConfigSearchPath() []string`: Returns list of configuration file paths to search
     - Reads `BKPFILE_CONFIG` environment variable
     - Returns default path if environment variable not set
     - Handles colon-separated path list parsing
   - `DisplayConfig() error`: Displays computed configuration values and exits
     - Processes configuration files from `BKPFILE_CONFIG` environment variable
     - Shows each configuration value with name, computed value, and source file
     - Displays format: `name: value (source: source_file)`
     - Application exits after displaying values

2. **File System Operations**
   - `CopyFile(src, dst string) error`: Creates an exact copy of the specified file
     - Preserves file permissions and modification time
     - Creates destination directories if needed
     - Handles both absolute and relative paths
   - `CopyFileWithContext(ctx context.Context, src, dst string) error`: Context-aware file copying
     - Same functionality as CopyFile with cancellation support
     - Checks for context cancellation at multiple points
     - Returns appropriate error on cancellation

3. **Enhanced Error Handling**
   - `NewBackupError(message string, statusCode int) *BackupError`: Creates structured errors
   - `isDiskFullError(err error) bool`: Enhanced disk space error detection
     - Detects multiple disk space indicators
     - Case-insensitive error message matching
     - Supports various disk full error patterns
   - Structured error handling with status codes throughout the application
   - Panic recovery mechanisms in critical operations

4. **Resource Management**
   - `NewResourceManager() *ResourceManager`: Creates new resource manager
   - `AddTempFile(path string)`: Registers temporary file for cleanup
   - `AddTempDir(path string)`: Registers temporary directory for cleanup
   - `Cleanup()`: Removes all registered resources
   - Thread-safe resource tracking with mutex protection
   - Error-resilient cleanup that continues even if individual operations fail

5. **Backup Management**
   - `GenerateBackupName(sourcePath, timestamp, note string) string`: Generates backup filename
     - Uses base filename from source path
     - Adds timestamp in YYYY-MM-DD-hh-mm format
     - Appends note with equals sign if provided
   - `ListBackups(backupDir string, sourceFile string) ([]Backup, error)`: Gets all backups for a specific file
     - Handles both absolute and relative paths
     - Sorts backups by creation time (most recent first)
     - Extracts notes from backup filenames
   - `CreateBackup(cfg *Config, filePath string, note string, dryRun bool) error`: Creates a backup of the specified file
     - Validates source file exists and is regular
     - Handles both absolute and relative paths
     - Creates backup directories if needed
     - Includes panic recovery for unexpected errors
   - `CreateBackupWithCleanup(cfg *Config, filePath string, note string, dryRun bool) error`: Backup with resource cleanup
     - All CreateBackup functionality plus automatic resource cleanup
     - Atomic operations using temporary files
     - No resource leaks on errors or panics
   - `CreateBackupWithContext(ctx context.Context, cfg *Config, filePath string, note string, dryRun bool) error`: Context-aware backup
     - All CreateBackup functionality plus cancellation support
     - Context cancellation checks at multiple points
     - Proper error handling on cancellation
   - `CreateBackupWithContextAndCleanup(ctx context.Context, cfg *Config, filePath string, note string, dryRun bool) error`: Most robust backup
     - Combines context support with resource cleanup
     - Atomic operations with cleanup on cancellation
     - Most reliable backup creation function
   - `CreateBackupWithTime(cfg *Config, filePath string, note string, dryRun bool, now func() time.Time) error`: Test helper for creating backups with custom time
   - `CompareFiles(file1, file2 string) (bool, error)`: Byte-by-byte file comparison
     - Efficient comparison with length check first
     - Handles various file types and sizes

## Main Application Structure

1. **CLI Interface**
   - Uses `cobra` for command-line interface
   - Global flags:
     - `--dry-run`: Show what would be done without creating backups
     - `--list`: List backups for the specified file
     - `--config`: Display computed configuration values and exit
   - Default behavior:
     - Creates backup of specified file with optional note
     - Usage: `bkpfile [FILE_PATH] [NOTE]`
     - Displays paths relative to current directory
     - When a new backup is created: Displays "Created backup: [PATH]"
     - When file is identical to existing backup: Displays "File is identical to existing backup: [PATH]"

2. **Workflow Implementation**
   - For creating backup:
     1. Load config
     2. Validate source file exists and is regular
     3. Convert file path to relative path if needed
     4. Compare file with most recent backup
        - If identical, report existing backup name and exit with `cfg.StatusFileIsIdenticalToExistingBackup`
        - If different, proceed with backup creation
     5. Generate backup name using base filename
     6. Create backup directory structure
     7. Create file copy (or simulate in dry-run) with atomic operations
     8. Clean up temporary resources
     9. Exit with `cfg.StatusCreatedBackup` on successful backup creation
     - Error handling uses configured status codes:
       - File not found: exit with `cfg.StatusFileNotFound`
       - Invalid file type: exit with `cfg.StatusInvalidFileType`
       - Permission denied: exit with `cfg.StatusPermissionDenied`
       - Disk full: exit with `cfg.StatusDiskFull`
       - Failed to create backup directory: exit with `cfg.StatusFailedToCreateBackupDirectory`
       - Configuration error: exit with `cfg.StatusConfigError`
     - Enhanced error handling:
       - Panic recovery with proper logging
       - Context cancellation support
       - Automatic resource cleanup on all error paths

   - For listing backups:
     1. Load config
     2. Convert source path to relative path if needed
     3. Find backup directory for the file
     4. List and filter backup files
     5. Extract backup information and notes
     6. Sort backups by creation time
     7. Display backup information

   - For displaying configuration:
     1. Read `BKPFILE_CONFIG` environment variable or use default search path
     2. Process configuration files in order with precedence rules
     3. Merge configuration values with defaults
     4. Track source file for each configuration value
     5. Display each configuration value with name, computed value, and source
     6. Exit application after display

3. **Utility Functions**
   - Backup naming follows format: `filename-YYYY-MM-DD-hh-mm[=note]`
   - File copy ensures all permissions are preserved
   - Source path structure is preserved in backup directory
   - Handles both absolute and relative paths consistently
   - Validates file types and existence
   - File comparison uses byte-by-byte comparison
   - Resource management ensures no temporary files remain
   - Enhanced error detection for various failure scenarios

## Quality Assurance Architecture

1. **Linting Infrastructure**
   - **Tool**: `revive` linter for Go code quality
   - **Configuration**: `.revive.toml` file with custom rules
   - **Integration**: `make lint` command in build system
   - **Standards**: Enforces Go best practices and error handling compliance
   - **Requirements**: All code must pass linting before commit

2. **Build System**
   - **Makefile**: Orchestrates build, test, and quality checks
   - **Commands**:
     - `make lint`: Run revive linter
     - `make test`: Run all tests with verbose output
     - `make build`: Build application (depends on lint and test)
     - `make clean`: Remove build artifacts
   - **Dependencies**: Build depends on successful linting and testing

3. **Testing Architecture**
   - **Unit Tests**: Test individual functions and components
   - **Integration Tests**: Test complete workflows and CLI interface
   - **Resource Cleanup Tests**: Verify no temporary files remain
   - **Context Tests**: Validate cancellation and timeout handling
   - **Performance Tests**: Benchmark critical operations
   - **Stress Tests**: Test under load and concurrent access

## Error Handling Architecture

1. **Structured Errors**
   - `BackupError` provides consistent error reporting
   - Status codes configurable via YAML configuration
   - Human-readable messages with technical details
   - Proper error wrapping and context preservation

2. **Panic Recovery**
   - Critical operations include panic recovery
   - Panics logged to stderr without exposing internals
   - Resource cleanup still occurs on panic
   - Application doesn't crash on unexpected errors

3. **Context Support**
   - Operations support cancellation via context
   - Timeout handling for long-running operations
   - Graceful shutdown with proper cleanup
   - Context cancellation checked at multiple points

## Resource Management Architecture

1. **Automatic Cleanup**
   - `ResourceManager` tracks all temporary resources
   - Cleanup occurs via defer mechanisms
   - Thread-safe resource tracking with mutex
   - Error-resilient cleanup continues on individual failures

2. **Atomic Operations**
   - File operations use temporary files for atomicity
   - Atomic rename prevents corruption
   - Temporary files registered for cleanup
   - Success removes files from cleanup list

3. **Leak Prevention**
   - All temporary resources tracked and cleaned
   - Cleanup occurs on success, failure, and cancellation
   - No resource leaks in any scenario
   - Comprehensive testing verifies cleanup

## Concurrency Architecture

1. **Thread Safety**
   - `ResourceManager` uses mutex for thread-safe access
   - Configuration loading is thread-safe
   - File operations handle concurrent access appropriately

2. **Context Propagation**
   - Context passed through operation chains
   - Cancellation signals propagated correctly
   - Timeout handling at appropriate levels

## Performance Architecture

1. **Efficient Operations**
   - File comparison checks length before byte comparison
   - Minimal overhead for resource tracking
   - Efficient cleanup with batch operations

2. **Memory Management**
   - Proper resource cleanup prevents memory leaks
   - Efficient file operations for large files
   - Minimal memory overhead for tracking

## Security Architecture

1. **File Permissions**
   - Preserves original file permissions
   - Creates backup directories with appropriate permissions
   - Handles permission denied errors gracefully

2. **Path Handling**
   - Secure path resolution and validation
   - Prevents directory traversal attacks
   - Proper handling of symbolic links

3. **Error Information**
   - Doesn't expose sensitive information in error messages
   - Logs technical details to stderr when appropriate
   - User-friendly error messages for common scenarios

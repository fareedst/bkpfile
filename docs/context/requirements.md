# Architecture and Specification Requirements Traceability

This document maps code components to their corresponding architecture requirements and specification requirements.

> **Note**: For testing requirements and architecture, see [Testing Requirements](testing.md).

## Code Quality and Linting Requirements

### Linting Standards
**Implementation**: `Makefile`, `.revive.toml`
**Specification Requirements**:
- **Linter**: Uses `revive` for Go code linting
  - Spec: "All code must pass revive linter checks"
  - Configuration: `.revive.toml` file in project root
  - Command: `make lint` runs revive linter
  - Rules: Standard Go best practices with custom configurations
- **Error Handling**: All errors must be properly handled
  - Spec: "No unhandled errors allowed"
  - All `fmt.Printf`, `fmt.Fprintf` return values must be checked
  - All file operations must handle errors appropriately
- **Code Style**: Consistent formatting and naming conventions
  - Spec: "Follow Go standard formatting and naming"
  - Function names must be descriptive and follow Go conventions
  - Variable names must be clear and meaningful
  - Comments must follow Go documentation standards

**Example Usage**:
```bash
# Run linter
make lint

# Build with linting
make build
```

### Resource Management Requirements
**Implementation**: `backup.go` - `ResourceManager`
**Specification Requirements**:
- **Resource Cleanup**: All temporary resources must be cleaned up
  - Spec: "No temporary files or directories should remain after operations"
  - Implementation: `ResourceManager` struct with automatic cleanup
  - Thread-safe: Uses mutex for concurrent access
  - Error-resilient: Continues cleanup even if individual operations fail
- **Atomic Operations**: File operations must be atomic
  - Spec: "Backup creation must be atomic to prevent corruption"
  - Implementation: Temporary files with atomic rename operations
  - Cleanup: Temporary files registered for automatic cleanup
- **Panic Recovery**: Operations must recover from panics
  - Spec: "Unexpected panics must not leave resources uncleaned"
  - Implementation: Defer functions with panic recovery
  - Logging: Panic information logged to stderr

**Example Usage**:
```go
// Create resource manager
rm := NewResourceManager()
defer rm.Cleanup()

// Register temporary resources
rm.AddTempFile("/path/to/temp.tmp")
rm.AddTempDir("/path/to/tempdir")
```

### Enhanced Error Handling Requirements
**Implementation**: `backup.go` - Enhanced error handling
**Specification Requirements**:
- **Structured Errors**: Use `BackupError` for consistent error handling
  - Spec: "All backup operations must return structured errors with status codes"
  - Fields:
    - `Message`: Human-readable error description
    - `StatusCode`: Numeric exit code for application
  - Implementation: Implements Go's `error` interface
  - Usage: Allows callers to extract both message and status code

**Example Usage**:
```go
// Create structured error
err := NewBackupError("file not found", cfg.StatusFileNotFound)

// Check for BackupError type
if backupErr, ok := err.(*BackupError); ok {
    fmt.Fprintf(os.Stderr, "Error: %s\n", backupErr.Message)
    os.Exit(backupErr.StatusCode)
}
```

## Data Objects

### Config
**Implementation**: `config.go`
**Specification Requirements**:
- Configuration stored in YAML file `.bkpfile.yml` at root directory
- Default values used if file not present
- Fields:
  - `BackupDirPath`: `Config.BackupDirPath`
    - Spec: "Specifies where backups are stored"
    - Default: "../.bkpfile" relative to current directory
    - YAML key: "backup_dir_path"
  - `UseCurrentDirName`: `Config.UseCurrentDirName`
    - Spec: "Controls whether to include current directory name in backup path"
    - Default: true
    - YAML key: "use_current_dir_name"
  - `StatusCreatedBackup`: `Config.StatusCreatedBackup`
    - Spec: "Exit code when a new backup is successfully created"
    - Default: 0
    - YAML key: "status_created_backup"
  - `StatusFailedToCreateBackupDirectory`: `Config.StatusFailedToCreateBackupDirectory`
    - Spec: "Exit code when backup directory creation fails"
    - Default: 31
    - YAML key: "status_failed_to_create_backup_directory"
  - `StatusFileIsIdenticalToExistingBackup`: `Config.StatusFileIsIdenticalToExistingBackup`
    - Spec: "Exit code when file is identical to most recent backup"
    - Default: 0
    - YAML key: "status_file_is_identical_to_existing_backup"
  - `StatusFileNotFound`: `Config.StatusFileNotFound`
    - Spec: "Exit code when source file does not exist"
    - Default: 20
    - YAML key: "status_file_not_found"
  - `StatusInvalidFileType`: `Config.StatusInvalidFileType`
    - Spec: "Exit code when source file is not a regular file"
    - Default: 21
    - YAML key: "status_invalid_file_type"
  - `StatusPermissionDenied`: `Config.StatusPermissionDenied`
    - Spec: "Exit code when file access is denied"
    - Default: 22
    - YAML key: "status_permission_denied"
  - `StatusDiskFull`: `Config.StatusDiskFull`
    - Spec: "Exit code when disk space is insufficient"
    - Default: 30
    - YAML key: "status_disk_full"
  - `StatusConfigError`: `Config.StatusConfigError`
    - Spec: "Exit code when configuration is invalid"
    - Default: 10
    - YAML key: "status_config_error"

**Example Usage**:
```go
// Load configuration from YAML or use defaults
cfg, err := LoadConfig(".")
if err != nil {
    log.Fatal(err)
}

// Access configuration values
backupPath := cfg.BackupDirPath
useCurrentDir := cfg.UseCurrentDirName

// Access status code configuration
createdBackupStatus := cfg.StatusCreatedBackup
identicalFileStatus := cfg.StatusFileIsIdenticalToExistingBackup
fileNotFoundStatus := cfg.StatusFileNotFound
```

### ConfigValue
**Implementation**: `config.go`
**Specification Requirements**:
- Fields:
  - `Name`: `ConfigValue.Name`
    - Spec: "Configuration parameter name"
  - `Value`: `ConfigValue.Value`
    - Spec: "Computed configuration value including defaults"
  - `Source`: `ConfigValue.Source`
    - Spec: "Source file path or 'default' for default values"

**Example Usage**:
```go
// Create configuration value entry
configValue := &ConfigValue{
    Name:   "backup_dir_path",
    Value:  "../.bkpfile",
    Source: "~/.bkpfile.yml",
}
```

### Backup
**Implementation**: `backup.go`
**Specification Requirements**:
- Fields:
  - `Name`: `Backup.Name`
    - Spec: "Backup filename in format: filename-YYYY-MM-DD-hh-mm[=note]"
  - `Path`: `Backup.Path`
    - Spec: "Full path to backup file"
  - `CreationTime`: `Backup.CreationTime`
    - Spec: "When the backup was created"
  - `SourceFile`: `Backup.SourceFile`
    - Spec: "Path to the original file"
  - `Note`: `Backup.Note`
    - Spec: "Optional note for the backup"

**Example Usage**:
```go
// Create a new backup object
backup := &Backup{
    Name: "file.txt-2024-03-20-15-30=important_backup",
    Path: "/path/to/backup/file.txt-2024-03-20-15-30=important_backup",
    CreationTime: time.Now(),
    SourceFile: "/path/to/original/file.txt",
    Note: "important_backup",
}
```

### BackupError
**Implementation**: `backup.go`
**Specification Requirements**:
- **Structured Error Handling**: Provides consistent error reporting
  - Spec: "All backup operations return structured errors with status codes"
  - Fields:
    - `Message`: Human-readable error description
    - `StatusCode`: Numeric exit code for application
  - Implementation: Implements Go's `error` interface
  - Usage: Allows callers to extract both message and status code

**Example Usage**:
```go
// Create structured error
err := NewBackupError("file not found", cfg.StatusFileNotFound)

// Check for BackupError type
if backupErr, ok := err.(*BackupError); ok {
    fmt.Fprintf(os.Stderr, "Error: %s\n", backupErr.Message)
    os.Exit(backupErr.StatusCode)
}
```

### ResourceManager
**Implementation**: `backup.go`
**Specification Requirements**:
- **Resource Tracking**: Thread-safe tracking of temporary resources
  - Spec: "Track all temporary files and directories for cleanup"
  - Fields:
    - `tempFiles`: List of temporary files to clean up
    - `tempDirs`: List of temporary directories to clean up
    - `mutex`: Mutex for thread-safe access
  - Methods:
    - `AddTempFile()`: Register temporary file for cleanup
    - `AddTempDir()`: Register temporary directory for cleanup
    - `Cleanup()`: Remove all registered resources

**Example Usage**:
```go
// Create and use resource manager
rm := NewResourceManager()
defer rm.Cleanup()

rm.AddTempFile("/tmp/backup.tmp")
rm.AddTempDir("/tmp/backup_work")
```

## Core Functions

### Configuration Management
**Implementation**: `config.go`
**Specification Requirements**:
- `DefaultConfig()`: `DefaultConfig()`
  - Spec: "Creates default configuration with specified defaults"
  - Input: None
  - Output: `*Config` - Returns a new Config with default values
  - Behavior: Creates a new Config struct with all fields set to their default values
  - Default Values:
    - BackupDirPath: "../.bkpfile"
    - UseCurrentDirName: true
    - StatusCreatedBackup: 0
    - StatusFailedToCreateBackupDirectory: 31
    - StatusFileIsIdenticalToExistingBackup: 0
    - StatusFileNotFound: 20
    - StatusInvalidFileType: 21
    - StatusPermissionDenied: 22
    - StatusDiskFull: 30
    - StatusConfigError: 10

- `LoadConfig()`: `LoadConfig(root string) (*Config, error)`
  - Spec: "Loads config from YAML or uses defaults"
  - Input: `root string` - Path to root directory containing .bkpfile.yml
  - Output: `(*Config, error)` - Returns config and any error encountered
  - Behavior:
    - Attempts to read .bkpfile.yml from root directory
    - If file exists, merges with default values
    - If file doesn't exist, returns default config
    - Returns error if file exists but is invalid YAML
  - Error Cases:
    - Invalid YAML format
    - Invalid configuration values
    - File system errors

- `DisplayConfig()`: `DisplayConfig() error`
  - Spec: "Displays computed configuration values and exits"
  - Input: None
  - Output: `error` - Any error encountered
  - Behavior:
    - Processes configuration files from `BKPFILE_CONFIG` environment variable
    - If `BKPFILE_CONFIG` not set, uses hard-coded default search path: `./.bkpfile.yml:~/.bkpfile.yml`
    - Shows each configuration value with name, computed value, and source file
    - Displays format: `name: value (source: source_file)`
    - Default values show source as "default"
    - Application exits after displaying values
  - Error Cases:
    - Configuration file read errors
    - Invalid YAML format
    - Environment variable parsing errors

**Example Usage**:
```go
// Create default configuration
cfg := DefaultConfig()

// Load configuration from YAML file
cfg, err := LoadConfig(".")
if err != nil {
    log.Fatal(err)
}

// Display configuration values and exit
err = DisplayConfig()
if err != nil {
    log.Fatal(err)
}
```

### File System Operations
**Implementation**: `backup.go`
**Specification Requirements**:
- `CopyFile(src, dst string) error`
  - Spec: "Creates an exact copy of the specified file"
  - Input:
    - `src string` - Path to source file
    - `dst string` - Path to destination file
  - Output: `error` - Any error encountered
  - Behavior:
    - Creates destination directory if needed
    - Copies file with all permissions preserved
    - Preserves original file's modification time
    - Handles both absolute and relative paths
  - Error Cases:
    - Source file not found
    - Permission denied
    - Disk full
    - Other file system errors

- `CopyFileWithContext(ctx context.Context, src, dst string) error`
  - Spec: "Context-aware file copying with cancellation support"
  - Input:
    - `ctx context.Context` - Context for cancellation
    - `src string` - Path to source file
    - `dst string` - Path to destination file
  - Output: `error` - Any error encountered
  - Behavior:
    - Same as CopyFile but with context cancellation checks
    - Checks for cancellation at multiple points during operation
    - Returns context.Canceled if operation is cancelled
  - Error Cases:
    - All CopyFile error cases plus context cancellation

**Example Usage**:
```go
// Copy a file
err := CopyFile("source.txt", "backup.txt")
if err != nil {
    log.Fatal(err)
}

// Copy with context and timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err = CopyFileWithContext(ctx, "source.txt", "backup.txt")
if err == context.Canceled {
    log.Println("Copy operation was cancelled")
}
```

### Enhanced Error Detection
**Implementation**: `backup.go`
**Specification Requirements**:
- `isDiskFullError(err error) bool`
  - Spec: "Enhanced disk space error detection"
  - Input: `err error` - Error to check
  - Output: `bool` - True if error indicates disk space issues
  - Behavior:
    - Checks error message for multiple disk space indicators
    - Indicators: "no space left", "disk full", "not enough space", "insufficient disk space", "device full", "quota exceeded", "file too large"
    - Case-insensitive matching
  - Error Cases: None (returns false for nil or unrelated errors)

**Example Usage**:
```go
// Check for disk space errors
if err := someFileOperation(); err != nil {
    if isDiskFullError(err) {
        return NewBackupError("Disk full", cfg.StatusDiskFull)
    }
    return err
}
```

### Backup Management
**Implementation**: `backup.go`
**Specification Requirements**:
- `GenerateBackupName(sourcePath, timestamp, note string) string`
  - Spec: "Generates backup filename according to specified format"
  - Input:
    - `sourcePath string` - Path to source file
    - `timestamp string` - Creation timestamp
    - `note string` - Optional note for backup
  - Output: `string` - Generated backup filename
  - Behavior:
    - Uses base filename from source path
    - Adds timestamp in YYYY-MM-DD-hh-mm format
    - Appends note with equals sign if provided
  - Error Cases: None

- `ListBackups(backupDir string, sourceFile string) ([]Backup, error)`
  - Spec: "Gets all backups for a specific file"
  - Input:
    - `backupDir string` - Path to backup directory
    - `sourceFile string` - Original file path
  - Output: `([]Backup, error)` - List of backups and any error
  - Behavior:
    - Handles both absolute and relative paths
    - Scans directory for backup files
    - Creates Backup objects for each file matching the source file
    - Extracts notes from backup filenames
    - Sorts backups by creation time (most recent first)
  - Error Cases:
    - Directory not found
    - Permission denied
    - Invalid backup files

- `CreateBackup(cfg *Config, filePath string, note string, dryRun bool) error`
  - Spec: "Creates backup of specified file"
  - Input:
    - `cfg *Config` - Configuration to use
    - `filePath string` - Path to file to backup
    - `note string` - Optional note for backup
    - `dryRun bool` - Whether to simulate creation
  - Output: `error` - Any error encountered
  - Behavior:
    - Validates source file exists and is regular
    - Handles both absolute and relative paths
    - Creates backup directory if needed
    - Generates backup name using base filename
    - Copies file (or simulates copy in dry-run)
    - Uses configured status codes for different exit conditions
    - Includes panic recovery for unexpected errors
  - Error Cases:
    - Invalid configuration (exits with `cfg.StatusConfigError`)
    - File not found (exits with `cfg.StatusFileNotFound`)
    - File is not a regular file (exits with `cfg.StatusInvalidFileType`)
    - Permission denied (exits with `cfg.StatusPermissionDenied`)
    - Disk full (exits with `cfg.StatusDiskFull`)
    - Failed to create backup directory (exits with `cfg.StatusFailedToCreateBackupDirectory`)
    - File identical to existing backup (exits with `cfg.StatusFileIsIdenticalToExistingBackup`)
    - Successful backup creation (exits with `cfg.StatusCreatedBackup`)

- `CreateBackupWithCleanup(cfg *Config, filePath string, note string, dryRun bool) error`
  - Spec: "Creates backup with automatic resource cleanup"
  - Input: Same as CreateBackup
  - Output: `error` - Any error encountered
  - Behavior:
    - All CreateBackup functionality plus:
    - Automatic resource cleanup via ResourceManager
    - Atomic operations using temporary files
    - Cleanup on errors or panics
    - No temporary files left behind
  - Error Cases: Same as CreateBackup

- `CreateBackupWithContext(ctx context.Context, cfg *Config, filePath string, note string, dryRun bool) error`
  - Spec: "Context-aware backup creation"
  - Input:
    - `ctx context.Context` - Context for cancellation
    - Other inputs same as CreateBackup
  - Output: `error` - Any error encountered
  - Behavior:
    - All CreateBackup functionality plus:
    - Context cancellation support
    - Cancellation checks at multiple points
    - Returns appropriate error on cancellation
  - Error Cases: Same as CreateBackup plus context cancellation

- `CreateBackupWithContextAndCleanup(ctx context.Context, cfg *Config, filePath string, note string, dryRun bool) error`
  - Spec: "Context-aware backup creation with resource cleanup"
  - Input:
    - `ctx context.Context` - Context for cancellation
    - Other inputs same as CreateBackup
  - Output: `error` - Any error encountered
  - Behavior:
    - Combines all features of CreateBackupWithCleanup and CreateBackupWithContext
    - Context cancellation support with automatic cleanup
    - Atomic operations with cleanup on cancellation
    - Most robust backup creation function
  - Error Cases: Same as CreateBackup plus context cancellation

- `CreateBackupWithTime(cfg *Config, filePath string, note string, dryRun bool, now func() time.Time) error`
  - Spec: "Test helper for creating backups with custom time"
  - Input:
    - Same as CreateBackup
    - `now func() time.Time` - Custom time function for testing
  - Output: `error` - Any error encountered
  - Behavior: Same as CreateBackup but uses provided time function
  - Error Cases: Same as CreateBackup

**Example Usage**:
```go
// Generate backup name
name := GenerateBackupName("file.txt", "2024-03-20-15-30", "important_data")

// List all backups for a file
backups, err := ListBackups("/path/to/backups", "/path/to/original/file.txt")
if err != nil {
    log.Fatal(err)
}

// Create backup (basic)
err = CreateBackup(cfg, "/path/to/file.txt", "monthly_backup", false)

// Create backup with cleanup (recommended for production)
err = CreateBackupWithCleanup(cfg, "/path/to/file.txt", "monthly_backup", false)

// Create backup with context and cleanup (most robust)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err = CreateBackupWithContextAndCleanup(ctx, cfg, "/path/to/file.txt", "monthly_backup", false)

// Create backup with custom time (for testing)
err = CreateBackupWithTime(cfg, "/path/to/file.txt", "test_backup", false, func() time.Time {
    return time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
})
```

### File Comparison
**Implementation**: `backup.go`
**Specification Requirements**:
- `CompareFiles(file1, file2 string) (bool, error)`
  - Spec: "Performs byte-by-byte comparison of two files"
  - Input:
    - `file1 string` - Path to first file
    - `file2 string` - Path to second file
  - Output: `(bool, error)` - True if files are identical, error if comparison fails
  - Behavior:
    - Reads both files completely
    - Compares file lengths first for efficiency
    - Performs byte-by-byte comparison if lengths match
    - Returns false immediately if any difference found
  - Error Cases:
    - File not found
    - Permission denied
    - File system errors

**Example Usage**:
```go
// Compare two files
identical, err := CompareFiles("file1.txt", "file2.txt")
if err != nil {
    log.Fatal(err)
}
if identical {
    fmt.Println("Files are identical")
}
```

## Main Application Structure

### CLI Interface
**Implementation**: `main.go`
**Specification Requirements**:
- Uses `cobra` for command-line interface
- Global flags:
  - `--dry-run`: Implemented in `main.go`
    - Spec: "Show what would be done without creating backups"
    - Shows paths relative to current directory
  - `--list`: Implemented in `main.go`
    - Spec: "List all backups for the specified file"
    - Usage: `bkpfile --list [FILE_PATH]`
    - Shows paths relative to current directory
  - `--config`: Implemented in `main.go`
    - Spec: "Display computed configuration values and exit"
    - Usage: `bkpfile --config`
- Default behavior:
  - Creates backup of specified file with optional note
  - Usage: `bkpfile [FILE_PATH] [NOTE]`
  - Output: Shows backup path (relative to current directory) and creation time
  - When a new backup is created: Displays "Created backup: [PATH]"
  - When file is identical to existing backup: Displays "File is identical to existing backup: [PATH]"
  - Uses configured status codes for application exit

### Workflow Implementation
**Implementation**: `backup.go`
**Specification Requirements**:
- Backup creation workflow: `CreateBackup()` and enhanced variants
  - Spec: "Creates a copy of the specified file with proper error handling and cleanup"
  - Steps:
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
    9. Exit with `cfg.StatusCreatedBackup` on success

- Backup listing workflow: `ListBackups()`
  - Spec: "Displays all backups for the specified file"
  - Steps:
    1. Load config
    2. Convert source path to relative path if needed
    3. Find backup directory for the file
    4. List and filter backup files
    5. Extract backup information and notes
    6. Sort backups by creation time
    7. Display backup information

- Configuration display workflow: `DisplayConfig()`
  - Spec: "Displays computed configuration values and exits"
  - Steps:
    1. Read `BKPFILE_CONFIG` environment variable or use default search path
    2. Process configuration files in order with precedence rules
    3. Merge configuration values with defaults
    4. Track source file for each configuration value
    5. Display each configuration value with name, computed value, and source
    6. Exit application after display

### Utility Functions
**Implementation**: Various files
**Specification Requirements**:
- Backup naming: `GenerateBackupName()` in `backup.go`
  - Spec: "Follows format: filename-YYYY-MM-DD-hh-mm[=note]"
- File copying: `CopyFile()` and `CopyFileWithContext()` in `backup.go`
  - Spec: "Creates exact copy with permissions preserved, supports cancellation"
- Path handling: Various functions in `backup.go`
  - Spec: "Handles both absolute and relative paths consistently"
- File comparison: `CompareFiles()` in `backup.go`
  - Spec: "Performs byte-by-byte comparison of files"
  - Input: Source file path and most recent backup path
  - Output: Boolean indicating if files are identical
  - Behavior: Compares files byte by byte to detect changes
- Resource management: `ResourceManager` in `backup.go`
  - Spec: "Tracks and cleans up temporary resources automatically"
- Error handling: `BackupError` and related functions in `backup.go`
  - Spec: "Provides structured error handling with status codes"

## Build and Development Requirements

### Build System
**Implementation**: `Makefile`
**Specification Requirements**:
- **Linting**: `make lint` command
  - Spec: "Run revive linter on all Go code"
  - Must pass before code can be committed
  - Uses `.revive.toml` configuration file
- **Testing**: `make test` command
  - Spec: "Run all unit tests with verbose output"
  - Must pass before code can be committed
  - Includes resource cleanup tests
- **Building**: `make build` command
  - Spec: "Build the application binary"
  - Depends on linting and testing passing
- **Cleaning**: `make clean` command
  - Spec: "Remove build artifacts"

### Development Workflow
**Specification Requirements**:
- **Code Quality**: All code must pass linting before commit
- **Testing**: All tests must pass before commit
- **Error Handling**: All errors must be properly handled
- **Resource Management**: All temporary resources must be cleaned up
- **Documentation**: All public functions must be documented
- **Backward Compatibility**: New features must not break existing functionality
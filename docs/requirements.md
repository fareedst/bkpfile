# Architecture and Specification Requirements Traceability

This document maps code components to their corresponding architecture requirements and specification requirements.

> **Note**: For testing requirements and architecture, see [Testing Requirements](testing.md).

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

**Example Usage**:
```go
// Load configuration from YAML or use defaults
cfg, err := LoadConfig(".")
if err != nil {
    log.Fatal(err)
}

// Access configuration values
backupPath := cfg.BackupDirPath
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

**Example Usage**:
```go
// Create default configuration
cfg := DefaultConfig()

// Load configuration from YAML file
cfg, err := LoadConfig(".")
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

**Example Usage**:
```go
// Copy a file
err := CopyFile("source.txt", "backup.txt")
if err != nil {
    log.Fatal(err)
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
  - Error Cases:
    - Invalid configuration
    - File not found
    - File is not a regular file
    - Permission denied
    - Disk full
    - Other file system errors

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

// Create backup
err = CreateBackup(cfg, "/path/to/file.txt", "monthly_backup", false)
if err != nil {
    log.Fatal(err)
}

// Create backup with custom time (for testing)
err = CreateBackupWithTime(cfg, "/path/to/file.txt", "test_backup", false, func() time.Time {
    return time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
})
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
- Default behavior:
  - Creates backup of specified file with optional note
  - Usage: `bkpfile [FILE_PATH] [NOTE]`
  - Output: Shows backup path (relative to current directory) and creation time
  - When a new backup is created: Displays "Created backup: [PATH]"
  - When file is identical to existing backup: Displays "File is identical to existing backup: [PATH]"

### Workflow Implementation
**Implementation**: `backup.go`
**Specification Requirements**:
- Backup creation workflow: `CreateBackup()`
  - Spec: "Creates a copy of the specified file"
  - Steps:
    1. Load config
    2. Validate source file exists and is regular
    3. Convert file path to relative path if needed
    4. Compare file with most recent backup
       - If identical, report existing backup name and exit
       - If different, proceed with backup creation
    5. Generate backup name using base filename
    6. Create backup directory structure
    7. Create file copy (or simulate in dry-run)

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

### Utility Functions
**Implementation**: Various files
**Specification Requirements**:
- Backup naming: `GenerateBackupName()` in `backup.go`
  - Spec: "Follows format: filename-YYYY-MM-DD-hh-mm[=note]"
- File copying: `CopyFile()` in `backup.go`
  - Spec: "Creates exact copy with permissions preserved"
- Path handling: Various functions in `backup.go`
  - Spec: "Handles both absolute and relative paths consistently"
- File comparison: `CompareFiles()` in `backup.go`
  - Spec: "Performs byte-by-byte comparison of files"
  - Input: Source file path and most recent backup path
  - Output: Boolean indicating if files are identical
  - Behavior: Compares files byte by byte to detect changes 
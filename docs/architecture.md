# Architecture

This document describes the system architecture and design of the BkpFile application.

> **Note**: For testing architecture and requirements, see [Testing Requirements](testing.md).

## Data Objects

1. **Config**
   - `Config`: string - Colon-separated list of configuration file paths to search
   - `BackupDirPath`: string - Path where backups are stored
   - `UseCurrentDirName`: bool - Whether to use current directory name in backup path

2. **Backup**
   - `Name`: string - Name of the backup file
   - `Path`: string - Full path to the backup file
   - `CreationTime`: time.Time - When the backup was created
   - `SourceFile`: string - Path to the original file
   - `Note`: string - Optional note for the backup

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

2. **File System Operations**
   - `CopyFile(src, dst string) error`: Creates an exact copy of the specified file
     - Preserves file permissions and modification time
     - Creates destination directories if needed
     - Handles both absolute and relative paths

3. **Backup Management**
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
   - `CreateBackupWithTime(cfg *Config, filePath string, note string, dryRun bool, now func() time.Time) error`: Test helper for creating backups with custom time

## Main Application Structure

1. **CLI Interface**
   - Uses `cobra` for command-line interface
   - Global flags:
     - `--dry-run`: Show what would be done without creating backups
     - `--list`: List backups for the specified file
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
        - If identical, report existing backup name and exit
        - If different, proceed with backup creation
     5. Generate backup name using base filename
     6. Create backup directory structure
     7. Create file copy (or simulate in dry-run)

   - For listing backups:
     1. Load config
     2. Convert source path to relative path if needed
     3. Find backup directory for the file
     4. List and filter backup files
     5. Extract backup information and notes
     6. Sort backups by creation time
     7. Display backup information

3. **Utility Functions**
   - Backup naming follows format: `filename-YYYY-MM-DD-hh-mm[=note]`
   - File copy ensures all permissions are preserved
   - Source path structure is preserved in backup directory
   - Handles both absolute and relative paths consistently
   - Validates file types and existence
   - File comparison uses byte-by-byte comparison

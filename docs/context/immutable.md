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
- These file operation rules are fundamental and must not be altered

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
   - This backup creation logic must remain unchanged

## Configuration Defaults
- Configuration discovery uses `BKPFILE_CONFIG` environment variable to specify search path
- Default configuration search path: `./.bkpfile.yml:~/.bkpfile.yml` (if `BKPFILE_CONFIG` not set)
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
- Platform support must never be reduced or modified

## Global Options
- Support `--dry-run` flag for previewing backup operations
- Existing flag behavior must be maintained

## Feature Preservation Rules
1. New Features:
   - Must not interfere with existing functionality
   - Must maintain all current behaviors
   - Must be optional and not affect existing workflows

2. Modifications:
   - Must preserve all existing command-line interfaces
   - Must maintain current file handling behaviors
   - Must keep existing configuration options
   - Must not change established backup naming patterns

3. Testing Requirements:
   - All new code must include tests for existing functionality
   - Regression tests must verify no existing features are broken
   - Platform compatibility tests must be maintained 
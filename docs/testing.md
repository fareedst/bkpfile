# Testing Requirements and Architecture

This document outlines the testing requirements, architecture, and approach for the BkpFile application.

## Test Categories

### Unit Tests
**Implementation**: `*_test.go` files
**Requirements**:
- `TestGenerateBackupName`: Tests backup naming
  - Validates naming format with various inputs
- `TestDefaultConfig` and `TestLoadConfig`: Tests configuration
  - Validates configuration loading and defaults
- `TestCopyFile`: Tests file copying
  - Validates file copying and permission preservation
- `TestListBackups`: Tests backup listing
  - Validates backup listing and sorting
- `TestCreateBackup`: Tests backup creation
  - Validates backup creation with various scenarios
- `TestCompareFiles`: Tests file comparison
  - Validates byte-by-byte file comparison
  - Test cases:
    - Identical files
    - Different files
    - Files of different sizes
    - Empty files
    - Large files
    - Files with special characters

### Integration Tests
**Implementation**: `*_test.go` files
**Requirements**:
- `TestDefaultBackupCmd`: Tests default backup behavior
  - Validates backup creation with various notes
  - Test cases:
    - Backup with note
    - Backup without note
    - Backup of identical file (should report existing)
    - Backup of modified file
- `TestListFlag`: Tests list flag functionality
  - Validates --list flag behavior
  - Test cases:
    - List with existing backups
    - List with no backups
    - List with invalid file path
- `TestCmdArgsValidation`: Tests command-line arguments
  - Validates command-line interface
  - Test cases:
    - Invalid flag combinations
    - Missing file path
    - Invalid file path
- `TestDryRun`: Tests dry-run mode
  - Validates dry-run mode behavior
  - Test cases:
    - Dry run with identical file
    - Dry run with modified file
    - Dry run with list flag

## Test Approach
**Implementation**: `*_test.go` files
**Requirements**:
- Uses temporary directories for testing
- Simulates user environment with test files
- Tests both dry-run and actual execution modes
- Verifies correct behavior for edge cases and error conditions
- Uses `CreateBackupWithTime` for consistent timestamps in tests
- Tests path handling for both absolute and relative paths
- Tests file comparison with various file types and sizes
- Verifies correct exit behavior when files are identical
- Tests flag-based command structure
- Validates proper reporting of existing backup names

## Test Environment
- Uses Go's testing package
- Leverages temporary directories for file operations
- Mocks time functions for consistent testing
- Simulates file system operations
- Tests on both macOS and Linux platforms

## Test Coverage Requirements
- All core functions must have unit tests
- All command-line operations must have integration tests
- Edge cases must be covered
- Error conditions must be tested
- Platform-specific behaviors must be verified
- File system operations must be tested thoroughly

## Running Tests
```bash
go test ./internal/bkpfile -v
```

## Test Data Management
- Test files are created in temporary directories
- Test data is cleaned up after tests
- Test files use consistent naming patterns
- Test data covers various file sizes and types
- Test data includes special characters and edge cases 
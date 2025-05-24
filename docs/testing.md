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
  - Tests configuration discovery with `BKPFILE_CONFIG` environment variable
  - Tests multiple configuration file precedence
  - Tests home directory expansion in paths
- `TestGetConfigSearchPath`: Tests configuration path discovery
  - Validates environment variable parsing
  - Tests default path when environment variable not set
  - Tests colon-separated path list parsing
  - Tests home directory expansion
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
- `TestConfigurationDiscovery`: Tests configuration file discovery
  - Tests multiple configuration files with different precedence
  - Tests environment variable override behavior
  - Tests missing configuration files handling
  - Tests invalid configuration file handling
  - Tests configuration merging with defaults

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
- `TestConfigurationIntegration`: Tests configuration discovery in full application context
  - Tests backup operations with custom configuration paths
  - Tests environment variable override in real scenarios
  - Tests configuration precedence with actual file operations
  - Test cases:
    - Backup creation with `BKPFILE_CONFIG` set
    - List operation with multiple configuration files
    - Error handling with invalid configuration paths
    - Dry-run with custom configuration discovery

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
- Tests environment variable handling with various configurations
- Creates multiple temporary configuration files for precedence testing
- Tests home directory expansion and path resolution
- Tests configuration discovery error conditions and edge cases

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
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
  - Tests status code configuration loading and defaults
  - Validates all status code fields are properly loaded from YAML
  - Tests status code configuration with custom values
- `TestGetConfigSearchPath`: Tests configuration path discovery
  - Validates environment variable parsing
  - Tests hard-coded default path when environment variable not set
  - Tests colon-separated path list parsing
  - Tests home directory expansion
- `TestDisplayConfig`: Tests configuration value display
  - Validates configuration value computation and source tracking
  - Tests environment variable processing for configuration paths
  - Tests hard-coded default path handling
  - Tests default value handling and source attribution
  - Tests output format with name, value, and source
  - Tests display of status code configuration values
  - Test cases:
    - Configuration with default values only
    - Configuration from single file
    - Configuration from multiple files with precedence
    - Configuration with missing files
    - Configuration with invalid files
    - Environment variable override scenarios
    - Status code configuration display
- `TestCopyFile`: Tests file copying
  - Validates file copying and permission preservation
- `TestListBackups`: Tests backup listing
  - Validates backup listing and sorting
- `TestCreateBackup`: Tests backup creation
  - Validates backup creation with various scenarios
  - Tests status code exit behavior for different conditions
  - Test cases:
    - Successful backup creation (should exit with `cfg.StatusCreatedBackup`)
    - File identical to existing backup (should exit with `cfg.StatusFileIsIdenticalToExistingBackup`)
    - File not found (should exit with `cfg.StatusFileNotFound`)
    - Invalid file type (should exit with `cfg.StatusInvalidFileType`)
    - Permission denied (should exit with `cfg.StatusPermissionDenied`)
    - Disk full scenarios (should exit with `cfg.StatusDiskFull`)
    - Backup directory creation failure (should exit with `cfg.StatusFailedToCreateBackupDirectory`)
    - Configuration errors (should exit with `cfg.StatusConfigError`)
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
  - Tests hard-coded default path behavior
  - Tests missing configuration files handling
  - Tests invalid configuration file handling
  - Tests configuration merging with defaults
  - Tests status code configuration precedence and merging
- `TestStatusCodeConfiguration`: Tests status code configuration
  - Validates status code loading from YAML
  - Tests status code defaults
  - Tests status code precedence with multiple configuration files
  - Tests invalid status code values handling
  - Test cases:
    - Default status codes (should match immutable specification defaults)
    - Custom status codes from configuration file
    - Status code precedence with multiple files
    - Invalid status code values (non-integer)
    - Missing status code fields (should use defaults)

### Integration Tests
**Implementation**: `*_test.go` files
**Requirements**:
- `TestDefaultBackupCmd`: Tests default backup behavior
  - Validates backup creation with various notes
  - Tests status code exit behavior in full application context
  - Test cases:
    - Backup with note (should exit with configured `status_created_backup`)
    - Backup without note (should exit with configured `status_created_backup`)
    - Backup of identical file (should exit with configured `status_file_is_identical_to_existing_backup`)
    - Backup of modified file (should exit with configured `status_created_backup`)
    - Backup with custom status code configuration
- `TestListFlag`: Tests list flag functionality
  - Validates --list flag behavior
  - Test cases:
    - List with existing backups
    - List with no backups
    - List with invalid file path
- `TestConfigFlag`: Tests config flag functionality
  - Validates --config flag behavior
  - Test cases:
    - Display config with default values only (including status codes)
    - Display config with values from single configuration file
    - Display config with values from multiple configuration files
    - Display config with `BKPFILE_CONFIG` environment variable set
    - Display config with hard-coded default path when environment variable not set
    - Display config with invalid configuration files (error handling)
    - Display config with custom status code values
    - Verify application exits after displaying configuration
- `TestCmdArgsValidation`: Tests command-line arguments
  - Validates command-line interface
  - Test cases:
    - Invalid flag combinations
    - Missing file path
    - Invalid file path
    - Config flag with other arguments (should be ignored)
- `TestDryRun`: Tests dry-run mode
  - Validates dry-run mode behavior
  - Test cases:
    - Dry run with identical file
    - Dry run with modified file
    - Dry run with list flag
- `TestConfigurationIntegration`: Tests configuration discovery in full application context
  - Tests backup operations with custom configuration paths
  - Tests environment variable override in real scenarios
  - Tests hard-coded default path behavior
  - Tests configuration precedence with actual file operations
  - Tests status code configuration in real application scenarios
  - Test cases:
    - Backup creation with `BKPFILE_CONFIG` set
    - List operation with multiple configuration files
    - Error handling with invalid configuration paths
    - Dry-run with custom configuration discovery
    - Configuration display with various environment setups
    - Status code behavior with custom configuration
    - Application exit codes with different error conditions
- `TestStatusCodeIntegration`: Tests status code configuration in full application context
  - Validates application exit codes match configuration
  - Tests status code behavior with various error conditions
  - Test cases:
    - Application exits with correct status code for successful backup
    - Application exits with correct status code for identical file
    - Application exits with correct status code for file not found
    - Application exits with correct status code for permission denied
    - Application exits with correct status code for invalid file type
    - Status code configuration from multiple files with precedence
    - Default status codes when no configuration provided

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
- Validates configuration display output format and source tracking
- Tests application exit behavior after configuration display
- Tests status code configuration loading and application
- Validates application exit codes match configured values
- Tests status code behavior with various error conditions and scenarios
- Creates test configuration files with custom status code values
- Verifies status code precedence with multiple configuration files

## Test Environment
- Uses Go's testing package
- Leverages temporary directories for file operations
- Mocks time functions for consistent testing
- Simulates file system operations
- Tests on both macOS and Linux platforms
- Mocks environment variables for configuration testing
- Creates temporary configuration files with various content

## Test Coverage Requirements
- All core functions must have unit tests
- All command-line operations must have integration tests
- Edge cases must be covered
- Error conditions must be tested
- Platform-specific behaviors must be verified
- File system operations must be tested thoroughly
- Configuration display functionality must be comprehensively tested
- Environment variable handling must be verified
- Status code configuration must be thoroughly tested
- Application exit codes must be validated for all conditions
- Status code precedence and merging must be tested
- Default status code behavior must be verified

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
- Test configuration files cover various YAML structures and edge cases 
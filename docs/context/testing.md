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
- `TestCopyFileWithContext`: Tests context-aware file copying
  - Validates file copying with cancellation support
  - Test cases:
    - Successful copy with context
    - Copy cancelled via context
    - Copy with timeout
    - Copy with already cancelled context
- `TestListBackups`: Tests backup listing
  - Validates backup listing and sorting
- `TestCreateBackup`: Tests backup creation
  - Validates backup creation with various scenarios
  - Tests status code exit behavior for different conditions
  - Tests panic recovery and error handling
  - Test cases:
    - Successful backup creation (should exit with `cfg.StatusCreatedBackup`)
    - File identical to existing backup (should exit with `cfg.StatusFileIsIdenticalToExistingBackup`)
    - File not found (should exit with `cfg.StatusFileNotFound`)
    - Invalid file type (should exit with `cfg.StatusInvalidFileType`)
    - Permission denied (should exit with `cfg.StatusPermissionDenied`)
    - Disk full scenarios (should exit with `cfg.StatusDiskFull`)
    - Backup directory creation failure (should exit with `cfg.StatusFailedToCreateBackupDirectory`)
    - Configuration errors (should exit with `cfg.StatusConfigError`)
    - Panic recovery scenarios
- `TestCreateBackupWithCleanup`: Tests backup creation with resource cleanup
  - Validates automatic resource cleanup functionality
  - Tests atomic operations with temporary files
  - Test cases:
    - Successful backup with cleanup verification
    - Backup failure with cleanup verification
    - No temporary files left after operations
    - Atomic file operations
- `TestCreateBackupWithContext`: Tests context-aware backup creation
  - Validates backup creation with cancellation support
  - Test cases:
    - Successful backup with context
    - Backup cancelled via context
    - Backup with timeout
    - Context cancellation at various stages
- `TestCreateBackupWithContextAndCleanup`: Tests context-aware backup with cleanup
  - Validates most robust backup creation functionality
  - Combines context support with resource cleanup
  - Test cases:
    - Successful backup with context and cleanup
    - Cancelled backup with proper cleanup
    - Timeout scenarios with cleanup verification
    - No resource leaks on cancellation
- `TestResourceManager`: Tests resource management functionality
  - Validates thread-safe resource tracking
  - Tests automatic cleanup mechanisms
  - Test cases:
    - Basic resource registration and cleanup
    - Thread-safe concurrent access
    - Cleanup of both files and directories
    - Error-resilient cleanup (continues on individual failures)
    - Cleanup warnings logged to stderr
- `TestBackupError`: Tests structured error handling
  - Validates BackupError functionality
  - Test cases:
    - Error creation with message and status code
    - Error interface implementation
    - Status code extraction
    - Error message formatting
- `TestIsDiskFullError`: Tests enhanced disk space detection
  - Validates disk full error detection
  - Test cases:
    - Various disk full error messages
    - Case-insensitive matching
    - Multiple disk space indicators
    - Non-disk-full errors (should return false)
    - Nil error handling
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
  - Tests enhanced error handling in real scenarios
  - Test cases:
    - Backup with note (should exit with configured `status_created_backup`)
    - Backup without note (should exit with configured `status_created_backup`)
    - Backup of identical file (should exit with configured `status_file_is_identical_to_existing_backup`)
    - Backup of modified file (should exit with configured `status_created_backup`)
    - Backup with custom status code configuration
    - Error scenarios with proper status codes
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
    - Dry run with resource cleanup verification
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
- `TestResourceCleanupIntegration`: Tests resource cleanup in full application context
  - Validates resource cleanup across entire application workflow
  - Tests cleanup with various error scenarios
  - Test cases:
    - Successful operations with cleanup verification
    - Failed operations with cleanup verification
    - Interrupted operations with cleanup verification
    - No temporary files left in any scenario

### Linting and Code Quality Tests
**Implementation**: `Makefile`, CI/CD pipeline
**Requirements**:
- `make lint`: Runs revive linter
  - Validates all Go code passes linting standards
  - Checks error handling compliance
  - Validates code style and formatting
  - Test cases:
    - All source files pass revive checks
    - No unhandled errors
    - Proper function and variable naming
    - Adequate documentation
- Code quality validation:
  - All `fmt.Printf`, `fmt.Fprintf` return values checked
  - All file operations handle errors appropriately
  - Consistent error handling patterns
  - Proper resource cleanup in all code paths

### Performance and Stress Tests
**Implementation**: `*_test.go` files with benchmarks
**Requirements**:
- `BenchmarkCreateBackup`: Benchmarks backup creation performance
  - Tests performance with various file sizes
  - Measures resource cleanup overhead
  - Validates memory usage patterns
- `BenchmarkResourceManager`: Benchmarks resource management performance
  - Tests performance with many temporary resources
  - Measures cleanup time with large resource lists
  - Validates thread-safe performance
- Stress tests:
  - Concurrent backup operations
  - Large file handling
  - Many temporary resources
  - Resource cleanup under load

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
- Tests resource cleanup in all scenarios including failures
- Validates context cancellation and timeout handling
- Tests panic recovery and error resilience
- Verifies no resource leaks in any test scenario
- Tests atomic operations and data integrity
- Validates enhanced error detection and handling

## Test Environment
- Uses Go's testing package
- Leverages temporary directories for file operations
- Mocks time functions for consistent testing
- Simulates file system operations
- Tests on both macOS and Linux platforms
- Mocks environment variables for configuration testing
- Creates temporary configuration files with various content
- Uses context with timeouts for cancellation testing
- Simulates disk full and permission denied scenarios
- Tests with various file sizes and types
- Validates cleanup in all test scenarios

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
- Resource cleanup must be tested in all scenarios
- Context cancellation and timeout handling must be verified
- Enhanced error handling must be thoroughly tested
- Panic recovery must be validated
- Atomic operations must be tested
- Thread safety must be verified
- Performance characteristics must be measured
- Memory usage patterns must be validated

## Running Tests
```bash
# Run all tests
go test ./internal/bkpfile -v

# Run tests with coverage
go test ./internal/bkpfile -v -cover

# Run specific test categories
go test ./internal/bkpfile -v -run TestResourceManager
go test ./internal/bkpfile -v -run TestCreateBackupWithCleanup
go test ./internal/bkpfile -v -run TestCreateBackupWithContext

# Run benchmarks
go test ./internal/bkpfile -v -bench=.

# Run linting
make lint

# Run all quality checks
make test
```

## Test Data Management
- Test files are created in temporary directories
- Test data is cleaned up after tests
- Test files use consistent naming patterns
- Test data covers various file sizes and types
- Test data includes special characters and edge cases
- Test configuration files cover various YAML structures and edge cases
- Temporary resources are tracked and verified for cleanup
- Test data includes scenarios for context cancellation
- Test files simulate various error conditions
- Test data covers atomic operation scenarios

## Continuous Integration Requirements
- All tests must pass before code merge
- Linting must pass before code merge
- Code coverage must meet minimum thresholds
- Performance benchmarks must not regress
- Resource cleanup must be verified in CI environment
- Tests must run on multiple platforms
- Memory leak detection must be performed
- Static analysis must pass 
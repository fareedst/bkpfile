# Resource Cleanup Documentation

## Overview

The bkpfile backup system now includes comprehensive resource cleanup functionality to handle temporary files and ensure proper cleanup on errors or cancellation. This enhancement provides better reliability and prevents resource leaks during backup operations.

## Features

### 1. ResourceManager

The `ResourceManager` is a thread-safe utility that tracks temporary resources and ensures they are cleaned up when operations complete or fail.

#### Key Features:
- **Thread-safe**: Uses mutex for concurrent access
- **Automatic cleanup**: Cleans up resources on function exit via defer
- **Error resilience**: Continues cleanup even if individual operations fail
- **Logging**: Reports cleanup failures to stderr without failing the operation

#### Usage:
```go
rm := NewResourceManager()
defer rm.Cleanup()

// Register temporary files and directories
rm.AddTempFile("/path/to/temp/file.tmp")
rm.AddTempDir("/path/to/temp/dir")
```

### 2. Enhanced Backup Functions

#### CreateBackupWithCleanup
- **Purpose**: Creates backups with automatic resource cleanup
- **Features**: 
  - Atomic operations using temporary files
  - Automatic cleanup on errors
  - Panic recovery with cleanup
  - No resource leaks

#### CreateBackupWithContextAndCleanup
- **Purpose**: Context-aware backup creation with cleanup
- **Features**:
  - All features of `CreateBackupWithCleanup`
  - Cancellation support via context
  - Timeout handling
  - Cleanup on cancellation

## Implementation Details

### Atomic Operations

The enhanced backup functions use atomic operations to ensure data integrity:

1. **Temporary File Creation**: Backup data is first written to a `.tmp` file
2. **Atomic Move**: The temporary file is atomically moved to the final location
3. **Cleanup Registration**: Temporary files are registered for cleanup
4. **Success Handling**: Successfully moved files are removed from cleanup list

### Error Handling

The resource cleanup system handles various error scenarios:

- **Disk Full**: Temporary files are cleaned up if disk space runs out
- **Permission Errors**: Cleanup continues even if some files can't be removed
- **Cancellation**: Context cancellation triggers immediate cleanup
- **Panics**: Panic recovery ensures cleanup still occurs

### Thread Safety

The `ResourceManager` uses a mutex to ensure thread-safe operations:
- Multiple goroutines can safely register resources
- Cleanup operations are atomic
- No race conditions during concurrent access

## Usage Examples

### Basic Cleanup Usage

```go
func ExampleBasicCleanup() {
    cfg := &Config{
        BackupDirPath: ".bkpfile",
        // ... other config
    }
    
    err := CreateBackupWithCleanup(cfg, "myfile.txt", "backup note", false)
    if err != nil {
        // Handle error - cleanup is automatic
    }
}
```

### Context-Aware Usage

```go
func ExampleContextCleanup() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    cfg := &Config{
        BackupDirPath: ".bkpfile",
        // ... other config
    }
    
    err := CreateBackupWithContextAndCleanup(ctx, cfg, "myfile.txt", "backup note", false)
    if err != nil {
        // Handle error - cleanup is automatic even on timeout/cancellation
    }
}
```

### Manual Resource Management

```go
func ExampleManualResourceManagement() {
    rm := NewResourceManager()
    defer rm.Cleanup()
    
    // Create temporary file
    tempFile := "/tmp/mytemp.txt"
    if err := os.WriteFile(tempFile, []byte("data"), 0644); err != nil {
        return err
    }
    rm.AddTempFile(tempFile)
    
    // Create temporary directory
    tempDir := "/tmp/mytemp_dir"
    if err := os.MkdirAll(tempDir, 0755); err != nil {
        return err
    }
    rm.AddTempDir(tempDir)
    
    // Do work...
    // Cleanup happens automatically on function exit
}
```

## Error Scenarios and Cleanup

### Disk Full Scenarios

When disk space is exhausted:
1. Temporary files are automatically removed
2. Partial backup data is cleaned up
3. Error is returned with appropriate status code
4. No orphaned files remain

### Cancellation Scenarios

When operations are cancelled:
1. Context cancellation is detected at multiple checkpoints
2. Temporary files created before cancellation are cleaned up
3. Partial operations are rolled back
4. Clean error state is maintained

### Panic Recovery

When unexpected panics occur:
1. Panic is recovered with defer function
2. Resource cleanup still executes
3. Error message is logged to stderr
4. Application doesn't crash

## Testing

The resource cleanup functionality includes comprehensive tests:

### TestResourceManager
- Tests basic resource registration and cleanup
- Verifies thread safety
- Confirms cleanup of both files and directories

### TestCreateBackupWithCleanup
- Tests successful backup creation with cleanup
- Verifies no temporary files remain after success
- Confirms atomic operations work correctly

### TestCreateBackupWithContextAndCleanup
- Tests context cancellation scenarios
- Verifies timeout handling
- Confirms cleanup occurs on cancellation

## Best Practices

### When to Use Enhanced Functions

Use the enhanced cleanup functions when:
- **Reliability is critical**: Operations must not leave temporary files
- **Long-running operations**: Risk of interruption or cancellation
- **Production environments**: Where resource leaks are unacceptable
- **Automated systems**: Where manual cleanup isn't possible

### When to Use Original Functions

Use the original functions when:
- **Simple operations**: Quick backups with low failure risk
- **Legacy compatibility**: Existing code that works well
- **Performance critical**: Minimal overhead is required

### Resource Management Guidelines

1. **Always use defer**: Ensure cleanup happens even on early returns
2. **Register early**: Add resources to cleanup as soon as they're created
3. **Handle errors gracefully**: Don't fail operations due to cleanup issues
4. **Log cleanup failures**: Help with debugging but don't block operations

## Migration Guide

### Upgrading Existing Code

To upgrade existing backup code to use resource cleanup:

```go
// Before
err := CreateBackup(cfg, filePath, note, dryRun)

// After
err := CreateBackupWithCleanup(cfg, filePath, note, dryRun)
```

For context-aware operations:

```go
// Before
err := CreateBackupWithContext(ctx, cfg, filePath, note, dryRun)

// After
err := CreateBackupWithContextAndCleanup(ctx, cfg, filePath, note, dryRun)
```

### Backward Compatibility

- Original functions remain unchanged and fully supported
- No breaking changes to existing APIs
- Enhanced functions are additive improvements
- Existing tests continue to pass

## Performance Considerations

### Overhead

The resource cleanup functionality adds minimal overhead:
- **Memory**: Small overhead for tracking resources
- **CPU**: Minimal impact from mutex operations
- **I/O**: Additional temporary file operations for atomicity

### Benefits

The benefits outweigh the overhead:
- **Reliability**: Prevents resource leaks and corruption
- **Maintainability**: Easier debugging and troubleshooting
- **Robustness**: Better handling of error conditions
- **User Experience**: More predictable behavior

## Troubleshooting

### Common Issues

#### Cleanup Warnings
If you see cleanup warnings in stderr:
- Check file permissions
- Verify disk space availability
- Ensure no other processes are using the files

#### Performance Impact
If cleanup seems slow:
- Check for large numbers of temporary files
- Verify filesystem performance
- Consider cleanup frequency

### Debugging

To debug resource cleanup issues:
1. Check stderr for cleanup warnings
2. Monitor temporary file creation/deletion
3. Use process monitoring tools
4. Enable verbose logging if available

## Future Enhancements

Potential future improvements:
- **Configurable cleanup policies**: Custom cleanup strategies
- **Metrics collection**: Track cleanup performance and failures
- **Advanced recovery**: More sophisticated error recovery mechanisms
- **Resource limits**: Prevent excessive temporary file creation 
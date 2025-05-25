package bkpfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateBackupName(t *testing.T) {
	tests := []struct {
		name       string
		sourcePath string
		timestamp  string
		note       string
		want       string
	}{
		{
			name:       "main.go backup",
			sourcePath: "main.go",
			timestamp:  "2025-05-12-13-49",
			note:       "",
			want:       "main.go-2025-05-12-13-49",
		},
		{
			name:       "main.go backup with note",
			sourcePath: "main.go",
			timestamp:  "2025-05-12-13-49",
			note:       "before_refactor",
			want:       "main.go-2025-05-12-13-49=before_refactor",
		},
		{
			name:       "config file backup",
			sourcePath: ".bkpfile.yml",
			timestamp:  "2025-05-12-13-49",
			note:       "",
			want:       ".bkpfile.yml-2025-05-12-13-49",
		},
		{
			name:       "test note format",
			sourcePath: "main.go",
			timestamp:  "2025-05-12-13-49",
			note:       "test_note",
			want:       "main.go-2025-05-12-13-49=test_note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBackupName(tt.sourcePath, tt.timestamp, tt.note)
			if got != tt.want {
				t.Errorf("GenerateBackupName() = %v, want %v", got, tt.want)
				t.Logf("Name analysis:")
				t.Logf("  Expected parts:")
				for i, part := range strings.Split(tt.want, "-") {
					t.Logf("    %d: %q", i, part)
				}
				t.Logf("  Actual parts:")
				for i, part := range strings.Split(got, "-") {
					t.Logf("    %d: %q", i, part)
				}
				if strings.Contains(tt.want, "=") {
					t.Logf("  Expected note part: %q", strings.Split(tt.want, "=")[1])
				}
				if strings.Contains(got, "=") {
					t.Logf("  Actual note part: %q", strings.Split(got, "=")[1])
				}
			}
		})
	}
}

func TestCreateBackup(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	testDir := filepath.Join(tmpDir, "cmd", "bkpfile")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test file
	testFile := filepath.Join(testDir, "main.go")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup directory
	backupDir := filepath.Join(tmpDir, ".bkpfile")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	// Create config
	cfg := &Config{
		BackupDirPath: backupDir,
	}

	// Change to temp directory for relative path testing
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	tests := []struct {
		name       string
		filePath   string
		note       string
		dryRun     bool
		wantErr    bool
		wantBackup string
	}{
		{
			name:       "create backup without note",
			filePath:   "cmd/bkpfile/main.go",
			note:       "",
			dryRun:     false,
			wantErr:    false,
			wantBackup: filepath.Join(".bkpfile", "cmd", "bkpfile", "main.go-2025-05-12-13-49"),
		},
		{
			name:       "create backup with note",
			filePath:   "cmd/bkpfile/main.go",
			note:       "test_note",
			dryRun:     false,
			wantErr:    false,
			wantBackup: filepath.Join(".bkpfile", "cmd", "bkpfile", "main.go-2025-05-12-13-49"),
		},
		{
			name:       "dry run backup",
			filePath:   "cmd/bkpfile/main.go",
			note:       "",
			dryRun:     true,
			wantErr:    false,
			wantBackup: filepath.Join(".bkpfile", "cmd", "bkpfile", "main.go-2025-05-12-13-49"),
		},
		{
			name:       "non-existent file",
			filePath:   "cmd/bkpfile/nonexistent.go",
			note:       "",
			dryRun:     false,
			wantErr:    true,
			wantBackup: "",
		},
	}

	// Mock time.Now for consistent timestamps
	mockTime := func() time.Time {
		t, _ := time.Parse("2006-01-02-15-04", "2025-05-12-13-49")
		return t
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create backup directory structure for each test
			if !tt.dryRun && !tt.wantErr {
				backupSubDir := filepath.Join(backupDir, filepath.Dir(tt.filePath))
				if err := os.MkdirAll(backupSubDir, 0755); err != nil {
					t.Fatalf("Failed to create backup subdirectory: %v", err)
				}
			}

			err := CreateBackupWithTime(cfg, tt.filePath, tt.note, tt.dryRun, mockTime)

			// Check for BackupError and determine if it's a success or failure
			if backupErr, ok := err.(*BackupError); ok {
				isSuccess := backupErr.Message == "backup created successfully" ||
					backupErr.Message == "dry run completed" ||
					backupErr.Message == "file is identical to existing backup"

				if tt.wantErr && isSuccess {
					t.Errorf("CreateBackup() expected error but got success: %v", backupErr.Message)
				} else if !tt.wantErr && !isSuccess {
					t.Errorf("CreateBackup() expected success but got error: %v", backupErr.Message)
				}
			} else if err != nil {
				// Regular error (not BackupError)
				if !tt.wantErr {
					t.Errorf("CreateBackup() unexpected error = %v", err)
				}
			} else {
				// No error at all
				if tt.wantErr {
					t.Errorf("CreateBackup() expected error but got nil")
				}
			}

			if !tt.dryRun && !tt.wantErr {
				// Verify backup was created
				backups, err := ListBackups(backupDir, tt.filePath)
				if err != nil {
					t.Errorf("ListBackups() error = %v", err)
				}
				if len(backups) == 0 {
					t.Error("No backups found after creation")
					// List directory contents for debugging
					if entries, err := os.ReadDir(backupDir); err == nil {
						t.Log("Backup directory contents:")
						for _, entry := range entries {
							t.Logf("- %s", entry.Name())
							if entry.IsDir() {
								subDir := filepath.Join(backupDir, entry.Name())
								if subEntries, err := os.ReadDir(subDir); err == nil {
									for _, subEntry := range subEntries {
										t.Logf("  - %s", subEntry.Name())
									}
								}
							}
						}
					}
				} else {
					// Only check the first backup, since identical content should not create a new backup even with a different note
					expectedName := filepath.Base(tt.wantBackup)
					if backups[0].Name != expectedName {
						t.Errorf("Backup with expected name %q not found in backups: %v", expectedName, backups)
					}
				}
			}
		})
	}
}

func TestCompareFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "bkpfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	file3 := filepath.Join(tempDir, "file3.txt")

	// Write test data
	testData := []byte("test data")
	if err := os.WriteFile(file1, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(file2, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("different data"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test cases
	tests := []struct {
		name    string
		file1   string
		file2   string
		want    bool
		wantErr bool
	}{
		{
			name:    "identical files",
			file1:   file1,
			file2:   file2,
			want:    true,
			wantErr: false,
		},
		{
			name:    "different files",
			file1:   file1,
			file2:   file3,
			want:    false,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file1:   file1,
			file2:   "nonexistent.txt",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareFiles(tt.file1, tt.file2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "bkpfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with specific modification time
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test data")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Set a specific modification time for the test file
	testTime := time.Date(2024, 3, 20, 15, 30, 0, 0, time.UTC)
	if err := os.Chtimes(testFile, time.Now(), testTime); err != nil {
		t.Fatalf("Failed to set test file time: %v", err)
	}

	// Create destination file
	dstFile := filepath.Join(tempDir, "copy.txt")

	// Copy the file
	if err := CopyFile(testFile, dstFile); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Verify file contents
	data, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Copied file contents = %v, want %v", string(data), string(testData))
	}

	// Verify file permissions
	srcInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to get source file info: %v", err)
	}
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Failed to get destination file info: %v", err)
	}
	if dstInfo.Mode() != srcInfo.Mode() {
		t.Errorf("Copied file mode = %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}

	// Verify modification time
	if !dstInfo.ModTime().Equal(testTime) {
		t.Errorf("Copied file modification time = %v, want %v", dstInfo.ModTime(), testTime)
	}
}

func TestRelativePathDisplay(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		note     string
		dryRun   bool
	}{
		{
			name:     "backup with relative path",
			filePath: "cmd/bkpfile/main.go",
			note:     "test_note",
			dryRun:   false,
		},
		{
			name:     "dry run with relative path",
			filePath: "cmd/bkpfile/main.go",
			note:     "test_note",
			dryRun:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "bkpfile-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test file
			testFile := filepath.Join(tmpDir, tt.filePath)
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("Failed to create test file directory: %v", err)
			}
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Create backup directory
			backupDir := filepath.Join(tmpDir, ".bkpfile")
			if err := os.MkdirAll(backupDir, 0755); err != nil {
				t.Fatalf("Failed to create backup directory: %v", err)
			}

			// Create config
			cfg := &Config{
				BackupDirPath: backupDir,
			}

			// Create backup with mocked time
			mockTime := time.Date(2025, 5, 12, 13, 49, 0, 0, time.Local)
			timeNow = func() time.Time { return mockTime }

			// Change to temp directory
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer os.Chdir(oldWd)

			// Create backup
			err = CreateBackupWithTime(cfg, tt.filePath, tt.note, tt.dryRun, timeNow)
			if err != nil {
				// Check if it's a BackupError with a success message
				if backupErr, ok := err.(*BackupError); ok {
					isSuccess := backupErr.Message == "backup created successfully" ||
						backupErr.Message == "dry run completed" ||
						backupErr.Message == "file is identical to existing backup"
					if !isSuccess {
						t.Fatalf("Failed to create backup: %v", backupErr.Message)
					}
				} else {
					t.Fatalf("Failed to create backup: %v", err)
				}
			}

			// Get backup path
			backupSubDir := filepath.Join(backupDir, filepath.Dir(tt.filePath))
			backupName := fmt.Sprintf("%s-%s=%s", filepath.Base(tt.filePath), mockTime.Format("2006-01-02-15-04"), tt.note)
			backupPath := filepath.Join(backupSubDir, backupName)

			// Get relative path
			relPath, err := filepath.Rel(tmpDir, backupPath)
			if err != nil {
				t.Fatalf("Failed to get relative path: %v", err)
			}

			// Verify relative path
			expectedPath := filepath.Join(".bkpfile", tt.filePath)
			expectedPath = fmt.Sprintf("%s-%s=%s", expectedPath, mockTime.Format("2006-01-02-15-04"), tt.note)

			// Add debug information
			t.Logf("Test case: %s", tt.name)
			t.Logf("File path: %s", tt.filePath)
			t.Logf("Note: %s", tt.note)
			t.Logf("Backup name: %s", backupName)
			t.Logf("Expected path: %s", expectedPath)
			t.Logf("Actual path: %s", relPath)
			t.Logf("Backup path: %s", backupPath)
			t.Logf("Temp dir: %s", tmpDir)

			if relPath != expectedPath {
				t.Errorf("Relative path = %s, want %s", relPath, expectedPath)
				t.Logf("Expected path components:")
				for _, part := range filepath.SplitList(expectedPath) {
					t.Logf("  - %s", part)
				}
				t.Logf("Actual path components:")
				for _, part := range filepath.SplitList(relPath) {
					t.Logf("  - %s", part)
				}
			}
		})
	}
}

// TestBackupOutputMessage verifies that the correct output message is displayed
// when creating a new backup or when a file is identical to an existing backup
func TestBackupOutputMessage(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup directory
	backupDir := filepath.Join(tmpDir, ".bkpfile")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	// Create config
	cfg := &Config{
		BackupDirPath: backupDir,
	}

	// Change to temp directory for relative path testing
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Mock time for consistent timestamps
	mockTime := func() time.Time {
		t, _ := time.Parse("2006-01-02-15-04", "2025-05-12-13-49")
		return t
	}

	// Test cases
	tests := []struct {
		name               string
		modifyFile         bool
		note               string
		wantOutputContains string
	}{
		{
			name:               "create new backup",
			modifyFile:         true, // Modify file to ensure it's different
			note:               "",
			wantOutputContains: "Created backup:",
		},
		{
			name:               "create new backup with note",
			modifyFile:         true, // Modify file to ensure it's different
			note:               "test_note",
			wantOutputContains: "Created backup:",
		},
		{
			name:               "identical file",
			modifyFile:         false, // Don't modify file to ensure it's identical
			note:               "",
			wantOutputContains: "File is identical to existing backup:",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Only modify the file before the first test to ensure the file is unchanged for the "identical" test
			if i == 0 || tt.modifyFile {
				// Create a unique file for each test or modify it to ensure it's different
				content := fmt.Sprintf("test content %d", i)
				if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to update test file: %v", err)
				}
			}

			// Create backup
			err := CreateBackupWithTime(cfg, testFile, tt.note, false, mockTime)
			if err != nil {
				// Check if it's a BackupError with a success message
				if backupErr, ok := err.(*BackupError); ok {
					isSuccess := backupErr.Message == "backup created successfully" ||
						backupErr.Message == "dry run completed" ||
						backupErr.Message == "file is identical to existing backup"
					if !isSuccess {
						t.Fatalf("Failed to create backup: %v", backupErr.Message)
					}
				} else {
					t.Fatalf("Failed to create backup: %v", err)
				}
			}

			// Restore stdout and get captured output
			w.Close()
			os.Stdout = oldStdout
			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to read captured output: %v", err)
			}
			output := strings.TrimSpace(buf.String())

			// Verify output message
			if !strings.Contains(output, tt.wantOutputContains) {
				t.Errorf("Output message = %q, doesn't contain %q", output, tt.wantOutputContains)
			}
		})
	}
}

// TestDuplicateFileDetection tests that files with identical content
// are correctly identified as duplicates, regardless of the note
func TestDuplicateFileDetection(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup directory
	backupDir := filepath.Join(tmpDir, ".bkpfile")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	// Create config
	cfg := &Config{
		BackupDirPath: backupDir,
	}

	// Change to temp directory for relative path testing
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Mock time for consistent timestamps
	mockTime := func() time.Time {
		t, _ := time.Parse("2006-01-02-15-04", "2025-05-12-13-49")
		return t
	}

	// Create initial backup with first note
	firstNote := "first_note"
	err = CreateBackupWithTime(cfg, testFile, firstNote, false, mockTime)
	if err != nil {
		// Check if it's a BackupError with a success message
		if backupErr, ok := err.(*BackupError); ok {
			isSuccess := backupErr.Message == "backup created successfully" ||
				backupErr.Message == "dry run completed" ||
				backupErr.Message == "file is identical to existing backup"
			if !isSuccess {
				t.Fatalf("Failed to create initial backup: %v", backupErr.Message)
			}
		} else {
			t.Fatalf("Failed to create initial backup: %v", err)
		}
	}

	// Count backups after first creation
	backups, err := ListBackups(backupDir, testFile)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	initialBackupCount := len(backups)
	if initialBackupCount != 1 {
		t.Fatalf("Expected 1 backup after initial creation, got %d", initialBackupCount)
	}

	// Try to create another backup with a different note but identical file content
	secondNote := "second_note"

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Attempt to create a backup with different note
	err = CreateBackupWithTime(cfg, testFile, secondNote, false, mockTime)
	if err != nil {
		// Check if it's a BackupError with a success message
		if backupErr, ok := err.(*BackupError); ok {
			isSuccess := backupErr.Message == "backup created successfully" ||
				backupErr.Message == "dry run completed" ||
				backupErr.Message == "file is identical to existing backup"
			if !isSuccess {
				t.Fatalf("Failed during second backup attempt: %v", backupErr.Message)
			}
		} else {
			t.Fatalf("Failed during second backup attempt: %v", err)
		}
	}

	// Restore stdout and get captured output
	w.Close()
	os.Stdout = oldStdout
	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	output := strings.TrimSpace(buf.String())

	// Count backups after second creation attempt
	backups, err = ListBackups(backupDir, testFile)
	if err != nil {
		t.Fatalf("Failed to list backups after second attempt: %v", err)
	}

	// CURRENT BEHAVIOR: A new backup is created even though the file content is identical
	// EXPECTED BEHAVIOR: No new backup should be created, and the existing backup should be reported

	// Check if the message indicates an identical file was found
	if !strings.Contains(output, "File is identical to existing backup:") {
		t.Errorf("Expected message about identical backup, got: %q", output)
	}

	// Check that no new backup was created
	if len(backups) > initialBackupCount {
		t.Errorf("Expected no new backup to be created for identical file content, got %d backups (expected %d)",
			len(backups), initialBackupCount)
	}
}

func TestStandardConfigFile(t *testing.T) {
	// Save original environment variable
	originalEnv := os.Getenv("BKPFILE_CONFIG")
	defer func() {
		if originalEnv != "" {
			os.Setenv("BKPFILE_CONFIG", originalEnv)
		} else {
			os.Unsetenv("BKPFILE_CONFIG")
		}
	}()

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create .bkpfile.yml file
	standardConfig := `backup_dir_path: "./standard-backup"
use_current_dir_name: true`

	if err := os.WriteFile(".bkpfile.yml", []byte(standardConfig), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Clear environment variable to test default behavior
	os.Unsetenv("BKPFILE_CONFIG")

	// Load configuration (should find .bkpfile.yml file)
	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	// Verify configuration is loaded
	if cfg.BackupDirPath != "./standard-backup" {
		t.Errorf("LoadConfig().BackupDirPath = %q, want %q", cfg.BackupDirPath, "./standard-backup")
	}

	if cfg.UseCurrentDirName != true {
		t.Errorf("LoadConfig().UseCurrentDirName = %v, want %v", cfg.UseCurrentDirName, true)
	}

	// Create test file and backup
	testFile := "config-test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup directory
	if err := os.MkdirAll(cfg.BackupDirPath, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Mock time for consistent testing
	mockTime := func() time.Time {
		t, _ := time.Parse("2006-01-02-15-04", "2025-05-12-13-49")
		return t
	}

	// Test backup creation with configuration
	err = CreateBackupWithTime(cfg, testFile, "config_test", false, mockTime)
	if err != nil {
		// Check if it's a BackupError with a success message
		if backupErr, ok := err.(*BackupError); ok {
			isSuccess := backupErr.Message == "backup created successfully" ||
				backupErr.Message == "dry run completed" ||
				backupErr.Message == "file is identical to existing backup"
			if !isSuccess {
				t.Errorf("CreateBackupWithTime() with config error: %v", backupErr.Message)
			}
		} else {
			t.Errorf("CreateBackupWithTime() with config error: %v", err)
		}
	}

	// Verify backup was created
	backups, err := ListBackups(cfg.BackupDirPath, testFile)
	if err != nil {
		t.Errorf("ListBackups() error: %v", err)
	}

	if len(backups) == 0 {
		t.Error("No backups found after creation with config")
	}
}

func TestConfigurationIntegration(t *testing.T) {
	// Save original environment variable
	originalEnv := os.Getenv("BKPFILE_CONFIG")
	defer func() {
		if originalEnv != "" {
			os.Setenv("BKPFILE_CONFIG", originalEnv)
		} else {
			os.Unsetenv("BKPFILE_CONFIG")
		}
	}()

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory for relative path testing
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create test file
	testFile := "test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create custom configuration files
	globalBackupDir := filepath.Join(tmpDir, "global-backup")
	globalConfig := fmt.Sprintf(`backup_dir_path: "%s"
use_current_dir_name: false`, globalBackupDir)

	localConfig := `backup_dir_path: "./local-backup"
use_current_dir_name: true`

	if err := os.WriteFile("global.yml", []byte(globalConfig), 0644); err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}

	if err := os.WriteFile("local.yml", []byte(localConfig), 0644); err != nil {
		t.Fatalf("Failed to create local config: %v", err)
	}

	tests := []struct {
		name           string
		envValue       string
		expectedBackup string
		expectedUseDir bool
	}{
		{
			name:           "backup with global config",
			envValue:       "global.yml",
			expectedBackup: globalBackupDir,
			expectedUseDir: false,
		},
		{
			name:           "backup with local config",
			envValue:       "local.yml",
			expectedBackup: "./local-backup",
			expectedUseDir: true,
		},
		{
			name:           "backup with config precedence",
			envValue:       "global.yml:local.yml",
			expectedBackup: globalBackupDir,
			expectedUseDir: false,
		},
	}

	// Mock time.Now for consistent timestamps
	mockTime := func() time.Time {
		t, _ := time.Parse("2006-01-02-15-04", "2025-05-12-13-49")
		return t
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("BKPFILE_CONFIG", tt.envValue)

			// Load configuration
			cfg, err := LoadConfig(".")
			if err != nil {
				t.Errorf("LoadConfig() unexpected error: %v", err)
				return
			}

			// Verify configuration values
			if cfg.BackupDirPath != tt.expectedBackup {
				t.Errorf("LoadConfig().BackupDirPath = %q, want %q", cfg.BackupDirPath, tt.expectedBackup)
			}

			if cfg.UseCurrentDirName != tt.expectedUseDir {
				t.Errorf("LoadConfig().UseCurrentDirName = %v, want %v", cfg.UseCurrentDirName, tt.expectedUseDir)
			}

			// Create backup directory for the test
			if err := os.MkdirAll(cfg.BackupDirPath, 0755); err != nil {
				t.Errorf("Failed to create backup directory: %v", err)
				return
			}

			// Test dry-run with custom configuration
			err = CreateBackupWithTime(cfg, testFile, "config_test", true, mockTime)
			if err != nil {
				// Check if it's a BackupError with a success message
				if backupErr, ok := err.(*BackupError); ok {
					isSuccess := backupErr.Message == "backup created successfully" ||
						backupErr.Message == "dry run completed" ||
						backupErr.Message == "file is identical to existing backup"
					if !isSuccess {
						t.Errorf("CreateBackupWithTime() dry-run error: %v", backupErr.Message)
					}
				} else {
					t.Errorf("CreateBackupWithTime() dry-run error: %v", err)
				}
			}

			// Test actual backup creation with custom configuration
			err = CreateBackupWithTime(cfg, testFile, "config_test", false, mockTime)
			if err != nil {
				// Check if it's a BackupError with a success message
				if backupErr, ok := err.(*BackupError); ok {
					isSuccess := backupErr.Message == "backup created successfully" ||
						backupErr.Message == "dry run completed" ||
						backupErr.Message == "file is identical to existing backup"
					if !isSuccess {
						t.Errorf("CreateBackupWithTime() error: %v", backupErr.Message)
					}
				} else {
					t.Errorf("CreateBackupWithTime() error: %v", err)
				}
			}

			// Verify backup was created in the correct location
			backups, err := ListBackups(cfg.BackupDirPath, testFile)
			if err != nil {
				t.Errorf("ListBackups() error: %v", err)
			}

			if len(backups) == 0 {
				t.Error("No backups found after creation with custom config")
			}

			// Clean up backup directory for next test
			os.RemoveAll(cfg.BackupDirPath)
		})
	}
}

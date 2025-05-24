package bkpfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// timeNow is a variable that can be replaced for testing
var timeNow = time.Now

// Backup represents a file backup
// Architecture: Data Objects - Backup
type Backup struct {
	// Name is the backup filename
	// Architecture: Backup.Name
	Name string

	// Path is the full path to the backup file
	// Architecture: Backup.Path
	Path string

	// CreationTime is when the backup was created
	// Architecture: Backup.CreationTime
	CreationTime time.Time

	// SourceFile is the path to the original file
	// Architecture: Backup.SourceFile
	SourceFile string

	// Note is an optional note for the backup
	// Architecture: Backup.Note
	Note string
}

// GenerateBackupName generates a backup filename according to the specified format
// Architecture: Core Functions - Backup Management - GenerateBackupName
func GenerateBackupName(sourcePath, timestamp, note string) string {
	// Use just the filename for the backup name
	name := filepath.Base(sourcePath)

	// Add timestamp
	name = fmt.Sprintf("%s-%s", name, timestamp)

	// Add note if provided
	if note != "" {
		// Add note with equals sign
		name = fmt.Sprintf("%s=%s", name, note)
	}

	return name
}

// CopyFile creates an exact copy of the specified file
// Architecture: Core Functions - File System Operations - CopyFile
func CopyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write to destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to copy file permissions: %w", err)
	}

	// Set modification time to match source file
	if err := os.Chtimes(dst, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to set file modification time: %w", err)
	}

	return nil
}

// ListBackups gets all backups for a specific file
// Architecture: Core Functions - Backup Management - ListBackups
func ListBackups(backupDir string, sourceFile string) ([]Backup, error) {
	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	// Get the source path relative to current directory
	sourcePath := sourceFile
	if !filepath.IsAbs(sourceFile) {
		absPath, err := filepath.Abs(sourceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		relPath, err := filepath.Rel(wd, absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}
		sourcePath = relPath
	}

	// Get the directory and filename parts
	dir := filepath.Dir(sourcePath)
	filename := filepath.Base(sourcePath)

	// Construct the backup directory path
	backupSubDir := filepath.Join(backupDir, dir)
	if _, err := os.Stat(backupSubDir); os.IsNotExist(err) {
		return nil, nil // No backups exist for this file
	}

	// List all files in backup directory
	entries, err := os.ReadDir(backupSubDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if the backup file matches the source filename
		// The backup name format is: filename-timestamp[=note]
		if !strings.HasPrefix(entry.Name(), filename+"-") {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Create backup object
		backup := Backup{
			Name:         entry.Name(),
			Path:         filepath.Join(backupSubDir, entry.Name()),
			CreationTime: info.ModTime(),
			SourceFile:   sourceFile,
		}

		// Extract note if present
		if idx := strings.LastIndex(entry.Name(), "="); idx > 0 {
			backup.Note = entry.Name()[idx+1:]
		}

		backups = append(backups, backup)
	}

	// Sort backups by creation time (most recent first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreationTime.After(backups[j].CreationTime)
	})

	return backups, nil
}

// CreateBackup creates a backup of the specified file
// Architecture: Core Functions - Backup Management - CreateBackup
func CreateBackup(cfg *Config, filePath string, note string, dryRun bool) error {
	// Check if source file exists and is a regular file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	// Get the source path relative to current directory
	sourcePath := filePath
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		relPath, err := filepath.Rel(wd, absPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		sourcePath = relPath
	}

	// Get the directory part
	dir := filepath.Dir(sourcePath)

	// Check for existing backups
	backups, err := ListBackups(cfg.BackupDirPath, filePath)
	if err != nil {
		return fmt.Errorf("failed to list existing backups: %w", err)
	}

	// If there are existing backups, compare with the most recent one
	if len(backups) > 0 {
		mostRecent := backups[0] // ListBackups sorts by most recent first
		identical, err := CompareFiles(filePath, mostRecent.Path)
		if err != nil {
			return fmt.Errorf("failed to compare files: %w", err)
		}
		// Only check if the content is identical, not the note
		if identical {
			// Get relative path for display
			relPath, err := filepath.Rel(".", mostRecent.Path)
			if err != nil {
				relPath = mostRecent.Path // Fallback to absolute path if relative path fails
			}
			fmt.Printf("File is identical to existing backup: %s\n", relPath)
			return nil
		}
	}

	// Generate backup name with just the filename and note
	filename := filepath.Base(sourcePath)
	timestamp := timeNow().Format("2006-01-02-15-04")
	backupName := GenerateBackupName(filename, timestamp, note)

	// Determine backup path
	backupDir := cfg.BackupDirPath
	backupSubDir := filepath.Join(backupDir, dir)
	backupPath := filepath.Join(backupSubDir, backupName)

	// Create backup
	if dryRun {
		// Get relative path for display
		relPath, err := filepath.Rel(".", backupPath)
		if err != nil {
			relPath = backupPath // Fallback to absolute path if relative path fails
		}
		fmt.Printf("Would create backup: %s\n", relPath)
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy the file
	if err := CopyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Get relative path for display
	relPath, err := filepath.Rel(".", backupPath)
	if err != nil {
		relPath = backupPath // Fallback to absolute path if relative path fails
	}
	fmt.Printf("Created backup: %s\n", relPath)

	return nil
}

// CreateBackupWithTime creates a backup of the specified file with a custom time function
// This is used for testing to provide consistent timestamps
func CreateBackupWithTime(cfg *Config, filePath string, note string, dryRun bool, now func() time.Time) error {
	// Check if source file exists and is a regular file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	// Get the source path relative to current directory
	sourcePath := filePath
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		relPath, err := filepath.Rel(wd, absPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		sourcePath = relPath
	}

	// Get the directory part
	dir := filepath.Dir(sourcePath)

	// Check for existing backups
	backups, err := ListBackups(cfg.BackupDirPath, filePath)
	if err != nil {
		return fmt.Errorf("failed to list existing backups: %w", err)
	}

	// If there are existing backups, compare with the most recent one
	if len(backups) > 0 {
		mostRecent := backups[0] // ListBackups sorts by most recent first
		identical, err := CompareFiles(filePath, mostRecent.Path)
		if err != nil {
			return fmt.Errorf("failed to compare files: %w", err)
		}
		// If the content is identical, skip backup regardless of note
		if identical {
			// Get relative path for display
			relPath, err := filepath.Rel(".", mostRecent.Path)
			if err != nil {
				relPath = mostRecent.Path // Fallback to absolute path if relative path fails
			}
			fmt.Printf("File is identical to existing backup: %s\n", relPath)
			return nil
		}
	}

	// Generate backup name with just the filename and note
	filename := filepath.Base(sourcePath)
	timestamp := now().Format("2006-01-02-15-04")
	backupName := GenerateBackupName(filename, timestamp, note)

	// Determine backup path
	backupDir := cfg.BackupDirPath
	backupSubDir := filepath.Join(backupDir, dir)
	backupPath := filepath.Join(backupSubDir, backupName)

	// Create backup
	if dryRun {
		// Get relative path for display
		relPath, err := filepath.Rel(".", backupPath)
		if err != nil {
			relPath = backupPath // Fallback to absolute path if relative path fails
		}
		fmt.Printf("Would create backup: %s\n", relPath)
		return nil
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy the file
	if err := CopyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Get relative path for display
	relPath, err := filepath.Rel(".", backupPath)
	if err != nil {
		relPath = backupPath // Fallback to absolute path if relative path fails
	}
	fmt.Printf("Created backup: %s\n", relPath)

	return nil
}

// CompareFiles performs a byte-by-byte comparison of two files
// Architecture: Core Functions - File System Operations - CompareFiles
func CompareFiles(file1, file2 string) (bool, error) {
	// Read both files
	data1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("failed to read first file: %w", err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("failed to read second file: %w", err)
	}

	// Compare lengths first
	if len(data1) != len(data2) {
		return false, nil
	}

	// Compare bytes
	for i := 0; i < len(data1); i++ {
		if data1[i] != data2[i] {
			return false, nil
		}
	}

	return true, nil
}

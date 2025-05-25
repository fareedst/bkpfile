package bkpfile

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BackupDirPath != "../.bkpfile" {
		t.Errorf("DefaultConfig().BackupDirPath = %q, want %q", cfg.BackupDirPath, "../.bkpfile")
	}

	if cfg.UseCurrentDirName != true {
		t.Errorf("DefaultConfig().UseCurrentDirName = %v, want %v", cfg.UseCurrentDirName, true)
	}
}

func TestGetConfigSearchPath(t *testing.T) {
	// Save original environment variable
	originalEnv := os.Getenv("BKPFILE_CONFIG")
	defer func() {
		if originalEnv != "" {
			os.Setenv("BKPFILE_CONFIG", originalEnv)
		} else {
			os.Unsetenv("BKPFILE_CONFIG")
		}
	}()

	tests := []struct {
		name     string
		envValue string
		want     []string
	}{
		{
			name:     "hard-coded default paths when env var not set",
			envValue: "",
			want:     []string{"./.bkpfile.yml", expandHomeDir("~/.bkpfile.yml")},
		},
		{
			name:     "single path from env var",
			envValue: "/etc/bkpfile.yml",
			want:     []string{"/etc/bkpfile.yml"},
		},
		{
			name:     "multiple paths from env var",
			envValue: "/etc/bkpfile.yml:~/.config/bkpfile.yml:./.bkpfile.yml",
			want:     []string{"/etc/bkpfile.yml", expandHomeDir("~/.config/bkpfile.yml"), "./.bkpfile.yml"},
		},
		{
			name:     "paths with home directory expansion",
			envValue: "~/.bkpfile.yml:~/config/bkpfile.yml",
			want:     []string{expandHomeDir("~/.bkpfile.yml"), expandHomeDir("~/config/bkpfile.yml")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("BKPFILE_CONFIG", tt.envValue)
			} else {
				os.Unsetenv("BKPFILE_CONFIG")
			}

			got := GetConfigSearchPath()

			if len(got) != len(tt.want) {
				t.Errorf("GetConfigSearchPath() returned %d paths, want %d", len(got), len(tt.want))
				t.Errorf("Got: %v", got)
				t.Errorf("Want: %v", tt.want)
				return
			}

			for i, path := range got {
				if path != tt.want[i] {
					t.Errorf("GetConfigSearchPath()[%d] = %q, want %q", i, path, tt.want[i])
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
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

	tests := []struct {
		name       string
		setupFiles map[string]string // filename -> content
		envValue   string
		wantBackup string
		wantUseDir bool
		wantErr    bool
	}{
		{
			name:       "no config files - use defaults",
			setupFiles: map[string]string{},
			envValue:   "",
			wantBackup: "../.bkpfile",
			wantUseDir: true,
			wantErr:    false,
		},
		{
			name: "single config file with custom values",
			setupFiles: map[string]string{
				".bkpfile.yml": `backup_dir_path: "/custom/backup"
use_current_dir_name: false`,
			},
			envValue:   "",
			wantBackup: "/custom/backup",
			wantUseDir: false,
			wantErr:    false,
		},
		{
			name: "multiple config files - precedence test",
			setupFiles: map[string]string{
				"first.yml": `backup_dir_path: "/first/backup"
use_current_dir_name: false`,
				"second.yml": `backup_dir_path: "/second/backup"
use_current_dir_name: true`,
			},
			envValue:   "first.yml:second.yml",
			wantBackup: "/first/backup",
			wantUseDir: false,
			wantErr:    false,
		},
		{
			name: ".bkpfile.yml file",
			setupFiles: map[string]string{
				".bkpfile.yml": `backup_dir_path: "/standard/backup"
use_current_dir_name: true`,
			},
			envValue:   "",
			wantBackup: "/standard/backup",
			wantUseDir: true,
			wantErr:    false,
		},
		{
			name: "invalid YAML",
			setupFiles: map[string]string{
				".bkpfile.yml": `invalid: yaml: content: [`,
			},
			envValue: "",
			wantErr:  true,
		},
		{
			name: "home directory expansion in backup path",
			setupFiles: map[string]string{
				"home.yml": `backup_dir_path: "~/test-backup"
use_current_dir_name: false`,
			},
			envValue:   "home.yml",
			wantBackup: expandHomeDir("~/test-backup"),
			wantUseDir: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp directory
			entries, _ := os.ReadDir(tmpDir)
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}

			// Set up test files
			for filename, content := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("BKPFILE_CONFIG", tt.envValue)
			} else {
				os.Unsetenv("BKPFILE_CONFIG")
			}

			// Load configuration
			cfg, err := LoadConfig(tmpDir)

			if tt.wantErr {
				if err == nil {
					t.Error("LoadConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadConfig() unexpected error: %v", err)
				return
			}

			if cfg.BackupDirPath != tt.wantBackup {
				t.Errorf("LoadConfig().BackupDirPath = %q, want %q", cfg.BackupDirPath, tt.wantBackup)
			}

			if cfg.UseCurrentDirName != tt.wantUseDir {
				t.Errorf("LoadConfig().UseCurrentDirName = %v, want %v", cfg.UseCurrentDirName, tt.wantUseDir)
			}
		})
	}
}

func TestDisplayConfig(t *testing.T) {
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

	tests := []struct {
		name       string
		setupFiles map[string]string // filename -> content
		envValue   string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "default values only",
			setupFiles: map[string]string{},
			envValue:   "",
			wantOutput: `backup_dir_path: ../.bkpfile (source: default)
status_config_error: 10 (source: default)
status_created_backup: 0 (source: default)
status_disk_full: 30 (source: default)
status_failed_to_create_backup_directory: 31 (source: default)
status_file_is_identical_to_existing_backup: 0 (source: default)
status_file_not_found: 20 (source: default)
status_invalid_file_type: 21 (source: default)
status_permission_denied: 22 (source: default)
use_current_dir_name: true (source: default)
`,
			wantErr: false,
		},
		{
			name: "single config file",
			setupFiles: map[string]string{
				".bkpfile.yml": `backup_dir_path: "/custom/backup"
use_current_dir_name: false`,
			},
			envValue: "",
			wantOutput: `backup_dir_path: /custom/backup (source: ./.bkpfile.yml)
status_config_error: 10 (source: default)
status_created_backup: 0 (source: default)
status_disk_full: 30 (source: default)
status_failed_to_create_backup_directory: 31 (source: default)
status_file_is_identical_to_existing_backup: 0 (source: default)
status_file_not_found: 20 (source: default)
status_invalid_file_type: 21 (source: default)
status_permission_denied: 22 (source: default)
use_current_dir_name: false (source: ./.bkpfile.yml)
`,
			wantErr: false,
		},
		{
			name: "multiple config files",
			setupFiles: map[string]string{
				"first.yml": `backup_dir_path: "/first/backup"
use_current_dir_name: false`,
				"second.yml": `backup_dir_path: "/second/backup"
use_current_dir_name: true`,
			},
			envValue: "first.yml:second.yml",
			wantOutput: `backup_dir_path: /first/backup (source: ./first.yml)
status_config_error: 10 (source: default)
status_created_backup: 0 (source: default)
status_disk_full: 30 (source: default)
status_failed_to_create_backup_directory: 31 (source: default)
status_file_is_identical_to_existing_backup: 0 (source: default)
status_file_not_found: 20 (source: default)
status_invalid_file_type: 21 (source: default)
status_permission_denied: 22 (source: default)
use_current_dir_name: false (source: ./first.yml)
`,
			wantErr: false,
		},
		{
			name: "invalid YAML",
			setupFiles: map[string]string{
				".bkpfile.yml": `invalid: yaml: content: [`,
			},
			envValue:   "",
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up temp directory
			entries, _ := os.ReadDir(tmpDir)
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
			}

			// Set up test files
			for filename, content := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("BKPFILE_CONFIG", tt.envValue)
			} else {
				os.Unsetenv("BKPFILE_CONFIG")
			}

			// Change to temp directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run DisplayConfig
			err = DisplayConfig()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			if tt.wantErr {
				if err == nil {
					t.Error("DisplayConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DisplayConfig() unexpected error: %v", err)
				return
			}

			got := buf.String()
			if got != tt.wantOutput {
				t.Errorf("DisplayConfig() output = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

// Helper function to expand home directory for testing
func expandHomeDir(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // fallback to original path
	}
	return filepath.Join(homeDir, path[2:])
}

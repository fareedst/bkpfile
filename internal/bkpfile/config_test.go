package bkpfile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Config != "./.bkpfile.yml:~/.bkpfile.yml" {
		t.Errorf("DefaultConfig().Config = %q, want %q", cfg.Config, "./.bkpfile.yml:~/.bkpfile.yml")
	}

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
			name:     "default paths when env var not set",
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
		wantConfig string
		wantBackup string
		wantUseDir bool
		wantErr    bool
	}{
		{
			name:       "no config files - use defaults",
			setupFiles: map[string]string{},
			envValue:   "",
			wantConfig: "./.bkpfile.yml:~/.bkpfile.yml",
			wantBackup: "../.bkpfile",
			wantUseDir: true,
			wantErr:    false,
		},
		{
			name: "single config file with custom values",
			setupFiles: map[string]string{
				".bkpfile.yml": `config: "custom.yml"
backup_dir_path: "/custom/backup"
use_current_dir_name: false`,
			},
			envValue:   "",
			wantConfig: "custom.yml",
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
			wantConfig: "./.bkpfile.yml:~/.bkpfile.yml",
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
			wantConfig: "./.bkpfile.yml:~/.bkpfile.yml",
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
			wantConfig: "./.bkpfile.yml:~/.bkpfile.yml",
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

			if cfg.Config != tt.wantConfig {
				t.Errorf("LoadConfig().Config = %q, want %q", cfg.Config, tt.wantConfig)
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

func TestConfigurationDiscovery(t *testing.T) {
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
	tmpDir, err := os.MkdirTemp("", "bkpfile-discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple configuration files
	configFiles := map[string]string{
		"global.yml": `backup_dir_path: "/global/backup"
use_current_dir_name: true`,
		"user.yml": `backup_dir_path: "/user/backup"
use_current_dir_name: false`,
		"local.yml": `backup_dir_path: "/local/backup"`,
	}

	for filename, content := range configFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name       string
		envValue   string
		wantBackup string
		wantUseDir bool
	}{
		{
			name:       "global config takes precedence",
			envValue:   "global.yml:user.yml:local.yml",
			wantBackup: "/global/backup",
			wantUseDir: true,
		},
		{
			name:       "user config takes precedence",
			envValue:   "user.yml:global.yml:local.yml",
			wantBackup: "/user/backup",
			wantUseDir: false,
		},
		{
			name:       "local config only",
			envValue:   "local.yml",
			wantBackup: "/local/backup",
			wantUseDir: true, // default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("BKPFILE_CONFIG", tt.envValue)

			// Load configuration
			cfg, err := LoadConfig(tmpDir)
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
	tmpDir, err := os.MkdirTemp("", "bkpfile-display-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	tests := []struct {
		name           string
		setupFiles     map[string]string // filename -> content
		envValue       string
		expectedOutput []string // lines to check for in output
		wantErr        bool
	}{
		{
			name:       "default values only",
			setupFiles: map[string]string{},
			envValue:   "",
			expectedOutput: []string{
				"config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)",
				"backup_dir_path: ../.bkpfile (source: default)",
				"use_current_dir_name: true (source: default)",
			},
			wantErr: false,
		},
		{
			name: "single configuration file",
			setupFiles: map[string]string{
				"test.yml": `backup_dir_path: "/custom/backup"
use_current_dir_name: false`,
			},
			envValue: "test.yml",
			expectedOutput: []string{
				"backup_dir_path: /custom/backup (source: ./test.yml)",
				"config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)",
				"use_current_dir_name: false (source: ./test.yml)",
			},
			wantErr: false,
		},
		{
			name: "multiple configuration files with precedence",
			setupFiles: map[string]string{
				"first.yml": `backup_dir_path: "/first/backup"
use_current_dir_name: false`,
				"second.yml": `backup_dir_path: "/second/backup"
use_current_dir_name: true
config: "custom.yml"`,
			},
			envValue: "first.yml:second.yml",
			expectedOutput: []string{
				"backup_dir_path: /first/backup (source: ./first.yml)",
				"config: custom.yml (source: ./second.yml)",
				"use_current_dir_name: false (source: ./first.yml)",
			},
			wantErr: false,
		},
		{
			name: "home directory expansion",
			setupFiles: map[string]string{
				"home.yml": `backup_dir_path: "~/test-backup"`,
			},
			envValue: "home.yml",
			expectedOutput: []string{
				fmt.Sprintf("backup_dir_path: %s (source: ./home.yml)", expandHomeDir("~/test-backup")),
				"config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)",
				"use_current_dir_name: true (source: default)",
			},
			wantErr: false,
		},
		{
			name: "missing configuration files",
			setupFiles: map[string]string{
				"exists.yml": `backup_dir_path: "/exists/backup"`,
			},
			envValue: "missing.yml:exists.yml",
			expectedOutput: []string{
				"backup_dir_path: /exists/backup (source: ./exists.yml)",
				"use_current_dir_name: true (source: default)",
			},
			wantErr: false,
		},
		{
			name: "invalid YAML format",
			setupFiles: map[string]string{
				"invalid.yml": `invalid: yaml: content: [`,
			},
			envValue: "invalid.yml",
			wantErr:  true,
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

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run DisplayConfig
			err := DisplayConfig()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Errorf("DisplayConfig() expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("DisplayConfig() unexpected error: %v", err)
				return
			}

			// Check expected output if no error expected
			if !tt.wantErr {
				for _, expectedLine := range tt.expectedOutput {
					if !strings.Contains(output, expectedLine) {
						t.Errorf("DisplayConfig() output missing expected line: %q\nGot output:\n%s", expectedLine, output)
					}
				}
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

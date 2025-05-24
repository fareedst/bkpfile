package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionFlag(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a new command for testing
	cmd := &cobra.Command{
		Use:     "bkpfile",
		Version: fmt.Sprintf("%s (compiled %s) [%s]", version, compileDate, platform),
	}

	// Set up the version flag
	cmd.SetVersionTemplate("{{.Version}}\n")

	// Execute the version command
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}

	// Close the write end of the pipe
	w.Close()

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check if the output matches the version format
	expectedPrefix := version
	if !bytes.Contains(buf.Bytes(), []byte(expectedPrefix)) {
		t.Errorf("Version output = %q, want prefix %q", output, expectedPrefix)
	}
}

func TestHelpScreenVersion(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a new command for testing
	cmd := &cobra.Command{
		Use:     "bkpfile",
		Version: fmt.Sprintf("%s (compiled %s) [%s]", version, compileDate, platform),
	}

	// Set the same help template as in main.go
	cmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
Version: {{.Version}}
`)

	// Execute the help command
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to execute help command: %v", err)
	}

	// Close the write end of the pipe
	w.Close()

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Check if the help output contains the version
	expectedVersion := fmt.Sprintf("Version: %s (compiled %s) [%s]", version, compileDate, platform)
	if !bytes.Contains(buf.Bytes(), []byte(expectedVersion)) {
		t.Errorf("Help output does not contain version line %q", expectedVersion)
	}
}

func TestConfigFlag(t *testing.T) {
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
	tmpDir, err := os.MkdirTemp("", "bkpfile-config-flag-test-*")
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
			name:       "config flag with default values",
			setupFiles: map[string]string{},
			envValue:   "",
			expectedOutput: []string{
				"backup_dir_path: ../.bkpfile (source: default)",
				"use_current_dir_name: true (source: default)",
				"config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)",
			},
			wantErr: false,
		},
		{
			name: "config flag with custom configuration file",
			setupFiles: map[string]string{
				"custom.yml": `backup_dir_path: "/tmp/custom-backup"
use_current_dir_name: false`,
			},
			envValue: "custom.yml",
			expectedOutput: []string{
				"backup_dir_path: /tmp/custom-backup (source: ./custom.yml)",
				"use_current_dir_name: false (source: ./custom.yml)",
			},
			wantErr: false,
		},
		{
			name: "config flag with multiple configuration files",
			setupFiles: map[string]string{
				"primary.yml": `backup_dir_path: "/tmp/primary"`,
				"secondary.yml": `backup_dir_path: "/tmp/secondary"
use_current_dir_name: false
config: "alternate.yml"`,
			},
			envValue: "primary.yml:secondary.yml",
			expectedOutput: []string{
				"backup_dir_path: /tmp/primary (source: ./primary.yml)",
				"use_current_dir_name: false (source: ./secondary.yml)",
				"config: alternate.yml (source: ./secondary.yml)",
			},
			wantErr: false,
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

			// Reset command state
			config = false
			list = false
			dryRun = false

			// Save original stdout
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()

			// Create a pipe to capture stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command with --config flag
			rootCmd.SetArgs([]string{"--config"})
			err := rootCmd.Execute()

			// Close the write end of the pipe
			w.Close()
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check expected output if no error expected
			if !tt.wantErr {
				for _, expectedLine := range tt.expectedOutput {
					if !strings.Contains(output, expectedLine) {
						t.Errorf("Output missing expected line: %q\nGot output:\n%s", expectedLine, output)
					}
				}
			}
		})
	}
}

func TestConfigFlagWithOtherArgs(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bkpfile-config-args-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Reset command state
	config = false
	list = false
	dryRun = false

	// Save original stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command with --config flag and other arguments (should ignore other args)
	rootCmd.SetArgs([]string{"--config", "somefile.txt", "note"})
	err = rootCmd.Execute()

	// Close the write end of the pipe
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not error and should show config output
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should contain default configuration output
	expectedLines := []string{
		"backup_dir_path: ../.bkpfile (source: default)",
		"use_current_dir_name: true (source: default)",
		"config: ./.bkpfile.yml:~/.bkpfile.yml (source: default)",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(output, expectedLine) {
			t.Errorf("Output missing expected line: %q\nGot output:\n%s", expectedLine, output)
		}
	}
}

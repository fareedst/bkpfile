package main

import (
	"bytes"
	"fmt"
	"os"
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

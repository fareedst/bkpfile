package main

import (
	"fmt"
	"os"
	"path/filepath"

	"bkpfile/internal/bkpfile"

	"github.com/spf13/cobra"
)

var (
	// Version information
	version     = "1.0.0"
	compileDate string
	platform    string

	// Global flags
	dryRun bool
	list   bool
	config bool

	// Root command
	rootCmd = &cobra.Command{
		Use:     "bkpfile [FILE_PATH] [NOTE]",
		Short:   "Single file backup CLI application",
		Long:    `A command-line application for creating and managing file backups.`,
		Version: fmt.Sprintf("%s (compiled %s) [%s]", version, compileDate, platform),
		Args:    cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle config flag first (exclusive operation)
			if config {
				return bkpfile.DisplayConfig()
			}

			// Require at least one argument for other operations
			if len(args) < 1 {
				return fmt.Errorf("file path is required")
			}

			filePath := args[0]
			note := ""
			if len(args) > 1 {
				note = args[1]
			}

			// Load configuration
			cfg, err := bkpfile.LoadConfig(".")
			if err != nil {
				// Configuration error should use the config error status code
				// Use default config to get the status code since loading failed
				defaultCfg := bkpfile.DefaultConfig()
				fmt.Fprintln(os.Stderr, fmt.Sprintf("failed to load config: %v", err))
				os.Exit(defaultCfg.StatusConfigError)
			}

			if list {
				// List backups
				backups, err := bkpfile.ListBackups(cfg.BackupDirPath, filePath)
				if err != nil {
					fmt.Fprintln(os.Stderr, fmt.Sprintf("failed to list backups: %v", err))
					os.Exit(cfg.StatusConfigError)
				}

				// Display backups
				for _, backup := range backups {
					// Get relative path for display
					relPath, err := filepath.Rel(".", backup.Path)
					if err != nil {
						relPath = backup.Path // Fallback to absolute path if relative path fails
					}
					fmt.Printf("%s (created: %s)\n", relPath, backup.CreationTime.Format("2006-01-02 15:04:05"))
				}
				return nil
			}

			// Create backup
			err = bkpfile.CreateBackup(cfg, filePath, note, dryRun)
			if err != nil {
				// Check if it's a BackupError with a status code
				if backupErr, ok := err.(*bkpfile.BackupError); ok {
					// Don't print error message for successful operations
					// Success operations are: backup created successfully, dry run completed, file is identical
					isSuccess := backupErr.Message == "backup created successfully" ||
						backupErr.Message == "dry run completed" ||
						backupErr.Message == "file is identical to existing backup"

					if !isSuccess {
						fmt.Fprintln(os.Stderr, backupErr.Message)
					}
					os.Exit(backupErr.StatusCode)
				}
				// For other errors, return them normally
				return err
			}
			return nil
		},
	}
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without creating backups")
	rootCmd.PersistentFlags().BoolVar(&list, "list", false, "List all backups for the specified file")
	rootCmd.PersistentFlags().BoolVar(&config, "config", false, "Display computed configuration values and exit")

	// Customize help template to include version
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
Version: {{.Version}}
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's a BackupError with a status code
		if backupErr, ok := err.(*bkpfile.BackupError); ok {
			// Don't print error message for successful operations
			// Success operations are: backup created successfully, dry run completed, file is identical
			isSuccess := backupErr.Message == "backup created successfully" ||
				backupErr.Message == "dry run completed" ||
				backupErr.Message == "file is identical to existing backup"

			if !isSuccess {
				fmt.Fprintln(os.Stderr, backupErr.Message)
			}
			os.Exit(backupErr.StatusCode)
		}

		// For other errors, print and exit with status 1
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

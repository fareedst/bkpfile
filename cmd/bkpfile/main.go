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

	// Root command
	rootCmd = &cobra.Command{
		Use:     "bkpfile [FILE_PATH] [NOTE]",
		Short:   "Single file backup CLI application",
		Long:    `A command-line application for creating and managing file backups.`,
		Version: fmt.Sprintf("%s (compiled %s) [%s]", version, compileDate, platform),
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			note := ""
			if len(args) > 1 {
				note = args[1]
			}

			// Load configuration
			cfg, err := bkpfile.LoadConfig(".")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if list {
				// List backups
				backups, err := bkpfile.ListBackups(cfg.BackupDirPath, filePath)
				if err != nil {
					return fmt.Errorf("failed to list backups: %w", err)
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
			return bkpfile.CreateBackup(cfg, filePath, note, dryRun)
		},
	}
)

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without creating backups")
	rootCmd.PersistentFlags().BoolVar(&list, "list", false, "List all backups for the specified file")

	// Customize help template to include version
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
Version: {{.Version}}
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

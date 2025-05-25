package bkpfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
// Architecture: Data Objects - Config
type Config struct {
	// BackupDirPath specifies where backups are stored
	// Architecture: Config.BackupDirPath
	BackupDirPath string `yaml:"backup_dir_path"`

	// UseCurrentDirName controls whether to include current directory name in backup path
	// Architecture: Config.UseCurrentDirName
	UseCurrentDirName bool `yaml:"use_current_dir_name"`

	// Status code configuration fields
	// Architecture: Config.StatusCreatedBackup
	StatusCreatedBackup int `yaml:"status_created_backup"`

	// Architecture: Config.StatusFailedToCreateBackupDirectory
	StatusFailedToCreateBackupDirectory int `yaml:"status_failed_to_create_backup_directory"`

	// Architecture: Config.StatusFileIsIdenticalToExistingBackup
	StatusFileIsIdenticalToExistingBackup int `yaml:"status_file_is_identical_to_existing_backup"`

	// Architecture: Config.StatusFileNotFound
	StatusFileNotFound int `yaml:"status_file_not_found"`

	// Architecture: Config.StatusInvalidFileType
	StatusInvalidFileType int `yaml:"status_invalid_file_type"`

	// Architecture: Config.StatusPermissionDenied
	StatusPermissionDenied int `yaml:"status_permission_denied"`

	// Architecture: Config.StatusDiskFull
	StatusDiskFull int `yaml:"status_disk_full"`

	// Architecture: Config.StatusConfigError
	StatusConfigError int `yaml:"status_config_error"`
}

// ConfigValue represents a configuration parameter with its computed value and source
// Architecture: Data Objects - ConfigValue
type ConfigValue struct {
	// Name is the configuration parameter name
	// Architecture: ConfigValue.Name
	Name string

	// Value is the computed configuration value including defaults
	// Architecture: ConfigValue.Value
	Value string

	// Source is the source file path or "default" for default values
	// Architecture: ConfigValue.Source
	Source string
}

// DefaultConfig creates a new Config with default values
// Architecture: Core Functions - Configuration Management - DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		BackupDirPath:                         "../.bkpfile",
		StatusConfigError:                     10,
		StatusCreatedBackup:                   0,
		StatusDiskFull:                        30,
		StatusFailedToCreateBackupDirectory:   31,
		StatusFileIsIdenticalToExistingBackup: 0,
		StatusFileNotFound:                    20,
		StatusInvalidFileType:                 21,
		StatusPermissionDenied:                22,
		UseCurrentDirName:                     true,
	}
}

// GetConfigSearchPath returns the list of configuration file paths to search
// Architecture: Core Functions - Configuration Management - GetConfigSearchPath
func GetConfigSearchPath() []string {
	// Read BKPFILE_CONFIG environment variable
	envConfig := os.Getenv("BKPFILE_CONFIG")

	var paths []string
	if envConfig != "" {
		// Split on colon to get path list
		paths = strings.Split(envConfig, ":")
	} else {
		// Use hard-coded default path list
		paths = []string{"./.bkpfile.yml", "~/.bkpfile.yml"}
	}

	// Expand home directory in paths
	for i, path := range paths {
		if strings.HasPrefix(path, "~/") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				paths[i] = filepath.Join(homeDir, path[2:])
			}
		}
	}

	return paths
}

// findConfigValueIndex returns the index of the config value with the given name
// Returns -1 if not found
func findConfigValueIndex(configValues []ConfigValue, name string) int {
	for i, cv := range configValues {
		if cv.Name == name {
			return i
		}
	}
	return -1
}

// DisplayConfig displays computed configuration values and exits
// Architecture: Core Functions - Configuration Management - DisplayConfig
func DisplayConfig() error {
	// Get configuration search paths
	searchPaths := GetConfigSearchPath()

	// Initialize with default values and track sources
	defaultCfg := DefaultConfig()
	configValues := []ConfigValue{
		{Name: "backup_dir_path", Value: defaultCfg.BackupDirPath, Source: "default"},
		{Name: "status_config_error", Value: fmt.Sprintf("%d", defaultCfg.StatusConfigError), Source: "default"},
		{Name: "status_created_backup", Value: fmt.Sprintf("%d", defaultCfg.StatusCreatedBackup), Source: "default"},
		{Name: "status_disk_full", Value: fmt.Sprintf("%d", defaultCfg.StatusDiskFull), Source: "default"},
		{Name: "status_failed_to_create_backup_directory", Value: fmt.Sprintf("%d", defaultCfg.StatusFailedToCreateBackupDirectory), Source: "default"},
		{Name: "status_file_is_identical_to_existing_backup", Value: fmt.Sprintf("%d", defaultCfg.StatusFileIsIdenticalToExistingBackup), Source: "default"},
		{Name: "status_file_not_found", Value: fmt.Sprintf("%d", defaultCfg.StatusFileNotFound), Source: "default"},
		{Name: "status_invalid_file_type", Value: fmt.Sprintf("%d", defaultCfg.StatusInvalidFileType), Source: "default"},
		{Name: "status_permission_denied", Value: fmt.Sprintf("%d", defaultCfg.StatusPermissionDenied), Source: "default"},
		{Name: "use_current_dir_name", Value: fmt.Sprintf("%t", defaultCfg.UseCurrentDirName), Source: "default"},
	}

	// Process configuration files in order with precedence rules
	for _, configPath := range searchPaths {
		// Store original path for source display
		originalPath := configPath

		// Handle relative paths by resolving them relative to current directory
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(".", configPath)
			// Update original path to include ./ prefix for relative paths
			if !strings.HasPrefix(originalPath, "./") && !strings.HasPrefix(originalPath, "/") {
				originalPath = "./" + originalPath
			}
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Read config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}

		// Parse YAML into a map to check which fields are actually set
		var yamlData map[string]interface{}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
		}

		// Parse YAML into a temporary config
		tempCfg := &Config{}
		if err := yaml.Unmarshal(data, tempCfg); err != nil {
			return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
		}

		// Update configuration values with source tracking (earlier files take precedence)
		if _, exists := yamlData["backup_dir_path"]; exists && tempCfg.BackupDirPath != "" {
			// Expand home directory in backup path
			backupPath := tempCfg.BackupDirPath
			if strings.HasPrefix(backupPath, "~/") {
				if homeDir, err := os.UserHomeDir(); err == nil {
					backupPath = filepath.Join(homeDir, backupPath[2:])
				}
			}
			// Update only if not already set by a previous (higher precedence) file
			if idx := findConfigValueIndex(configValues, "backup_dir_path"); idx >= 0 && configValues[idx].Source == "default" {
				configValues[idx].Value = backupPath
				configValues[idx].Source = originalPath
			}
		}

		if _, exists := yamlData["use_current_dir_name"]; exists {
			// Update only if not already set by a previous (higher precedence) file
			if idx := findConfigValueIndex(configValues, "use_current_dir_name"); idx >= 0 && configValues[idx].Source == "default" {
				configValues[idx].Value = fmt.Sprintf("%t", tempCfg.UseCurrentDirName)
				configValues[idx].Source = originalPath
			}
		}

		// Handle status code configuration fields
		statusFields := []struct {
			yamlKey string
			value   int
		}{
			{"status_config_error", tempCfg.StatusConfigError},
			{"status_created_backup", tempCfg.StatusCreatedBackup},
			{"status_disk_full", tempCfg.StatusDiskFull},
			{"status_failed_to_create_backup_directory", tempCfg.StatusFailedToCreateBackupDirectory},
			{"status_file_is_identical_to_existing_backup", tempCfg.StatusFileIsIdenticalToExistingBackup},
			{"status_file_not_found", tempCfg.StatusFileNotFound},
			{"status_invalid_file_type", tempCfg.StatusInvalidFileType},
			{"status_permission_denied", tempCfg.StatusPermissionDenied},
		}

		for _, field := range statusFields {
			if _, exists := yamlData[field.yamlKey]; exists {
				// Update only if not already set by a previous (higher precedence) file
				if idx := findConfigValueIndex(configValues, field.yamlKey); idx >= 0 && configValues[idx].Source == "default" {
					configValues[idx].Value = fmt.Sprintf("%d", field.value)
					configValues[idx].Source = originalPath
				}
			}
		}
	}

	// Display each configuration value with name, computed value, and source
	for _, cv := range configValues {
		fmt.Printf("%s: %s (source: %s)\n", cv.Name, cv.Value, cv.Source)
	}

	return nil
}

// LoadConfig loads configuration from YAML files using discovery path or returns default config
// Architecture: Core Functions - Configuration Management - LoadConfig
func LoadConfig(root string) (*Config, error) {
	cfg := DefaultConfig()

	// Get configuration search paths
	searchPaths := GetConfigSearchPath()

	// Process configuration files in order (earlier files take precedence)
	foundConfig := false
	for _, configPath := range searchPaths {
		// Handle relative paths by resolving them relative to root
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(root, configPath)
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Read config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		// Parse YAML into a map to check which fields are actually set
		var yamlData map[string]interface{}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return nil, err
		}

		// Parse YAML into a temporary config
		tempCfg := &Config{}
		if err := yaml.Unmarshal(data, tempCfg); err != nil {
			return nil, err
		}

		// Merge configuration with precedence (only update fields that are explicitly set)
		if !foundConfig {
			// First config file found sets values for fields that are explicitly present
			if _, exists := yamlData["backup_dir_path"]; exists && tempCfg.BackupDirPath != "" {
				// Expand home directory in backup path
				backupPath := tempCfg.BackupDirPath
				if strings.HasPrefix(backupPath, "~/") {
					if homeDir, err := os.UserHomeDir(); err == nil {
						backupPath = filepath.Join(homeDir, backupPath[2:])
					}
				}
				cfg.BackupDirPath = backupPath
			}
			if _, exists := yamlData["use_current_dir_name"]; exists {
				cfg.UseCurrentDirName = tempCfg.UseCurrentDirName
			}

			// Handle status code configuration fields
			if _, exists := yamlData["status_created_backup"]; exists {
				cfg.StatusCreatedBackup = tempCfg.StatusCreatedBackup
			}
			if _, exists := yamlData["status_failed_to_create_backup_directory"]; exists {
				cfg.StatusFailedToCreateBackupDirectory = tempCfg.StatusFailedToCreateBackupDirectory
			}
			if _, exists := yamlData["status_file_is_identical_to_existing_backup"]; exists {
				cfg.StatusFileIsIdenticalToExistingBackup = tempCfg.StatusFileIsIdenticalToExistingBackup
			}
			if _, exists := yamlData["status_file_not_found"]; exists {
				cfg.StatusFileNotFound = tempCfg.StatusFileNotFound
			}
			if _, exists := yamlData["status_invalid_file_type"]; exists {
				cfg.StatusInvalidFileType = tempCfg.StatusInvalidFileType
			}
			if _, exists := yamlData["status_permission_denied"]; exists {
				cfg.StatusPermissionDenied = tempCfg.StatusPermissionDenied
			}
			if _, exists := yamlData["status_disk_full"]; exists {
				cfg.StatusDiskFull = tempCfg.StatusDiskFull
			}
			if _, exists := yamlData["status_config_error"]; exists {
				cfg.StatusConfigError = tempCfg.StatusConfigError
			}

			foundConfig = true
		} else {
			// Subsequent config files only override if the field is explicitly set
			// For this simple implementation, we'll take the first config found
			// since we want earlier files to take precedence
		}
	}

	// Check for .bkpfile.yml file in the root directory
	if !foundConfig {
		configPath := filepath.Join(root, ".bkpfile.yml")
		if _, err := os.Stat(configPath); err == nil {
			// Read config file
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, err
			}

			// Parse YAML into a map to check which fields are actually set
			var yamlData map[string]interface{}
			if err := yaml.Unmarshal(data, &yamlData); err != nil {
				return nil, err
			}

			// Parse YAML into a temporary config
			tempCfg := &Config{}
			if err := yaml.Unmarshal(data, tempCfg); err != nil {
				return nil, err
			}

			// Merge with defaults, only overriding explicitly set fields
			if _, exists := yamlData["backup_dir_path"]; exists && tempCfg.BackupDirPath != "" {
				// Expand home directory in backup path
				backupPath := tempCfg.BackupDirPath
				if strings.HasPrefix(backupPath, "~/") {
					if homeDir, err := os.UserHomeDir(); err == nil {
						backupPath = filepath.Join(homeDir, backupPath[2:])
					}
				}
				cfg.BackupDirPath = backupPath
			}
			if _, exists := yamlData["use_current_dir_name"]; exists {
				cfg.UseCurrentDirName = tempCfg.UseCurrentDirName
			}

			// Handle status code configuration fields
			if _, exists := yamlData["status_created_backup"]; exists {
				cfg.StatusCreatedBackup = tempCfg.StatusCreatedBackup
			}
			if _, exists := yamlData["status_failed_to_create_backup_directory"]; exists {
				cfg.StatusFailedToCreateBackupDirectory = tempCfg.StatusFailedToCreateBackupDirectory
			}
			if _, exists := yamlData["status_file_is_identical_to_existing_backup"]; exists {
				cfg.StatusFileIsIdenticalToExistingBackup = tempCfg.StatusFileIsIdenticalToExistingBackup
			}
			if _, exists := yamlData["status_file_not_found"]; exists {
				cfg.StatusFileNotFound = tempCfg.StatusFileNotFound
			}
			if _, exists := yamlData["status_invalid_file_type"]; exists {
				cfg.StatusInvalidFileType = tempCfg.StatusInvalidFileType
			}
			if _, exists := yamlData["status_permission_denied"]; exists {
				cfg.StatusPermissionDenied = tempCfg.StatusPermissionDenied
			}
			if _, exists := yamlData["status_disk_full"]; exists {
				cfg.StatusDiskFull = tempCfg.StatusDiskFull
			}
			if _, exists := yamlData["status_config_error"]; exists {
				cfg.StatusConfigError = tempCfg.StatusConfigError
			}

			foundConfig = true
		}
	}

	return cfg, nil
}

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
	// Config specifies the colon-separated list of configuration file paths to search
	// Architecture: Config.Config
	Config string `yaml:"config"`

	// BackupDirPath specifies where backups are stored
	// Architecture: Config.BackupDirPath
	BackupDirPath string `yaml:"backup_dir_path"`

	// UseCurrentDirName controls whether to include current directory name in backup path
	// Architecture: Config.UseCurrentDirName
	UseCurrentDirName bool `yaml:"use_current_dir_name"`
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
		Config:            "./.bkpfile.yml:~/.bkpfile.yml",
		BackupDirPath:     "../.bkpfile",
		UseCurrentDirName: true,
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
		// Use default path list
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

// DisplayConfig displays computed configuration values and exits
// Architecture: Core Functions - Configuration Management - DisplayConfig
func DisplayConfig() error {
	// Get configuration search paths
	searchPaths := GetConfigSearchPath()

	// Initialize with default values and track sources
	defaultCfg := DefaultConfig()
	configValues := []ConfigValue{
		{Name: "backup_dir_path", Value: defaultCfg.BackupDirPath, Source: "default"},
		{Name: "use_current_dir_name", Value: fmt.Sprintf("%t", defaultCfg.UseCurrentDirName), Source: "default"},
		{Name: "config", Value: defaultCfg.Config, Source: "default"},
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
			if configValues[0].Source == "default" {
				configValues[0].Value = backupPath
				configValues[0].Source = originalPath
			}
		}

		if _, exists := yamlData["use_current_dir_name"]; exists {
			// Update only if not already set by a previous (higher precedence) file
			if configValues[1].Source == "default" {
				configValues[1].Value = fmt.Sprintf("%t", tempCfg.UseCurrentDirName)
				configValues[1].Source = originalPath
			}
		}

		if _, exists := yamlData["config"]; exists && tempCfg.Config != "" {
			// Update only if not already set by a previous (higher precedence) file
			if configValues[2].Source == "default" {
				configValues[2].Value = tempCfg.Config
				configValues[2].Source = originalPath
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
			if _, exists := yamlData["config"]; exists && tempCfg.Config != "" {
				cfg.Config = tempCfg.Config
			}
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
			foundConfig = true
		}
	}

	return cfg, nil
}

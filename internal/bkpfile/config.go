package bkpfile

import (
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

package bkpfile

import (
	"os"
	"path/filepath"

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
}

// DefaultConfig creates a new Config with default values
// Architecture: Core Functions - Configuration Management - DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		BackupDirPath:     "../.bkpfile",
		UseCurrentDirName: true,
	}
}

// LoadConfig loads configuration from YAML file or returns default config
// Architecture: Core Functions - Configuration Management - LoadConfig
func LoadConfig(root string) (*Config, error) {
	configPath := filepath.Join(root, ".bkpfile.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

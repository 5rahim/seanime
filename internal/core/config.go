package core

import (
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	Server struct {
		Host string
		Port int
	}
	Web struct {
		Host string
		Port int
	}
	Database struct {
		Name string
	}
	Data struct {
		AppDataDir string
	}
}

type ConfigOptions struct {
	DataDirPath string
}

var DefaultConfig = ConfigOptions{
	DataDirPath: "",
}

// NewConfig initializes the config, checks if the config file exists, and generates a default one if not.
func NewConfig(options *ConfigOptions) (*Config, error) {

	// Get the user's config path
	configPath, appDataDir, err := getUserConfigPath(options.DataDirPath)

	if err != nil {
		return nil, err
	}

	// Set the config file name and type
	viper.SetConfigName(constants.ConfigFileName)
	viper.SetConfigType("json")
	viper.SetConfigFile(configPath)

	// Check if the config file exists, and generate a default one if not
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		cfg := &Config{
			Server: struct {
				Host string
				Port int
			}{
				Host: "127.0.0.1",
				Port: 43210,
			},
			Web: struct {
				Host string
				Port int
			}{
				Host: "127.0.0.1",
				Port: 43211,
			},
			Database: struct {
				Name string
			}{
				Name: "seanime",
			},
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
			return nil, err
		}

		if err := cfg.saveConfigToFile(); err != nil {
			return nil, err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.Data.AppDataDir = appDataDir

	return cfg, nil
}

// saveConfigToFile saves the config to the config file.
func (cfg *Config) saveConfigToFile() error {
	viper.Set("server.host", cfg.Server.Host)
	viper.Set("server.port", cfg.Server.Port)
	viper.Set("web.host", cfg.Web.Host)
	viper.Set("web.port", cfg.Web.Port)
	viper.Set("database.name", cfg.Database.Name)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// getUserConfigPath returns the path to the user's config file and the app data directory.
func getUserConfigPath(_dataDir string) (string, string, error) {
	dataDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}

	// Override the data directory if one is provided
	if _dataDir != "" {
		dataDir = _dataDir
	}

	// Get the app directory
	appDataDir := filepath.Join(dataDir, "Seanime")
	// Create the app directory if it doesn't exist
	if err := os.MkdirAll(appDataDir, 0700); err != nil {
		return "", "", err
	}

	return filepath.Join(appDataDir, constants.ConfigFileName), appDataDir, nil
}

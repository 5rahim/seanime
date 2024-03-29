package core

import (
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type Config struct {
	Version string
	Server  struct {
		Host string
		Port int
	}
	Database struct {
		Name string
	}
	Web struct {
		Dir      string
		AssetDir string
	}
	Logs struct {
		Dir string
	}
	Cache struct {
		Dir string
	}
	Manga struct {
		Enabled bool
	}
	Data struct { // Hydrated after config is loaded
		AppDataDir string
	}
}

var defaultConfigValues = Config{
	Version: constants.Version,
	Server: struct {
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
	Web: struct {
		Dir      string
		AssetDir string
	}{
		Dir:      "$SEANIME_WORKING_DIR/web",
		AssetDir: "$SEANIME_DATA_DIR/assets",
	},
	Cache: struct {
		Dir string
	}{
		Dir: "$SEANIME_DATA_DIR/cache",
	},
	Manga: struct {
		Enabled bool
	}{
		Enabled: false,
	},
	Logs: struct {
		Dir string
	}{
		Dir: "$SEANIME_DATA_DIR/logs",
	},
}

type ConfigOptions struct {
	DataDirPath string // The path to the Seanime data directory, if any
}

var DefaultConfig = ConfigOptions{
	DataDirPath: "",
}

// NewConfig initializes the config, checks if the config file exists, and generates a default one if not.
func NewConfig(options *ConfigOptions) (*Config, error) {

	// Get the user's config path
	configPath, dataDir, err := getUserPaths(options.DataDirPath)
	if err != nil {
		return nil, err
	}

	// Set the app data directory environment variable
	if os.Getenv("SEANIME_DATA_DIR") == "" {
		if err = os.Setenv("SEANIME_DATA_DIR", dataDir); err != nil {
			return nil, err
		}
	}
	if os.Getenv("SEANIME_WORKING_DIR") == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		if err = os.Setenv("SEANIME_WORKING_DIR", filepath.FromSlash(wd)); err != nil {
			return nil, err
		}
	}

	// Set the config file name and type
	viper.SetConfigName(constants.ConfigFileName)
	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)

	// Set the default values
	viper.SetDefault("version", defaultConfigValues.Version)
	viper.SetDefault("server.host", defaultConfigValues.Server.Host)
	viper.SetDefault("server.port", defaultConfigValues.Server.Port)
	viper.SetDefault("database.name", defaultConfigValues.Database.Name)
	viper.SetDefault("web.dir", defaultConfigValues.Web.Dir)
	viper.SetDefault("web.assetDir", defaultConfigValues.Web.AssetDir)
	viper.SetDefault("cache.dir", defaultConfigValues.Cache.Dir)
	viper.SetDefault("manga.enabled", defaultConfigValues.Manga.Enabled)
	viper.SetDefault("logs.dir", defaultConfigValues.Logs.Dir)

	// Check if the config file exists, and generate a default one if not
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		cfg := &defaultConfigValues

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

	// Update the config if the version has changed
	updateVersion(cfg)

	// Save the values to the config file
	if err := cfg.saveConfigToFile(); err != nil {
		return nil, err
	}

	// Hydrate the config values
	hydrateValues(cfg)

	cfg.Data.AppDataDir = dataDir

	// Check validity of the config
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func validateConfig(cfg *Config) error {
	if cfg.Server.Host == "" {
		return errInvalidConfigValue("server.host", "cannot be empty")
	}
	if cfg.Server.Port == 0 {
		return errInvalidConfigValue("server.port", "cannot be 0")
	}
	if cfg.Database.Name == "" {
		return errInvalidConfigValue("database.name", "cannot be empty")
	}
	if cfg.Web.Dir == "" {
		return errInvalidConfigValue("web.dir", "cannot be empty")
	}
	if cfg.Web.AssetDir == "" {
		return errInvalidConfigValue("web.assetDir", "cannot be empty")
	}
	if cfg.Cache.Dir == "" {
		return errInvalidConfigValue("cache.dir", "cannot be empty")
	}
	if cfg.Logs.Dir == "" {
		return errInvalidConfigValue("logs.dir", "cannot be empty")
	}

	return nil
}

func errInvalidConfigValue(s string, s2 string) error {
	return errors.New(fmt.Sprintf("invalid config value: \"%s\" %s", s, s2))
}

func updateVersion(cfg *Config) {
	defer func() {
		if r := recover(); r != nil {
			// Do nothing
		}
	}()

	if cfg.Version != constants.Version {
		cfg.Version = constants.Version
	}
}

func hydrateValues(cfg *Config) {
	defer func() {
		if r := recover(); r != nil {
			// Do nothing
		}
	}()
	cfg.Web.AssetDir = filepath.FromSlash(os.ExpandEnv(cfg.Web.AssetDir))
	cfg.Web.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Web.Dir))
	cfg.Cache.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Cache.Dir))
	cfg.Logs.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Logs.Dir))
}

// saveConfigToFile saves the config to the config file.
func (cfg *Config) saveConfigToFile() error {
	viper.Set("version", constants.Version)
	viper.Set("server.host", cfg.Server.Host)
	viper.Set("server.port", cfg.Server.Port)
	viper.Set("database.name", cfg.Database.Name)
	viper.Set("web.dir", cfg.Web.Dir)
	viper.Set("web.assetDir", cfg.Web.AssetDir)
	viper.Set("cache.dir", cfg.Cache.Dir)
	viper.Set("logs.dir", cfg.Logs.Dir)
	viper.Set("manga.enabled", cfg.Manga.Enabled)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// getUserPaths returns the path to the user's config file and the app data directory.
//   - configPath: The path to the user's config file
//   - dataDir: The path to the Seanime data directory
func getUserPaths(_definedDataDir string) (configPath string, dataDir string, err error) {
	// DEVNOTE: We can use environment variables to override the data directory if needed
	dataDir, err = os.UserConfigDir()
	if err != nil {
		return "", "", err
	}

	// Get the app directory
	dataDir = filepath.Join(dataDir, "Seanime")

	// Override the Seanime data directory if one is provided
	if _definedDataDir != "" {
		dataDir = _definedDataDir
	}

	// Create the app directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", "", err
	}

	configPath = filepath.FromSlash(filepath.Join(dataDir, constants.ConfigFileName))
	dataDir = filepath.FromSlash(dataDir)

	return configPath, dataDir, nil
}

package core

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Version string
	Server  struct {
		Host    string
		Port    int
		Offline bool
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
	Offline struct {
		Dir      string
		AssetDir string
	}
	Manga struct {
		BackupDir   string
		DownloadDir string
	}
	Data struct { // Hydrated after config is loaded
		AppDataDir string
		WorkingDir string
	}
}

type ConfigOptions struct {
	DataDir         string // The path to the Seanime data directory, if any
	OnVersionChange []func(oldVersion string, newVersion string)
	TrueWd          bool
}

// NewConfig initializes the config
func NewConfig(options *ConfigOptions, logger *zerolog.Logger) (*Config, error) {

	logger.Debug().Msg("app: Initializing config")

	// Set Seanime's environment variables
	if os.Getenv("SEANIME_DATA_DIR") != "" {
		options.DataDir = os.Getenv("SEANIME_DATA_DIR")
	}

	defaultHost := "127.0.0.1"
	defaultPort := 43211

	if os.Getenv("SEANIME_SERVER_HOST") != "" {
		defaultHost = os.Getenv("SEANIME_SERVER_HOST")
	}
	if os.Getenv("SEANIME_SERVER_PORT") != "" {
		var err error
		defaultPort, err = strconv.Atoi(os.Getenv("SEANIME_SERVER_PORT"))
		if err != nil {
			return nil, fmt.Errorf("invalid SEANIME_SERVER_PORT environment variable: %s", os.Getenv("SEANIME_SERVER_PORT"))
		}
	}

	// Initialize the app data directory
	dataDir, configPath, err := initAppDataDir(options.DataDir, logger)
	if err != nil {
		return nil, err
	}

	// Set Seanime's default custom environment variables
	if err = setDefaultEnvironmentVariables(dataDir, options.TrueWd); err != nil {
		return nil, err
	}

	// Configure viper
	viper.SetConfigName(constants.ConfigFileName)
	viper.SetConfigType("toml")
	viper.SetConfigFile(configPath)

	// Set default values
	viper.SetDefault("version", constants.Version)
	viper.SetDefault("server.host", defaultHost)
	viper.SetDefault("server.port", defaultPort)
	viper.SetDefault("server.offline", false)
	viper.SetDefault("database.name", "seanime")
	viper.SetDefault("web.dir", "$SEANIME_WORKING_DIR/web")
	viper.SetDefault("web.assetDir", "$SEANIME_DATA_DIR/assets")
	viper.SetDefault("cache.dir", "$SEANIME_DATA_DIR/cache")
	viper.SetDefault("manga.backupDir", "$SEANIME_DATA_DIR/cache/manga")
	viper.SetDefault("manga.downloadDir", "$SEANIME_DATA_DIR/manga")
	viper.SetDefault("logs.dir", "$SEANIME_DATA_DIR/logs")
	viper.SetDefault("offline.dir", "$SEANIME_DATA_DIR/offline")
	viper.SetDefault("offline.assetDir", "$SEANIME_DATA_DIR/offline/assets")

	// Create and populate the config file if it doesn't exist
	if err = createConfigFile(configPath); err != nil {
		return nil, err
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal the config values
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Update the config if the version has changed
	if err := updateVersion(cfg, options); err != nil {
		return nil, err
	}

	// Expand the values, replacing environment variables
	expandEnvironmentValues(cfg)
	cfg.Data.AppDataDir = dataDir
	cfg.Data.WorkingDir = os.Getenv("SEANIME_WORKING_DIR")

	// Check validity of the config
	if err := validateConfig(cfg, logger); err != nil {
		return nil, err
	}

	return cfg, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (cfg *Config) GetServerAddr(df ...string) string {
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

func (cfg *Config) GetServerURI(df ...string) string {
	pAddr := fmt.Sprintf("http://%s", cfg.GetServerAddr(df...))
	if cfg.Server.Host == "" || cfg.Server.Host == "0.0.0.0" {
		pAddr = fmt.Sprintf(":%d", cfg.Server.Port)
		if len(df) > 0 {
			pAddr = fmt.Sprintf("http://%s:%d", df[0], cfg.Server.Port)
		}
	}
	return pAddr
}

func setDefaultEnvironmentVariables(dataDir string, trueWd bool) error {
	if os.Getenv("SEANIME_DATA_DIR") == "" {
		if err := os.Setenv("SEANIME_DATA_DIR", dataDir); err != nil {
			return err
		}
	}

	var useGetwd bool
	if trueWd {
		if os.Getenv("SEANIME_WORKING_DIR") == "" {
			wd, err := os.Executable()
			if err != nil {
				useGetwd = true
			}
			wd, err = filepath.EvalSymlinks(wd)
			if err != nil {
				useGetwd = true
			}
			wd = filepath.Dir(wd)
			if err = os.Setenv("SEANIME_WORKING_DIR", filepath.FromSlash(wd)); err != nil {
				return err
			}
			useGetwd = false
		}
	} else {
		useGetwd = true
	}

	if useGetwd {
		wd, _ := os.Getwd()
		if err := os.Setenv("SEANIME_WORKING_DIR", filepath.FromSlash(wd)); err != nil {
			return err
		}
	}
	return nil
}

// validateConfig checks if the config values are valid
func validateConfig(cfg *Config, logger *zerolog.Logger) error {
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
	if cfg.Manga.BackupDir == "" {
		return errInvalidConfigValue("manga.backupDir", "cannot be empty")
	}
	if cfg.Manga.DownloadDir == "" {
		return errInvalidConfigValue("manga.downloadDir", "cannot be empty")
	}

	// Uncomment if "mediastream" is no longer an experimental feature
	//if cfg.Experimental.Mediastream != nil {
	//	logger.Warn().Msgf("app: 'Media streaming' feature is no longer experimental, please remove the flag from your config file")
	//}

	return nil
}

// errInvalidConfigValue returns an error for an invalid config value
func errInvalidConfigValue(s string, s2 string) error {
	return errors.New(fmt.Sprintf("invalid config value: \"%s\" %s", s, s2))
}

func updateVersion(cfg *Config, opts *ConfigOptions) error {
	defer func() {
		if r := recover(); r != nil {
			// Do nothing
		}
	}()

	if cfg.Version != constants.Version {
		for _, f := range opts.OnVersionChange {
			f(cfg.Version, constants.Version)
		}
		cfg.Version = constants.Version
	}

	viper.Set("version", constants.Version)

	return viper.WriteConfig()
}

func expandEnvironmentValues(cfg *Config) {
	defer func() {
		if r := recover(); r != nil {
			// Do nothing
		}
	}()
	cfg.Web.AssetDir = filepath.FromSlash(os.ExpandEnv(cfg.Web.AssetDir))
	cfg.Web.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Web.Dir))
	cfg.Cache.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Cache.Dir))
	cfg.Logs.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Logs.Dir))
	cfg.Manga.BackupDir = filepath.FromSlash(os.ExpandEnv(cfg.Manga.BackupDir))
	cfg.Manga.DownloadDir = filepath.FromSlash(os.ExpandEnv(cfg.Manga.DownloadDir))
	cfg.Offline.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Offline.Dir))
	cfg.Offline.AssetDir = filepath.FromSlash(os.ExpandEnv(cfg.Offline.AssetDir))
}

// createConfigFile creates a default config file if it doesn't exist
func createConfigFile(configPath string) error {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
			return err
		}
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}
	return nil
}

func initAppDataDir(definedDataDir string, logger *zerolog.Logger) (dataDir string, configPath string, err error) {

	// User defined data directory
	if definedDataDir != "" {
		// Normalize the data directory path
		dataDir = filepath.FromSlash(os.ExpandEnv(definedDataDir))

		logger.Trace().Str("dataDir", dataDir).Msg("app: Overriding default data directory")
	} else {
		// Default OS data directory
		// windows: %APPDATA%
		// unix: $XDG_CONFIG_HOME or $HOME
		// darwin: $HOME/Library/Application Support
		dataDir, err = os.UserConfigDir()
		if err != nil {
			return "", "", err
		}
		// Get the app directory
		dataDir = filepath.Join(dataDir, "Seanime")
	}

	// Create data dir if it doesn't exist
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", "", err
	}

	// Get the config file path
	// Normalize the config file path
	configPath = filepath.FromSlash(filepath.Join(dataDir, constants.ConfigFileName))
	// Normalize the data directory path
	dataDir = filepath.FromSlash(dataDir)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

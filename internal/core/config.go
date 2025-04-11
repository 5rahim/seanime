package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/constants"
	"seanime/internal/util"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Config struct {
	Version string
	Server  struct {
		Host          string
		Port          int
		Offline       bool
		UseBinaryPath bool // Makes $SEANIME_WORKING_DIR point to the binary's directory
		Systray       bool
		DoHUrl        string
	}
	Database struct {
		Name string
	}
	Web struct {
		AssetDir string
	}
	Logs struct {
		Dir string
	}
	Cache struct {
		Dir          string
		TranscodeDir string
	}
	Offline struct {
		Dir      string
		AssetDir string
	}
	Manga struct {
		DownloadDir string
	}
	Data struct { // Hydrated after config is loaded
		AppDataDir string
		WorkingDir string
	}
	Extensions struct {
		Dir string
	}
	Anilist struct {
		ClientID string
	}
	Experimental struct {
		MainServerTorrentStreaming bool
	}
}

type ConfigOptions struct {
	DataDir          string // The path to the Seanime data directory, if any
	OnVersionChange  []func(oldVersion string, newVersion string)
	EmbeddedLogo     []byte // The embedded logo
	IsDesktopSidecar bool   // Run as the desktop sidecar
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
	if err = setDataDirEnv(dataDir); err != nil {
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
	// Use the binary's directory as the working directory environment variable on macOS
	viper.SetDefault("server.useBinaryPath", true)
	//viper.SetDefault("server.systray", true)
	viper.SetDefault("database.name", "seanime")
	viper.SetDefault("web.assetDir", "$SEANIME_DATA_DIR/assets")
	viper.SetDefault("cache.dir", "$SEANIME_DATA_DIR/cache")
	viper.SetDefault("cache.transcodeDir", "$SEANIME_DATA_DIR/cache/transcode")
	viper.SetDefault("manga.downloadDir", "$SEANIME_DATA_DIR/manga")
	viper.SetDefault("logs.dir", "$SEANIME_DATA_DIR/logs")
	viper.SetDefault("offline.dir", "$SEANIME_DATA_DIR/offline")
	viper.SetDefault("offline.assetDir", "$SEANIME_DATA_DIR/offline/assets")
	viper.SetDefault("extensions.dir", "$SEANIME_DATA_DIR/extensions")

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

	// Before expanding the values, check if we need to override the working directory
	if err = setWorkingDirEnv(cfg.Server.UseBinaryPath); err != nil {
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

	go loadLogo(options.EmbeddedLogo, dataDir)

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

func getWorkingDir(useBinaryPath bool) (string, error) {
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	binaryDir := ""
	if exe, err := os.Executable(); err == nil {
		if p, err := filepath.EvalSymlinks(exe); err == nil {
			binaryDir = filepath.Dir(p)
			binaryDir = filepath.FromSlash(binaryDir)
		}
	}

	if useBinaryPath && binaryDir != "" {
		return binaryDir, nil
	}

	//// Use the binary's directory as the working directory if needed
	//if useBinaryPath {
	//	exe, err := os.Executable()
	//	if err != nil {
	//		return wd, nil // Fallback to working dir
	//	}
	//	p, err := filepath.EvalSymlinks(exe)
	//	if err != nil {
	//		return wd, nil // Fallback to working dir
	//	}
	//	wd = filepath.Dir(p) // Set the binary's directory as the working directory
	//	return wd, nil
	//}
	return wd, nil
}

func setDataDirEnv(dataDir string) error {
	// Set the data directory environment variable
	if os.Getenv("SEANIME_DATA_DIR") == "" {
		if err := os.Setenv("SEANIME_DATA_DIR", dataDir); err != nil {
			return err
		}
	}

	return nil
}

func setWorkingDirEnv(useBinaryPath bool) error {
	// Set the working directory environment variable
	wd, err := getWorkingDir(useBinaryPath)
	if err != nil {
		return err
	}
	if err = os.Setenv("SEANIME_WORKING_DIR", filepath.FromSlash(wd)); err != nil {
		return err
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
	if cfg.Web.AssetDir == "" {
		return errInvalidConfigValue("web.assetDir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Web.AssetDir); err != nil {
		return wrapInvalidConfigValue("web.assetDir", err)
	}

	if cfg.Cache.Dir == "" {
		return errInvalidConfigValue("cache.dir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Cache.Dir); err != nil {
		return wrapInvalidConfigValue("cache.dir", err)
	}

	if cfg.Cache.TranscodeDir == "" {
		return errInvalidConfigValue("cache.transcodeDir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Cache.TranscodeDir); err != nil {
		return wrapInvalidConfigValue("cache.transcodeDir", err)
	}

	if cfg.Logs.Dir == "" {
		return errInvalidConfigValue("logs.dir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Logs.Dir); err != nil {
		return wrapInvalidConfigValue("logs.dir", err)
	}

	if cfg.Manga.DownloadDir == "" {
		return errInvalidConfigValue("manga.downloadDir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Manga.DownloadDir); err != nil {
		return wrapInvalidConfigValue("manga.downloadDir", err)
	}

	if cfg.Extensions.Dir == "" {
		return errInvalidConfigValue("extensions.dir", "cannot be empty")
	}
	if err := checkIsValidPath(cfg.Extensions.Dir); err != nil {
		return wrapInvalidConfigValue("extensions.dir", err)
	}

	// Uncomment if "MainServerTorrentStreaming" is no longer an experimental feature
	if cfg.Experimental.MainServerTorrentStreaming {
		logger.Warn().Msgf("app: 'Main Server Torrent Streaming' feature is no longer experimental, remove the flag from your config file")
	}

	return nil
}

func checkIsValidPath(path string) error {
	ok := filepath.IsAbs(path)
	if !ok {
		return errors.New("path is not an absolute path")
	}
	return nil
}

// errInvalidConfigValue returns an error for an invalid config value
func errInvalidConfigValue(s string, s2 string) error {
	return fmt.Errorf("invalid config value: \"%s\" %s", s, s2)
}
func wrapInvalidConfigValue(s string, err error) error {
	return fmt.Errorf("invalid config value: \"%s\" %w", s, err)
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
	cfg.Cache.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Cache.Dir))
	cfg.Cache.TranscodeDir = filepath.FromSlash(os.ExpandEnv(cfg.Cache.TranscodeDir))
	cfg.Logs.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Logs.Dir))
	cfg.Manga.DownloadDir = filepath.FromSlash(os.ExpandEnv(cfg.Manga.DownloadDir))
	cfg.Offline.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Offline.Dir))
	cfg.Offline.AssetDir = filepath.FromSlash(os.ExpandEnv(cfg.Offline.AssetDir))
	cfg.Extensions.Dir = filepath.FromSlash(os.ExpandEnv(cfg.Extensions.Dir))
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

		// Expand environment variables
		definedDataDir = filepath.FromSlash(os.ExpandEnv(definedDataDir))

		if !filepath.IsAbs(definedDataDir) {
			return "", "", errors.New("app: Data directory path must be absolute")
		}

		// Replace the default data directory
		dataDir = definedDataDir

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

func loadLogo(embeddedLogo []byte, dataDir string) (err error) {
	defer util.HandlePanicInModuleWithError("core/loadLogo", &err)

	if len(embeddedLogo) == 0 {
		return nil
	}

	logoPath := filepath.Join(dataDir, "logo.png")
	if _, err = os.Stat(logoPath); os.IsNotExist(err) {
		if err = os.WriteFile(logoPath, embeddedLogo, 0644); err != nil {
			return err
		}
	}
	return nil
}

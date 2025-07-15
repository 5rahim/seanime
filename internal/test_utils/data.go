package test_utils

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
)

var ConfigData = &Config{}

const (
	TwoLevelDeepTestConfigPath   = "../../test"
	TwoLevelDeepDataPath         = "../../test/data"
	TwoLevelDeepTestDataPath     = "../../test/testdata"
	ThreeLevelDeepTestConfigPath = "../../../test"
	ThreeLevelDeepDataPath       = "../../../test/data"
	ThreeLevelDeepTestDataPath   = "../../../test/testdata"
)

var ConfigPath = ThreeLevelDeepTestConfigPath
var TestDataPath = ThreeLevelDeepTestDataPath
var DataPath = ThreeLevelDeepDataPath

type (
	Config struct {
		Provider ProviderConfig `mapstructure:"provider"`
		Path     PathConfig     `mapstructure:"path"`
		Database DatabaseConfig `mapstructure:"database"`
		Flags    FlagsConfig    `mapstructure:"flags"`
	}

	FlagsConfig struct {
		EnableAnilistTests         bool `mapstructure:"enable_anilist_tests"`
		EnableAnilistMutationTests bool `mapstructure:"enable_anilist_mutation_tests"`
		EnableMalTests             bool `mapstructure:"enable_mal_tests"`
		EnableMalMutationTests     bool `mapstructure:"enable_mal_mutation_tests"`
		EnableMediaPlayerTests     bool `mapstructure:"enable_media_player_tests"`
		EnableTorrentClientTests   bool `mapstructure:"enable_torrent_client_tests"`
		EnableTorrentstreamTests   bool `mapstructure:"enable_torrentstream_tests"`
	}

	ProviderConfig struct {
		AnilistJwt           string `mapstructure:"anilist_jwt"`
		AnilistUsername      string `mapstructure:"anilist_username"`
		MalJwt               string `mapstructure:"mal_jwt"`
		QbittorrentHost      string `mapstructure:"qbittorrent_host"`
		QbittorrentPort      int    `mapstructure:"qbittorrent_port"`
		QbittorrentUsername  string `mapstructure:"qbittorrent_username"`
		QbittorrentPassword  string `mapstructure:"qbittorrent_password"`
		QbittorrentPath      string `mapstructure:"qbittorrent_path"`
		TransmissionHost     string `mapstructure:"transmission_host"`
		TransmissionPort     int    `mapstructure:"transmission_port"`
		TransmissionPath     string `mapstructure:"transmission_path"`
		TransmissionUsername string `mapstructure:"transmission_username"`
		TransmissionPassword string `mapstructure:"transmission_password"`
		MpcHost              string `mapstructure:"mpc_host"`
		MpcPort              int    `mapstructure:"mpc_port"`
		MpcPath              string `mapstructure:"mpc_path"`
		VlcHost              string `mapstructure:"vlc_host"`
		VlcPort              int    `mapstructure:"vlc_port"`
		VlcPassword          string `mapstructure:"vlc_password"`
		VlcPath              string `mapstructure:"vlc_path"`
		MpvPath              string `mapstructure:"mpv_path"`
		MpvSocket            string `mapstructure:"mpv_socket"`
		IinaPath             string `mapstructure:"iina_path"`
		IinaSocket           string `mapstructure:"iina_socket"`
		TorBoxApiKey         string `mapstructure:"torbox_api_key"`
		RealDebridApiKey     string `mapstructure:"realdebrid_api_key"`
	}
	PathConfig struct {
		DataDir string `mapstructure:"dataDir"`
	}

	DatabaseConfig struct {
		Name string `mapstructure:"name"`
	}

	FlagFunc func() bool
)

func Anilist() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableAnilistTests
	}
}

func AnilistMutation() FlagFunc {
	return func() bool {
		f := ConfigData.Flags.EnableAnilistMutationTests
		if !f {
			fmt.Println("skipping anilist mutation tests")
			return false
		}
		if ConfigData.Provider.AnilistJwt == "" {
			fmt.Println("skipping anilist mutation tests, no anilist jwt")
			return false
		}
		return true
	}
}
func MyAnimeList() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableMalTests
	}
}
func MyAnimeListMutation() FlagFunc {
	return func() bool {
		f := ConfigData.Flags.EnableMalMutationTests
		if !f {
			fmt.Println("skipping mal mutation tests")
			return false
		}
		if ConfigData.Provider.MalJwt == "" {
			fmt.Println("skipping mal mutation tests, no mal jwt")
			return false
		}
		return true
	}
}
func MediaPlayer() FlagFunc {
	return func() bool {
		f := ConfigData.Flags.EnableMediaPlayerTests
		if !f {
			fmt.Println("skipping media player tests")
			return false
		}
		if ConfigData.Provider.MpvPath == "" {
			fmt.Println("skipping media player tests, no mpv path")
			return false
		}
		return true
	}
}
func TorrentClient() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableTorrentClientTests
	}
}
func Torrentstream() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableTorrentstreamTests
	}
}

// InitTestProvider populates the ConfigData and skips the test if the given flags are not set
func InitTestProvider(t *testing.T, args ...FlagFunc) {
	if os.Getenv("TEST_CONFIG_PATH") == "" {
		err := os.Setenv("TEST_CONFIG_PATH", ConfigPath)
		if err != nil {
			log.Fatalf("couldn't set TEST_CONFIG_PATH: %s", err)
		}
	}
	ConfigData = getConfig()

	for _, fn := range args {
		if !fn() {
			t.Skip()
			break
		}
	}
}

func SetTestConfigPath(path string) {
	err := os.Setenv("TEST_CONFIG_PATH", path)
	if err != nil {
		log.Fatalf("couldn't set TEST_CONFIG_PATH: %s", err)
	}
}
func SetTwoLevelDeep() {
	ConfigPath = TwoLevelDeepTestConfigPath
	TestDataPath = TwoLevelDeepTestDataPath
	DataPath = TwoLevelDeepDataPath
}

func getConfig() *Config {
	configPath, exists := os.LookupEnv("TEST_CONFIG_PATH")
	if !exists {
		log.Fatalf("TEST_CONFIG_PATH not set")
	}

	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("couldn't load config: %s", err)
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		fmt.Printf("couldn't read config: %s", err)
	}
	return &c
}

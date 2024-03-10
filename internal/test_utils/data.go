package test_utils

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"testing"
)

var ConfigData = &Config{}
var TwoLevelDeepTestDataPath = "../../test/testdata"

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
	}

	ProviderConfig struct {
		AnilistJwt      string `mapstructure:"anilist_jwt"`
		AnilistUsername string `mapstructure:"anilist_username"`
		MalJwt          string `mapstructure:"mal_jwt"`
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
		return ConfigData.Flags.EnableAnilistMutationTests
	}
}

func MyAnimeList() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableMalTests
	}
}

func MyAnimeListMutation() FlagFunc {
	return func() bool {
		return ConfigData.Flags.EnableMalMutationTests
	}
}

// InitTestProvider populates the ConfigData and skips the test if the given flags are not set
func InitTestProvider(t *testing.T, args ...FlagFunc) {
	err := os.Setenv("TEST_CONFIG_PATH", "../../test")
	ConfigData = getConfig()
	if err != nil {
		log.Fatalf("couldn't set TEST_CONFIG_PATH: %s", err)
	}
	for _, fn := range args {
		if !fn() {
			t.Skip()
			break
		}
	}
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

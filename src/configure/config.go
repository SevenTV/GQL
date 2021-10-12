package configure

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Level      string `mapstructure:"level" json:"level"`
	ConfigFile string `mapstructure:"config_file" json:"config_file"`

	WebsiteURL string `mapstructure:"website_url" json:"website_url"`

	NodeName string `mapstructure:"node_name" json:"node_name"`

	GoogleCredentials string `mapstructure:"google_credentials" json:"google_credentials"`

	CdnURL string `mapstructure:"cdn_url" json:"cdn_url"`

	Redis struct {
		URI string `mapstructure:"uri" json:"uri"`
	} `mapstructure:"redis" json:"redis"`

	Mongo struct {
		URI string `mapstructure:"uri" json:"uri"`
		DB  string `mapstructure:"db" json:"db"`
	} `mapstructure:"http" json:"mongo"`

	Http struct {
		URI              string `mapstructure:"uri" json:"uri"`
		Type             string `mapstructure:"type" json:"type"`
		OauthRedirectURI string `mapstructure:"oauth_redirect_uri" json:"oauth_redirect_uri"`
	} `mapstructure:"http" json:"http"`

	Auth struct {
		Secret string `mapstructure:"secret" json:"secret"`

		Platforms []struct {
			Name    string `mapstructure:"name" json:"name"`
			Enabled bool   `mapstructure:"enabled" json:"enabled"`
		} `mapstructure:"platforms" json:"platforms"`
	} `mapstructure:"auth" json:"auth"`
}

func checkErr(err error) {
	if err != nil {
		logrus.WithError(err).Fatal("config")
	}
}

func New() *Config {
	config := viper.New()

	// Default config
	b, _ := json.Marshal(Config{
		ConfigFile: "config.yaml",
	})
	tmp := viper.New()
	defaultConfig := bytes.NewReader(b)
	tmp.SetConfigType("json")
	checkErr(tmp.ReadConfig(defaultConfig))
	checkErr(config.MergeConfigMap(viper.AllSettings()))

	// File
	config.SetConfigFile(config.GetString("config_file"))
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		logrus.Warning(err)
		logrus.Info("Using default config")
	} else {
		checkErr(config.MergeInConfig())
	}

	// Environment
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.SetEnvPrefix("7TV")
	config.AllowEmptyEnv(true)
	config.AutomaticEnv()

	// Print final config
	c := &Config{}
	checkErr(config.Unmarshal(&c))

	InitLogging(c.Level)

	return c
}
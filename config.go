package main

import "github.com/spf13/viper"

type config struct {
	Sessdata      string `mapstructure:"SESSDATA"`
	GmailPassword string `mapstructure:"GMAIL_PASSWORD"`
}

func loadConfig(path string) (config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return config{}, err
	}

	var cfg config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

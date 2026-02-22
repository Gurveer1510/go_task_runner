package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl  string
	Port   string
	AppEnv string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("PORT", "8080")

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			return nil, fileLookupError
		} else {
			return nil, err
		}
	}

	config := &Config{
		AppEnv: viper.GetString("APP_ENV"),
		DBUrl:  viper.GetString("DATABASE_URL"),
		Port:   viper.GetString("PORT"),
	}

	return config, nil
}

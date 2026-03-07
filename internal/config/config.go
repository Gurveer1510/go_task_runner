package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl       string
	Port        string
	AppEnv      string
	Concurrency int
	BaseDelay   time.Duration
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("WORKER_CONCURRENCY", 5)
	viper.SetDefault("WORKER_BASE_DELAY_MS", 2000)

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			return nil, fileLookupError
		} else {
			return nil, err
		}
	}

	config := &Config{
		AppEnv:      viper.GetString("APP_ENV"),
		DBUrl:       viper.GetString("DATABASE_URL"),
		Port:        viper.GetString("PORT"),
		Concurrency: viper.GetInt("WORKER_CONCURRENCY"),
		BaseDelay:   time.Duration(viper.GetInt("WORKER_BASE_DELAY_MS")) * time.Millisecond,
	}

	return config, nil
}

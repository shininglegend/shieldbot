// internal/config/config.go
package config

import "github.com/spf13/viper"

type Config struct {
	Token       string
	DatabaseURL string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		Token:       viper.GetString("token"),
		DatabaseURL: viper.GetString("database_url"),
	}, nil
}

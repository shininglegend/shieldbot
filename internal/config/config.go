// internal/config/config.go
package config

import "github.com/spf13/viper"

type Config struct {
	Token        string
	DatabasePath string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// In the Load function:
	return &Config{
		Token:        viper.GetString("token"),
		DatabasePath: viper.GetString("database_path"),
	}, nil
}

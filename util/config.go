package util

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/viper"
)

//this config file stores all the configurations of the application
//The values are read by viper from a config file or environment variable
type Config struct {
	DBdriver string `mapstructure:"DB_DRIVER"`
	DBsource string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey string `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config,err  error){
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	
	viper.AutomaticEnv()
	// Ensure env-only values are included when unmarshalling into Config.
	_ = viper.BindEnv("DB_DRIVER")
	_ = viper.BindEnv("DB_SOURCE")
	_ = viper.BindEnv("SERVER_ADDRESS")
	_ = viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	_ = viper.BindEnv("ACCESS_TOKEN_DURATION")
	_ = viper.BindEnv("PORT")

	err = viper.ReadInConfig()
	if err != nil{
		// Allow running without a config file (e.g., in Docker) as long as
		// required values are provided via environment variables.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
		return
		}
		err = nil
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return 
	}

	// Render (and similar platforms) usually provide PORT. If SERVER_ADDRESS isn't
	// set, default to binding on 0.0.0.0:$PORT.
	if config.ServerAddress == "" {
		if port := os.Getenv("PORT"); port != "" {
			config.ServerAddress = "0.0.0.0:" + port
		}
	}
	return 

}
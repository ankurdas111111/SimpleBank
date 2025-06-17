package util

import (
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

	err = viper.ReadInConfig()
	if err != nil{
		return
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return 
	}
	return 

}
package config

import (
	"github.com/dennishilgert/cloud-computing-2/pkg/logger"
	"github.com/spf13/viper"
)

var log = logger.NewLogger("app.config")

type Config struct {
	AppPort      int
	GpcProjectId string
	RedisHost    string
	RedisPort    int
	Logger       logger.Options
}

func Load() (*Config, error) {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()

	// loading the values from the environment or use default values
	loadOrDefault("Logger.AppId", "LOG_APP_ID", logger.DefaultOptions().AppId)
	loadOrDefault("Logger.JSONFormatEnabled", "LOG_FORMAT_JSON", logger.DefaultOptions().JSONFormatEnabled)
	loadOrDefault("Logger.OutputLevel", "LOG_LEVEL", logger.DefaultOptions().OutputLevel)

	loadOrDefault("AppPort", "APP_PORT", 80)
	loadOrDefault("GpcProjectId", "GOOGLE_CLOUD_PROJECT_ID", nil)
	loadOrDefault("RedisHost", "REDIS_HOST", nil)
	loadOrDefault("RedisPort", "REDIS_PORT", 6379)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to unmarshal config: %v", err)
		return nil, err
	}

	return &config, nil
}

func loadOrDefault(configVar string, envVar string, defaultVal any) {
	if defaultVal != nil {
		viper.SetDefault(configVar, defaultVal)
	}
	viper.BindEnv(configVar, envVar)
	if defaultVal == nil {
		if !viper.IsSet(configVar) {
			log.Fatalf("required environment variable %s is not set", envVar)
		}
	}
}

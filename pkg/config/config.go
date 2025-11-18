package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type MongoDB struct {
	Host     string `mapstructure:"DB_HOST" json:"dbHost" validate:"required,hostname|ip"`
	Port     int    `mapstructure:"DB_PORT" json:"dbPort" validate:"required,min=1000,max=65535"`
	Username string `mapstructure:"DB_USERNAME" json:"dbUsername" validate:"required"`
	Password string `mapstructure:"DB_PASSWORD" json:"dbPassword" validate:"required"`
}

type Config struct {
	Port    int      `mapstructure:"PORT" json:"port" validate:"required,min=1000,max=65535"`
	GinMode string   `mapstructure:"GIN_MODE" validate:"required,oneof=debug test release"`
	DB      MongoDB  `mapstructure:",squash" json:"db"`
	Origins []string `mapstructure:"ORIGINS" json:"origins"`
}

var Cfg *Config
var once sync.Once

func Load() *Config {
	once.Do(func() {
		viper.SetConfigType("env")  // treat as key=value pairs
		viper.SetConfigFile(".env") // or viper.AutomaticEnv() for direct env vars
		viper.AutomaticEnv()        // override with actual ENV variables

		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln(fmt.Errorf("error reading config file: %w", err))
		}
		var cfg Config
		if err := viper.Unmarshal(&cfg); err != nil {
			log.Fatalln(fmt.Errorf("error decode config file: %w", err))
		}

		// Validate config using go-playground/validator
		validate := validator.New()
		if err := validate.Struct(cfg); err != nil {
			log.Fatalln(fmt.Errorf("invalid config: %w", err))
		}
		Cfg = &cfg
	})
	return Cfg
}

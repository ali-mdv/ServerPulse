package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type MongoDB struct {
	Host     string `mapstructure:"DB_HOST" json:"dbHost" validate:"required,hostname|ip"`
	Port     int    `mapstructure:"DB_PORT" json:"dbPort" validate:"required,min=1000,max=65535"`
	Username string `mapstructure:"DB_USERNAME" json:"dbUsername" validate:"required"`
	Password string `mapstructure:"DB_PASSWORD" json:"dbPassword" validate:"required"`
}

type Config struct {
	Port          int      `mapstructure:"PORT" json:"port" validate:"required,min=1000,max=65535"`
	GinMode       string   `mapstructure:"GIN_MODE" validate:"required,oneof=debug test release"`
	DB            MongoDB  `mapstructure:",squash" json:"db"`
	Origins       []string `mapstructure:"ORIGINS" json:"origins"`
	PM2SocketPath string   `mapstructure:"PM2_SOCKET_PATH" json:"pm2SocketPath"`
}

var Cfg *Config
var once sync.Once

func Load() *Config {
	once.Do(func() {
		viper.SetConfigType("env") // treat as key=value pairs
		viper.AutomaticEnv()       // override with actual ENV variables

		// Read from .env if exists
		if _, err := os.Stat(".env"); err == nil {
			// .env exists (local dev) — load it
			viper.SetConfigFile(".env")
			if err := viper.ReadInConfig(); err != nil {
				log.Fatalln(fmt.Errorf("error reading config file: %w", err))
			}
		} else {
			log.Println("no .env file found, relying on real environment variables")
		}

		bindEnvs(Config{})

		var cfg Config

		decodeHook := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			trimStringHook,                          // strip stray CR/whitespace/quotes BEFORE type conversion
			mapstructure.StringToSliceHookFunc(","), // ORIGINS=a,b,c -> []string{"a","b","c"}
			mapstructure.StringToTimeDurationHookFunc(),
		))

		// This is the key missing piece: explicitly bind every field
		// so Unmarshal can actually see the env vars.
		// bindEnvs(cfg)

		if err := viper.Unmarshal(&cfg, decodeHook); err != nil {
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

// trimStringHook strips leading/trailing whitespace, stray CR characters
// (common when a .env file has Windows/CRLF line endings, or when Docker's
// env_file: parsing doesn't strip quotes) and surrounding quote characters
// from every string value BEFORE mapstructure converts it to its target
// type. This is what causes "27017\r" to fail strconv.ParseInt and
// "mongo\r" to fail the hostname regex even though both look correct
// when printed.
func trimStringHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String {
		return data, nil
	}
	s, ok := data.(string)
	if !ok {
		return data, nil
	}
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return s, nil
}

// bindEnvs walks a struct (recursing into squashed/nested structs) and
// registers each mapstructure tag with Viper via BindEnv, so that
// AutomaticEnv + Unmarshal will find the value even when no config
// file was loaded at all.
func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for i := 0; i < ift.NumField(); i++ {
		fv := ifv.Field(i)
		field := ift.Field(i)

		tag := field.Tag.Get("mapstructure")

		// Squashed embedded struct: recurse without adding to the key path.
		if strings.Contains(tag, "squash") {
			bindEnvs(fv.Interface(), parts...)
			continue
		}

		if tag == "" {
			tag = field.Name
		}

		switch fv.Kind() {
		case reflect.Struct:
			bindEnvs(fv.Interface(), append(parts, tag)...)
		default:
			key := strings.Join(append(parts, tag), ".")
			if err := viper.BindEnv(key); err != nil {
				log.Fatalln(fmt.Errorf("error binding env var %s: %w", key, err))
			}
		}
	}
}

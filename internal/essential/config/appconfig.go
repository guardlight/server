package config

import (
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const (
	glEnvPrefix = "GUARDLIGHT_"
)

type server struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type cors struct {
	Origins []string `koanf:"origins"`
}

type database struct {
	Url  string `koanf:"url"`
	Name string `koanf:"name"`
}

type GLConfig struct {
	Env      string   `koanf:"env"`
	Domain   string   `koanf:"domain"`
	Website  string   `koanf:"website"`
	Server   server   `koanf:"server"`
	Cors     cors     `koanf:"cors"`
	Database database `koanf:"database"`
	Console  console  `koanf:"console"`
}

type jwt struct {
	MaxAge     int    `koanf:"maxAge"`
	SigningKey string `koanf:"signingKey"`
}

type console struct {
	Jwt jwt `koanf:"jwt"`
}

var (
	conf *GLConfig
)

func SetupConfig(envFilePath string) {
	k := koanf.New(".")

	// Load environment variables from file
	if err := k.Load(file.Provider(envFilePath), yaml.Parser()); err != nil {
		zap.S().Fatalw("error loading config", "error", err)
	}

	// Load environment variables from environment with ORBIT_ prefix.
	// Will override properties from the file
	k.Load(env.Provider(glEnvPrefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, glEnvPrefix)), "_", ".", -1)
	}), nil)

	ffc := &GLConfig{}
	k.Unmarshal("", ffc)

	os.Setenv("TZ", "UTC")

	if _, ok := os.LookupEnv("SHOW_CONFIG"); ok {
		zap.S().Infow("config", "config", ffc)
	}

	conf = ffc
	// zap.S().Infow("configuration loaded", "config", conf)
	zap.S().Infow("configuration loaded", "file", envFilePath)
}

func Get() GLConfig {
	return *conf
}

func (fc GLConfig) IsProduction() bool {
	return fc.Env == "production"
}

func (fc GLConfig) IsStaging() bool {
	return fc.Env == "staging"
}

func (fc GLConfig) IsDevelopment() bool {
	return fc.Env == "development"
}

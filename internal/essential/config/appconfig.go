package config

import (
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	glEnvPrefix = "GUARDLIGHT_"
)

type GLConfig struct {
	Env          string       `koanf:"env"`
	Domain       string       `koanf:"domain"`
	Server       server       `koanf:"server"`
	Cors         cors         `koanf:"cors"`
	Database     database     `koanf:"database"`
	Orchestrator orchestrator `koanf:"orchestrator"`
	Console      console      `koanf:"console"`
	Parsers      []parser     `koanf:"parsers"`
	Analyzers    []analyzer   `koanf:"analyzers"`
}

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

type orchestrator struct {
	ScheduleRateCron string `koanf:"scheduleRateCron"`
}

type jwt struct {
	MaxAge     int    `koanf:"maxAge"`
	SigningKey string `koanf:"signingKey"`
}

type console struct {
	Jwt jwt `koanf:"jwt"`
}

type parser struct {
	Type        string `koanf:"type"`
	Name        string `koanf:"name"`
	Description string `koanf:"description"`
	Concurrency int    `koanf:"concurrency"`
}

type analyzer struct {
	Key            string          `koanf:"key"`
	Name           string          `koanf:"name"`
	Description    string          `koanf:"description"`
	ContenxtWindow string          `koanf:"contentWindow"`
	Model          string          `koanf:"model"`
	Concurrency    int             `koanf:"concurrency"`
	Inputs         []analyzerInput `koanf:"inputs"`
}

type analyzerInput struct {
	Key         string `koanf:"key"`
	Name        string `koanf:"name"`
	Description string `koanf:"description"`
	Type        string `koanf:"type"`
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

func (fc GLConfig) GetParser(parserType string) (parser, bool) {
	return lo.Find(Get().Parsers, func(a parser) bool {
		return a.Type == parserType
	})
}

func (fc GLConfig) GetAnalyzer(analyzerKey string) (analyzer, bool) {
	return lo.Find(Get().Analyzers, func(a analyzer) bool {
		return a.Key == analyzerKey
	})
}

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/creasty/defaults"
	"github.com/google/uuid"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	glEnvPrefix = "GUARDLIGHT_"
)

type GLConfig struct {
	Timezone     string       `koanf:"tz" default:"UTC"`
	Env          string       `koanf:"env" default:"production"`
	Domain       string       `koanf:"domain" default:"127.0.0.1"`
	Server       server       `koanf:"server"`
	Cors         cors         `koanf:"cors"`
	Database     database     `koanf:"database"`
	Orchestrator orchestrator `koanf:"orchestrator"`
	Console      console      `koanf:"console"`
	Parsers      []parser     `koanf:"parsers"`
	Analyzers    []analyzer   `koanf:"analyzers"`
	Users        []User       `koanf:"users"`
}

type User struct {
	Username string    `koanf:"username" default:"-"`
	Password string    `koanf:"password" default:"-"`
	Role     string    `koanf:"role" default:"-"`
	Id       uuid.UUID `koanf:"id" default:"-"`
}

type server struct {
	Host string `koanf:"host" default:"0.0.0.0"`
	Port int    `koanf:"port" default:"6842"`
}

type cors struct {
	Origin string `koanf:"origin" default:"0.0.0.0"`
}

type database struct {
	Server   string `koanf:"server" default:"127.0.0.1"`
	Port     int    `koanf:"port" default:"5432"`
	Name     string `koanf:"name" default:"guardlight"`
	User     string `koanf:"user" default:"root"`
	Password string `koanf:"password" default:"root"`
}

type orchestrator struct {
	ScheduleRateCron string `koanf:"scheduleRateCron" default:"*/5 * * * * *"`
}

type jwt struct {
	MaxAge     int    `koanf:"maxAge" default:"3600"`
	SigningKey string `koanf:"signingKey" default:"-"`
}

type console struct {
	Jwt jwt `koanf:"jwt"`
}

type parser struct {
	Type        string `koanf:"type" default:"-"`
	Name        string `koanf:"name" default:"-"`
	Description string `koanf:"description" default:"-"`
	Concurrency int    `koanf:"concurrency" default:"-"`
	Image       string `koanf:"image" default:"-"`
}

type analyzer struct {
	Key           string          `koanf:"key" default:"-"`
	Name          string          `koanf:"name" default:"-"`
	Description   string          `koanf:"description" default:"-"`
	ContextWindow int             `koanf:"contextWindow" default:"-"`
	Model         string          `koanf:"model" default:"-"`
	Concurrency   int             `koanf:"concurrency" default:"-"`
	Inputs        []AnalyzerInput `koanf:"inputs" default:"-"`
	Image         string          `koanf:"image" default:"-"`
}

type AnalyzerInput struct {
	Key         string `koanf:"key" default:"-"`
	Name        string `koanf:"name"  default:"-"`
	Description string `koanf:"description" default:"-"`
	Type        string `koanf:"type" default:"-"`
}

var (
	conf *GLConfig
)

func SetupConfig(envFilePath string) {
	k := koanf.New(".")

	defaultedConfig := &GLConfig{}
	if err := defaults.Set(defaultedConfig); err != nil {
		zap.S().Fatalw("error loading config from defaults", "error", err)
	}

	configBasicAdapters(defaultedConfig)

	// Load defaults variables
	if err := k.Load(structs.Provider(defaultedConfig, "koanf"), nil); err != nil {
		zap.S().Fatalw("error loading config from defaults", "error", err)
	}

	// Check if file exist
	if _, err := os.Stat(envFilePath); err == nil {
		// Load environment variables from file
		if err := k.Load(file.Provider(envFilePath), yaml.Parser()); err != nil {
			zap.S().Fatalw("error loading config", "error", err)
		}
	}

	if tz, ok := os.LookupEnv("TZ"); ok {
		k.Set("tz", tz)
	}

	configSetEnvironment(k)
	// Generate signing key
	configSigningKey(k)
	configAdminUser(k)

	data, err := yaml.Parser().Marshal(k.Raw())
	if err != nil {
		zap.S().Fatalw("error marshaling yaml config", "error", err)
		return
	}

	// Write to a new YAML file
	err = os.WriteFile(envFilePath, data, 0644)
	if err != nil {
		zap.S().Fatalw("error writing config to file", "error", err)
		return
	}

	// Load environment variables from environment with ORBIT_ prefix.
	// Will override properties from the file
	k.Load(env.Provider(glEnvPrefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, glEnvPrefix)), "_", ".", -1)
	}), nil)

	ffc := &GLConfig{}
	k.Unmarshal("", ffc)

	if _, ok := os.LookupEnv("SHOW_CONFIG"); ok {
		zap.S().Infow("config", "config", ffc)
	}

	conf = ffc
	// zap.S().Infow("configuration loaded", "config", conf)
	zap.S().Infow("configuration loaded", "file", envFilePath)
}

func configSetEnvironment(k *koanf.Koanf) {
	value, ok := os.LookupEnv("environment")
	if ok {
		k.Set("env", value)
	}
}

func Get() GLConfig {
	return *conf
}

func (fc GLConfig) GetDbDsn() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?timezone=%s", Get().Database.Name, Get().Database.Password, Get().Database.Server, Get().Database.Port, Get().Database.Name, Get().Timezone)
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

func configBasicAdapters(defaultedConfig *GLConfig) {
	defaultedConfig.Analyzers = append(defaultedConfig.Analyzers, analyzer{
		Key:           "word_search",
		Name:          "Word Search",
		Description:   "Uses a basic word list to scan content.",
		ContextWindow: 32000,
		Model:         "text",
		Concurrency:   4,
		Image:         "builtin",
		Inputs: []AnalyzerInput{
			{
				Key:         "strict_words",
				Name:        "Strict Words",
				Description: "Words in this list will flag content.",
				Type:        "textarea",
			},
		},
	})

	defaultedConfig.Parsers = append(defaultedConfig.Parsers, parser{
		Type:        "freetext",
		Name:        "Freetext Parser",
		Description: "Parses a text to an utf-8 formated text.",
		Concurrency: 4,
		Image:       "builtin",
	})
}

func configSigningKey(k *koanf.Koanf) {
	// Will be overriden if provided by Environment variable: GUARDLIGHT_CONSOLE_JWT_SIGNING_KEY
	sMapKey := "console.jwt.signingKey"
	skey := k.Get(sMapKey)
	if skey == nil || skey == "" {
		k.Set(sMapKey, lo.RandomString(32, lo.AlphanumericCharset))
		zap.S().Info("Created Signing Key")
	}
}

// Create user if not yet created or loaded from file.
func configAdminUser(k *koanf.Koanf) {
	// Will be overriden if provided by Environment variable: GUARDLIGHT_USERS_*
	var users []User
	if err := k.Unmarshal("users", &users); err != nil {
		zap.S().Fatalw("Error unmarshaling users", "error", err)
	}
	if len(users) == 0 {
		users = append(users, User{
			Username: "admin@guardlight.org",
			Password: lo.RandomString(16, lo.AllCharset),
			Role:     "admin",
			Id:       uuid.New(),
		})
		k.Set("users", users)
		zap.S().Info("Created Admin User. Admin Password: " + users[0].Password)
	}
}

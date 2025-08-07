package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	Nats         nats         `koanf:"nats"`
	Parsers      []Parser     `koanf:"parsers"`
	Analyzers    []analyzer   `koanf:"analyzers"`
	Reporters    []reporter   `koanf:"reporters"`
	Users        []User       `koanf:"users"`
	Data         data         `koanf:"data"`
}

type data struct {
	ExportProcessedText bool   `koanf:"exportProcessedText" default:"false"`
	ExportPath          string `koanf:"exportPath" default:"/data/books/processed"`
}

type nats struct {
	Server   string `koanf:"server" default:"-"`
	Port     int    `koanf:"port" default:"4222"`
	User     string `koanf:"user" default:"-"`
	Password string `koanf:"password" default:"-"`
}

type User struct {
	Username string    `koanf:"username" default:"-"`
	Password string    `koanf:"password" default:"-"`
	Role     string    `koanf:"role" default:"-"`
	Id       uuid.UUID `koanf:"id" default:"-"`
	ApiKey   string    `koanf:"apiKey" default:"-"`
}

type server struct {
	Host string `koanf:"host" default:"0.0.0.0"`
	Port int    `koanf:"port" default:"6842"`
}

type cors struct {
	Origin string `koanf:"origin" default:"http://0.0.0.0"`
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

type Parser struct {
	Image       string `koanf:"image" default:"-"`
	External    bool   `koanf:"external" default:"-"`
	Key         string `koanf:"key" default:"-"`
	Name        string `koanf:"name" default:"-"`
	Description string `koanf:"description" default:"-"`
	Type        string `koanf:"type" default:"-"`
	Concurrency int    `koanf:"concurrency" default:"-"`
}

type reporter struct {
	Image       string `koanf:"image" default:"-"`
	External    bool   `koanf:"external" default:"-"`
	Key         string `koanf:"key" default:"-"`
	Name        string `koanf:"name" default:"-"`
	Description string `koanf:"description" default:"-"`
	Concurrency int    `koanf:"concurrency" default:"-"`
}

type analyzer struct {
	Image         string          `koanf:"image" default:"-"`
	External      bool            `koanf:"external" default:"-"`
	Key           string          `koanf:"key" default:"-"`
	Name          string          `koanf:"name" default:"-"`
	Description   string          `koanf:"description" default:"-"`
	ContextWindow int             `koanf:"contextWindow" default:"-"`
	Model         string          `koanf:"model" default:"-"`
	Concurrency   int             `koanf:"concurrency" default:"-"`
	Inputs        []AnalyzerInput `koanf:"inputs" default:"-"`
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
	configNatsCredentials(k)

	data, err := yaml.Parser().Marshal(k.Raw())
	if err != nil {
		zap.S().Fatalw("error marshaling yaml config", "error", err)
		return
	}

	// Write to a new YAML file
	if err := os.MkdirAll(filepath.Dir(envFilePath), os.ModePerm); err != nil {
		zap.S().Fatalw("error creating directories for config file", "error", err)
		return
	}

	if err := os.WriteFile(envFilePath, data, 0644); err != nil {
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

	if err := validateAnalyzers(ffc); err != nil {
		zap.S().Fatalw("Invalid analyzer config", "error", err)
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
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?timezone=%s", Get().Database.User, Get().Database.Password, Get().Database.Server, Get().Database.Port, Get().Database.Name, Get().Timezone)
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

func (fc GLConfig) GetParser(parserType string) (Parser, bool) {
	return lo.Find(Get().Parsers, func(a Parser) bool {
		return a.Type == parserType
	})
}

func (fc GLConfig) GetAnalyzer(analyzerKey string) (analyzer, bool) {
	return lo.Find(Get().Analyzers, func(a analyzer) bool {
		return a.Key == analyzerKey
	})
}

func (fc GLConfig) GetReporter(reporterKey string) (reporter, bool) {
	return lo.Find(Get().Reporters, func(a reporter) bool {
		return a.Key == reporterKey
	})
}

func configBasicAdapters(defaultedConfig *GLConfig) {
	defaultedConfig.Analyzers = append(defaultedConfig.Analyzers, analyzer{
		Key:           "word_search",
		Name:          "Word Search",
		Description:   "Uses a basic word list to scan content.",
		Image:         "builtin",
		External:      true,
		ContextWindow: 32000,
		Model:         "text",
		Concurrency:   4,
		Inputs: []AnalyzerInput{
			{
				Key:         "threshold",
				Name:        "Threshold",
				Description: "Allows you to specificy at which point the analyzer should flag the media content.",
				Type:        "threshold",
			},
			{
				Key:         "strict_words",
				Name:        "Strict Words",
				Description: "Words in this list will be used to flag media content.",
				Type:        "textarea",
			},
		},
	})

	defaultedConfig.Parsers = append(defaultedConfig.Parsers, Parser{
		Key:         "freetext",
		Name:        "Freetext",
		Description: "Parses a text to an utf-8 formated text.",
		Image:       "builtin",
		External:    true,
		Type:        "freetext",
		Concurrency: 4,
	})

	defaultedConfig.Reporters = append(defaultedConfig.Reporters, reporter{
		Key:         "word_count",
		Name:        "Word Count",
		Description: "This reporter will match the threshold to the amount of lines.",
		Image:       "builtin",
		External:    true,
		Concurrency: 4,
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
			Password: lo.RandomString(16, lo.AlphanumericCharset),
			Role:     "admin",
			Id:       uuid.New(),
			ApiKey:   lo.RandomString(32, lo.AlphanumericCharset),
		})
		k.Set("users", users)
		zap.S().Infow("Created Admin User.", "password", users[0].Password, "apiKey", users[0].ApiKey)
	}
}

func validateAnalyzers(gc *GLConfig) error {
	for _, _ = range gc.Analyzers {
		// Validate analyzer input
	}
	return nil
}

func configNatsCredentials(k *koanf.Koanf) {
	nsKey := k.Get("nats.server")

	// Use internal NATS if server is not specified
	if nsKey == nil || nsKey == "" {

		// Set the internal NATS user to use
		nuKey := k.Get("nats.user")
		if nuKey == nil || nuKey == "" {
			k.Set("nats.user", "gl_nats_user")
		}

		npKey := k.Get("nats.password")
		if npKey == nil || npKey == "" {
			k.Set("nats.password", lo.RandomString(16, lo.AlphanumericCharset))
		}

	}

}

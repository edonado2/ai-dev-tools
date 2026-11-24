package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Option configures the loader
type Option func(*loaderOptions)

type loaderOptions struct {
	configDir string
	configYAML string
}

// WithConfigDir sets the config directory to search for config file
func WithConfigDir(dir string) Option {
	return func(o *loaderOptions) {
		o.configDir = dir
	}
}

// WithConfigYAML provides config directly as YAML string (for testing)
func WithConfigYAML(yaml string) Option {
	return func(o *loaderOptions) {
		o.configYAML = yaml
	}
}

// Load loads configuration with priority: CLI flags > env vars > config file > defaults
func Load(opts ...Option) (Config, error) {
	options := loaderOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	// Apply defaults first
	setDefaults(v)

	// Bind env vars explicitly for all config keys
	bindEnvVars(v)

	// If configYAML is provided, read from it directly
	if options.configYAML != "" {
		if err := v.ReadConfig(bytes.NewReader([]byte(options.configYAML))); err != nil {
			return Config{}, fmt.Errorf("read config yaml: %w", err)
		}
	} else {
		// Add config paths in priority order (last added wins for same config name)
		// 1. Global config: config.yaml in ~/.config/ai-dev-tools (lowest priority)
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			v.AddConfigPath(filepath.Join(homeDir, ".config", "ai-dev-tools"))
			v.SetConfigName("config")
			if err := v.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return Config{}, fmt.Errorf("read global config file: %w", err)
				}
			}
		}

		// 2. Repo config: .ai-dev-tools.yaml in current dir (higher priority)
		v.AddConfigPath(".")
		v.SetConfigName(".ai-dev-tools")
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return Config{}, fmt.Errorf("read repo config file: %w", err)
			}
		}

		// 3. Custom config dir (for testing) - highest priority for file config
		if options.configDir != "" {
			v.AddConfigPath(options.configDir)
			v.SetConfigName(".ai-dev-tools")
			if err := v.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return Config{}, fmt.Errorf("read custom config file: %w", err)
				}
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Resolve returns a config with defaults applied (no file/env loading)
func Resolve() Config {
	v := viper.New()
	setDefaults(v)
	var cfg Config
	_ = v.Unmarshal(&cfg)
	return cfg
}

func setDefaults(v *viper.Viper) {
	// Provider defaults
	v.SetDefault("provider.name", "anthropic")
	v.SetDefault("provider.model", "")
	v.SetDefault("provider.api_key_env", "")
	v.SetDefault("provider.base_url", "")
	v.SetDefault("provider.timeout", 60)
	v.SetDefault("provider.max_retries", 3)

	// Tools defaults
	v.SetDefault("tools.enabled", []string{})
	v.SetDefault("tools.timeout", 120)
	v.SetDefault("tools.commit.style", "conventional")
	v.SetDefault("tools.review.severity_threshold", "warning")
	v.SetDefault("tools.review.exclude_patterns", []string{})
	v.SetDefault("tools.review.include_patterns", []string{})

	// GitHub defaults
	v.SetDefault("github.token_env", "GITHUB_TOKEN")
	v.SetDefault("github.owner", "")
	v.SetDefault("github.repo", "")

	// Output defaults
	v.SetDefault("output.format", "text")
	v.SetDefault("output.color", true)
}

func bindEnvVars(v *viper.Viper) {
	// Provider
	v.BindEnv("provider.name", "AI_DEV_PROVIDER_NAME")
	v.BindEnv("provider.model", "AI_DEV_PROVIDER_MODEL")
	v.BindEnv("provider.api_key_env", "AI_DEV_PROVIDER_API_KEY_ENV")
	v.BindEnv("provider.base_url", "AI_DEV_PROVIDER_BASE_URL")
	v.BindEnv("provider.timeout", "AI_DEV_PROVIDER_TIMEOUT")
	v.BindEnv("provider.max_retries", "AI_DEV_PROVIDER_MAX_RETRIES")

	// Tools
	v.BindEnv("tools.enabled", "AI_DEV_TOOLS_ENABLED")
	v.BindEnv("tools.timeout", "AI_DEV_TOOLS_TIMEOUT")

	// GitHub
	v.BindEnv("github.token_env", "AI_DEV_GITHUB_TOKEN_ENV")
	v.BindEnv("github.owner", "AI_DEV_GITHUB_OWNER")
	v.BindEnv("github.repo", "AI_DEV_GITHUB_REPO")

	// Output
	v.BindEnv("output.format", "AI_DEV_OUTPUT_FORMAT")
	v.BindEnv("output.color", "AI_DEV_OUTPUT_COLOR")
}
package config

// Config holds all configuration for ai-dev-tools
type Config struct {
	Provider ProviderConfig `mapstructure:"provider"`
	Tools    ToolsConfig    `mapstructure:"tools"`
	GitHub   GitHubConfig   `mapstructure:"github"`
	Output   OutputConfig   `mapstructure:"output"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	Name       string `mapstructure:"name" validate:"oneof=openai anthropic ollama custom"`
	Model      string `mapstructure:"model"`
	APIKeyEnv  string `mapstructure:"api_key_env"`
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout" validate:"min=1,max=300"`
	MaxRetries int    `mapstructure:"max_retries" validate:"min=0,max=10"`
}

// ToolsConfig holds tools configuration
type ToolsConfig struct {
	Enabled  []string      `mapstructure:"enabled"`
	Timeout  int           `mapstructure:"timeout" validate:"min=1,max=600"`
	Commit   CommitConfig  `mapstructure:"commit"`
	Review   ReviewConfig  `mapstructure:"review"`
	TestGen  TestGenConfig `mapstructure:"testgen"`
}

// ReviewConfig holds code review configuration
type ReviewConfig struct {
	SeverityThreshold string   `mapstructure:"severity_threshold" validate:"oneof=error warning info"`
	ExcludePatterns   []string `mapstructure:"exclude_patterns"`
	IncludePatterns   []string `mapstructure:"include_patterns"`
}

// CommitConfig holds commit command configuration
type CommitConfig struct {
	Style string `mapstructure:"style"`
}

// TestGenConfig holds test generation configuration
type TestGenConfig struct {
	Framework string `mapstructure:"framework"`
	Output    string `mapstructure:"output"`
	Coverage  bool   `mapstructure:"coverage"`
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	TokenEnv string `mapstructure:"token_env"`
	Owner    string `mapstructure:"owner"`
	Repo     string `mapstructure:"repo"`
}

// OutputConfig holds output configuration
type OutputConfig struct {
	Format string `mapstructure:"format" validate:"oneof=json yaml text"`
	Color  bool   `mapstructure:"color"`
}
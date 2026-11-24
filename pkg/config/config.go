package config

// Config holds all configuration for ai-dev-tools
type Config struct {
	Provider ProviderConfig `mapstructure:"provider" yaml:"provider"`
	Tools    ToolsConfig    `mapstructure:"tools" yaml:"tools"`
	GitHub   GitHubConfig   `mapstructure:"github" yaml:"github"`
	Output   OutputConfig   `mapstructure:"output" yaml:"output"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	Name       string `mapstructure:"name" yaml:"name" validate:"oneof=openai anthropic ollama custom"`
	Model      string `mapstructure:"model" yaml:"model"`
	APIKeyEnv  string `mapstructure:"api_key_env" yaml:"api_key_env"`
	BaseURL    string `mapstructure:"base_url" yaml:"base_url"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout" validate:"min=1,max=300"`
	MaxRetries int    `mapstructure:"max_retries" yaml:"max_retries" validate:"min=0,max=10"`
}

// ToolsConfig holds tools configuration
type ToolsConfig struct {
	Enabled   []string        `mapstructure:"enabled" yaml:"enabled"`
	Timeout   int             `mapstructure:"timeout" yaml:"timeout" validate:"min=1,max=600"`
	Commit    CommitConfig    `mapstructure:"commit" yaml:"commit"`
	Review    ReviewConfig    `mapstructure:"review" yaml:"review"`
	TestGen   TestGenConfig   `mapstructure:"testgen" yaml:"testgen"`
	Changelog ChangelogConfig `mapstructure:"changelog" yaml:"changelog"`
}

// ReviewConfig holds code review configuration
type ReviewConfig struct {
	SeverityThreshold string   `mapstructure:"severity_threshold" yaml:"severity_threshold" validate:"oneof=error warning info"`
	ExcludePatterns   []string `mapstructure:"exclude_patterns" yaml:"exclude_patterns"`
	IncludePatterns   []string `mapstructure:"include_patterns" yaml:"include_patterns"`
}

// CommitConfig holds commit command configuration
type CommitConfig struct {
	Style string `mapstructure:"style" yaml:"style"`
}

// TestGenConfig holds test generation configuration
type TestGenConfig struct {
	Framework string `mapstructure:"framework" yaml:"framework"`
	Output    string `mapstructure:"output" yaml:"output"`
	Coverage  bool   `mapstructure:"coverage" yaml:"coverage"`
}

// ChangelogConfig holds changelog command configuration
type ChangelogConfig struct {
	Format     string `mapstructure:"format" yaml:"format"`
	IncludePRs bool   `mapstructure:"include_prs" yaml:"include_prs"`
	Output     string `mapstructure:"output" yaml:"output"`
}

// GitHubConfig holds GitHub-specific configuration
type GitHubConfig struct {
	TokenEnv string `mapstructure:"token_env" yaml:"token_env"`
	Owner    string `mapstructure:"owner" yaml:"owner"`
	Repo     string `mapstructure:"repo" yaml:"repo"`
}

// OutputConfig holds output configuration
type OutputConfig struct {
	Format string `mapstructure:"format" yaml:"format" validate:"oneof=json yaml text"`
	Color  bool   `mapstructure:"color" yaml:"color"`
}
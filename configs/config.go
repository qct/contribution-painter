package configs

type GitInfo struct {
	RepoUrl string `mapstructure:"repo_url"`
	GhToken string `mapstructure:"gh_token"`
	Author  string `mapstructure:"author"`
	Email   string `mapstructure:"email"`
}

type Rewriter struct {
	DryRun                  bool   `mapstructure:"dry_run"`
	BackgroundCommitsPerDay int    `mapstructure:"background_commits_per_day"`
	ForegroundCommitsPerDay int    `mapstructure:"foreground_commits_per_day"`
	TargetLetters           string `mapstructure:"target_letters"`
	LeadingColumns          int    `mapstructure:"leading_columns"`
	TrailingColumns         int    `mapstructure:"trailing_columns"`
	LetterSpacing           int    `mapstructure:"letter_spacing"`
	Font                    string `mapstructure:"font"`
}

type Configuration struct {
	GitInfo  GitInfo  `mapstructure:"git_info"`
	Rewriter Rewriter `mapstructure:"rewriter"`
}

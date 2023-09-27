package configs

import "github.com/spf13/viper"

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
	WeekOffset              int    `mapstructure:"week_offset"`
	TargetLetters           string `mapstructure:"target_letters"`
}

type Configuration struct {
	GitInfo  GitInfo  `mapstructure:"git_info"`
	Rewriter Rewriter `mapstructure:"rewriter"`
}

func LoadConfig(file string, cfg *Configuration) error {
	if file != "" {
		viper.SetConfigFile(file)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	return nil
}

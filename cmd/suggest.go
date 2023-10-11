/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"fmt"
	"os"
	"rewriting-history/internal/pkg/graphql"
	"rewriting-history/internal/pkg/stat"

	"github.com/spf13/cobra"
)

// suggestCmd represents the suggest command
var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Give suggested config values",
	Long: `Give suggested config values for the following based on your existing commit history:
background_commits_per_day
foreground_commits_per_day
`,
	Run: suggestFunc,
}

var suggestFunc = func(cmd *cobra.Command, args []string) {
	ghGraphql := graphql.NewGhGraphql(config.GitInfo)
	stats := stat.NewContributionStats(ghGraphql)

	contributionStats, err := stats.GetContributionStats()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "get contribution stats failed:", err)
		os.Exit(1)
	}

	err = stats.PrintCommitStat(contributionStats...)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "print commit stat failed:", err)
		os.Exit(1)
	}

	cfg, err := stats.GetSuggestedConfig(contributionStats...)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "get suggested config failed:", err)
		os.Exit(1)
	}
	fmt.Println("suggested config:")
	fmt.Printf("background_commits_per_day: %d\n", cfg.BackgroundCommitsPerDay)
	fmt.Printf("foreground_commits_per_day: %d\n", cfg.ForegroundCommitsPerDay)
}

func init() {
	rootCmd.AddCommand(suggestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// suggestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// suggestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

package stat

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/graphql"
	"sort"

	"github.com/sirupsen/logrus"
)

var printFormat = "color%s(%s), total commits: %-4d, [%d, %d], mean: %d, median: %d"

type ContributionStats struct {
	ghGraphql *graphql.GhGraphql
}

func NewContributionStats(ghGraphql *graphql.GhGraphql) *ContributionStats {
	return &ContributionStats{ghGraphql: ghGraphql}
}

func (c *ContributionStats) PrintCommitStat(stats ...ContributionStat) (err error) {
	statsToPrint := stats
	if len(statsToPrint) == 0 {
		if statsToPrint, err = c.GetContributionStats(); err != nil {
			return fmt.Errorf("get contribution stats failed: %w", err)
		}
	}

	sortedStats := contributionStats(statsToPrint)
	sort.Sort(sort.Reverse(sortedStats))

	logrus.Info("color 1 --> 4, light to dark")
	for _, s := range sortedStats {
		logrus.Infof(printFormat, s.HumanReadableColor, s.Color, s.TotalCommits, s.Min, s.Max, s.Mean, s.Median)
	}

	return nil
}

func (c *ContributionStats) GetContributionStats() ([]ContributionStat, error) {
	resp, err := c.ghGraphql.GetContributionCollection()
	if err != nil {
		return nil, err
	}

	groupByColor := make(map[string]contributionDays)
	for _, week := range resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			groupByColor[day.Color] = append(groupByColor[day.Color], day)
		}
	}

	return convertToContributionStats(groupByColor), nil
}

func (c *ContributionStats) PrintSuggestedConfig(stats ...ContributionStat) error {
	statsToPrint := stats
	if len(statsToPrint) == 0 {
		var err error
		if statsToPrint, err = c.GetContributionStats(); err != nil {
			return fmt.Errorf("get contribution stats failed: %w", err)
		}
	}

	sortedStats := contributionStats(statsToPrint)
	sort.Sort(sort.Reverse(sortedStats))

	logrus.Info("suggested config values:")
	for _, s := range sortedStats {
		logrus.Infof("%s: %d", s.HumanReadableColor, s.Mean)
	}

	return nil
}

func (c *ContributionStats) GetSuggestedConfig(stats ...ContributionStat) (configs.Rewriter, error) {
	sortedStats := contributionStats(stats)
	sort.Sort(sort.Reverse(sortedStats))

	return configs.Rewriter{
		BackgroundCommitsPerDay: sortedStats[0].Median,
		ForegroundCommitsPerDay: sortedStats[1].Median,
	}, nil
}

func convertToContributionStats(contributionDaysWithColor map[string]contributionDays) []ContributionStat {
	var stats []ContributionStat
	for color, days := range contributionDaysWithColor {
		stats = append(stats, ContributionStat{
			Color:              color,
			HumanReadableColor: colorToHumanReadable[color],
			TotalCommits:       len(days),
			Min:                days.min(),
			Max:                days.max(),
			Mean:               days.mean(),
			Median:             days.median(),
		})
	}
	return stats
}

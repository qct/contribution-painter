package stat

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/graphql"
	"rewriting-history/internal/pkg/helper"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

var printFormat = "color%s(%s), total days: %-4d, [%d, %d], mean: %d, median: %d"

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

	logrus.Info("color 0 --> 4, light to dark")
	for _, s := range sortedStats {
		logrus.Infof(printFormat, s.HumanReadableColor, s.Color, s.TotalDays, s.Min, s.Max, s.Mean, s.Median)
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

func (c *ContributionStats) CommitsByDay() ([]CommitStat, error) {
	resp, err := c.ghGraphql.GetContributionCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to get contribution collection: %w", err)
	}

	// Process the response to retrieve the commit count per day
	contributionDays := resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks

	// Create an array of CommitStat structs
	var dailyCommits []CommitStat

	for _, week := range contributionDays {
		for _, day := range week.ContributionDays {
			date, err := time.Parse(helper.DateFormat, day.Date)
			if err != nil {
				logrus.WithError(err).Error("failed to parse date")
				continue
			}

			count := CommitStat{
				Date:    date,
				Commits: day.ContributionCount,
			}
			dailyCommits = append(dailyCommits, count)
		}
	}

	return dailyCommits, nil
}

func (c *ContributionStats) GetSuggestedConfig(stats ...ContributionStat) (configs.Rewriter, error) {
	var statsArray contributionStats
	for _, stat := range stats {
		if stat.Min == 0 && stat.Max == 0 {
			continue
		}
		statsArray = append(statsArray, stat)
	}
	sort.Sort(sort.Reverse(statsArray))

	return configs.Rewriter{
		BackgroundCommitsPerDay: statsArray[0].Median,
		ForegroundCommitsPerDay: statsArray[1].Median,
	}, nil
}

func convertToContributionStats(contributionDaysWithColor map[string]contributionDays) []ContributionStat {
	var stats []ContributionStat
	for color, days := range contributionDaysWithColor {
		stats = append(stats, ContributionStat{
			Color:              color,
			HumanReadableColor: colorToHumanReadable[color],
			TotalDays:          len(days),
			Min:                days.min(),
			Max:                days.max(),
			Mean:               days.mean(),
			Median:             days.median(),
		})
	}
	return stats
}

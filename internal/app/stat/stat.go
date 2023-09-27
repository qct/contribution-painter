package stat

import (
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/graphql"
	"sort"

	"github.com/sirupsen/logrus"
)

type Stat struct {
	ghGraphql *graphql.GhGraphql
}

func NewStat(c *configs.Config) *Stat {
	return &Stat{ghGraphql: graphql.NewGhGraphql(c)}
}

func (s *Stat) Run() error {
	logrus.Info("color 1 --> 4, light to dark")

	resp, err := s.ghGraphql.GetContributionCollection()
	if err != nil {
		return err
	}

	groupByColor := make(map[string][]graphql.ContributionDay)
	for _, week := range resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			groupByColor[day.Color] = append(groupByColor[day.Color], day)
		}
	}

	var stats contributionStats
	for color, days := range groupByColor {
		stats = append(stats, contributionStat{color: color, contributionDays: days})
	}

	sort.Sort(sort.Reverse(stats))
	for _, stat := range stats {
		// min, max, mean, median
		min := contributionDays(stat.contributionDays).min()
		max := contributionDays(stat.contributionDays).max()
		mean := contributionDays(stat.contributionDays).mean()
		median := contributionDays(stat.contributionDays).median()

		// Print the stats
		logrus.Infof("color%s(%s), total commits: %4d, [%d, %d], mean: %d, median: %d",
			colorToHuman[stat.color], stat.color, len(stat.contributionDays), min, max, mean, median)
	}
	return nil
}

package rewriter

import (
	"sort"

	"github.com/sirupsen/logrus"
)

var (
	colorToHuman = map[string]string{
		"#216e39": "4",
		"#30a14e": "3",
		"#40c463": "2",
		"#9be9a8": "1",
	}
	printFormat = "color%s(%s), total commits: %-4d, [%d, %d], mean: %d, median: %d"
)

func (r *Rewriter) printCommitStat() error {
	logrus.Info("color 1 --> 4, light to dark")

	resp, err := r.ghGraphql.GetContributionCollection()
	if err != nil {
		return err
	}

	groupByColor := make(map[string]contributionDays)
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
		min := stat.contributionDays.min()
		max := stat.contributionDays.max()
		mean := stat.contributionDays.mean()
		median := stat.contributionDays.median()

		// Print the stats
		logrus.Infof(printFormat, colorToHuman[stat.color], stat.color, len(stat.contributionDays), min, max, mean, median)
	}
	return nil
}

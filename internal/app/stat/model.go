package stat

import "rewriting-history/internal/pkg/graphql"

var colorToHuman = map[string]string{
	"#216e39": "4",
	"#30a14e": "3",
	"#40c463": "2",
	"#9be9a8": "1",
}

type contributionDays []graphql.ContributionDay

func (c contributionDays) min() int {
	min := c[0].ContributionCount
	for _, day := range c {
		if day.ContributionCount < min {
			min = day.ContributionCount
		}
	}
	return min
}

func (c contributionDays) max() int {
	max := c[0].ContributionCount
	for _, day := range c {
		if day.ContributionCount > max {
			max = day.ContributionCount
		}
	}
	return max
}

func (c contributionDays) mean() int {
	var sum int
	for _, day := range c {
		sum += day.ContributionCount
	}
	return sum / len(c)
}

func (c contributionDays) median() int {
	return c[len(c)/2].ContributionCount
}

type contributionStat struct {
	color            string
	contributionDays []graphql.ContributionDay
}

// implement sort
type contributionStats []contributionStat

func (c contributionStats) Len() int { return len(c) }
func (c contributionStats) Less(i, j int) bool {
	return len(c[i].contributionDays) < len(c[j].contributionDays)
}
func (c contributionStats) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

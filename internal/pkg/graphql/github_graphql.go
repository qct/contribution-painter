package graphql

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/helper"
	"time"

	"github.com/sirupsen/logrus"
)

type GhGraphql struct {
	User string
	C    *GraphClient
}

func NewGhGraphql(config configs.GitInfo) *GhGraphql {
	return &GhGraphql{
		User: config.Author,
		C:    NewClient(helper.GitHubGraphQLEndpoint, config.GhToken, 10*time.Second),
	}
}

func (g *GhGraphql) CommitsByDay() ([]CommitStats, error) {
	resp, err := g.GetContributionCollection()
	if err != nil {
		return nil, fmt.Errorf("failed to get contribution collection: %w", err)
	}

	// Process the response to retrieve the commit count per day
	contributionDays := resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks

	// Create an array of CommitStats structs
	var dailyCommits []CommitStats

	for _, week := range contributionDays {
		for _, day := range week.ContributionDays {
			date, err := time.Parse(helper.DateFormat, day.Date)
			if err != nil {
				logrus.WithError(err).Error("failed to parse date")
				continue
			}

			count := CommitStats{
				Date:    date,
				Commits: day.ContributionCount,
			}
			dailyCommits = append(dailyCommits, count)
		}
	}

	return dailyCommits, nil
}

func (g *GhGraphql) GetContributionCollection() (ContributionsCollectionResp, error) {
	query := fmt.Sprintf(`
	{
		user(login: "%s") {
			contributionsCollection {
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							date
							contributionCount
							color
						}
					}
				}
			}
		}
	}`, g.User)

	var resp ContributionsCollectionResp
	err := g.C.GraphQLRequest(query, &resp)
	if err != nil {
		return ContributionsCollectionResp{}, err
	}
	return resp, nil
}

func MaxCommits(from, to time.Time, dailyCommits []CommitStats) (*CommitStats, error) {
	var max *CommitStats

	for _, dc := range dailyCommits {
		if (dc.Date.After(from) || dc.Date.Equal(from)) && (dc.Date.Before(to) || dc.Date.Equal(to)) {
			if max == nil || dc.Commits > max.Commits {
				temp := dc
				max = &temp
			}
		}
	}

	if max == nil {
		return nil, fmt.Errorf("no commits found between %s and %s", from.Format(helper.DateFormat), to.Format(helper.DateFormat))
	}

	return max, nil
}

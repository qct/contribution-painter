package graphql

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/helper"
	"time"
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

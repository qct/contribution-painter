package graphql

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/helper"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type GhGraphql struct {
	user    string
	ghToken string
}

func NewGhGraphql(config *configs.Config) *GhGraphql {
	return &GhGraphql{
		user:    config.Author,
		ghToken: config.GhToken,
	}
}

func (g *GhGraphql) CommitsByDay() ([]DailyCommit, error) {
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
	}`, g.user)

	// Create the GraphQL request payload
	reqJSON, err := json.Marshal(graphqlRequest{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{}
	// Create the HTTP POST request to the GraphQL API
	req, err := http.NewRequest("POST", helper.GitHubGraphQLEndpoint, strings.NewReader(string(reqJSON)))
	if err != nil {
		log.Fatal("Failed to create HTTP request:", err)
	}
	// Set the necessary headers, including the access token
	req.Header.Set("Authorization", "Bearer "+g.ghToken)
	req.Header.Set("Content-Type", helper.ContentTypeJSON)
	req.Header.Set("Accept", helper.ContentTypeJSON)

	// Send the HTTP request
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("failed to get response: status code %d, status %s", resp.StatusCode, resp.Status)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	// Parse the GraphQL response
	var graphResp graphqlResponse
	err = json.NewDecoder(resp.Body).Decode(&graphResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process the response to retrieve the commit count per day
	contributionDays := graphResp.Data.User.ContributionsCollection.ContributionCalendar.Weeks

	// Create an array of DailyCommit structs
	var dailyCommits []DailyCommit

	for _, week := range contributionDays {
		for _, day := range week.ContributionDays {
			date, err := time.Parse(helper.DateFormat, day.Date)
			if err != nil {
				logrus.WithError(err).Error("failed to parse date")
				continue
			}

			count := DailyCommit{
				Date:    date,
				Commits: day.ContributionCount,
			}
			dailyCommits = append(dailyCommits, count)
		}
	}

	return dailyCommits, nil
}

func MaxCommits(from, to time.Time, dailyCommits []DailyCommit) (*DailyCommit, error) {
	var max *DailyCommit

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

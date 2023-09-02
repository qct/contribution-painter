package graphql

import (
	"time"
)

type CommitStats struct {
	Date    time.Time
	Commits int
}

type graphqlRequest struct {
	Query string `json:"query"`
}

type ContributionsCollectionResp struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []ContributionDay `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
}

type ContributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
	Color             string `json:"color"`
}

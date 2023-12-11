package stat

import (
	"contribution-painter/internal/pkg/graphql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestContributionStats_PrintCommitStat(t *testing.T) {
	resp := &graphql.ContributionsCollectionResp{}
	err := jsonModelFromFilePath("mocks/contributions_collection_resp.json", resp)
	assert.NoError(t, err)

	groupByColor := make(map[string]contributionDays)
	for _, week := range resp.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			groupByColor[day.Color] = append(groupByColor[day.Color], day)
		}
	}

	sortedStats := contributionStats(convertToContributionStats(groupByColor))
	sort.Sort(sort.Reverse(sortedStats))

	logrus.Info("color 0 --> 4, light to dark")
	for _, s := range sortedStats {
		logrus.Infof(printFormat, s.HumanReadableColor, s.Color, s.TotalDays, s.Min, s.Max, s.Mean, s.Median)
	}
}

func TestCommitsByDay(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    []CommitStat
		wantErr error
	}{
		{
			name: "should succeed",
			handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte(`{
					"data": {
						"user": {
							"contributionsCollection": {
								"contributionCalendar": {
									"totalContributions": 65,
									"weeks": [
										{
											"contributionDays": [
												{
													"date": "2023-06-18",
													"contributionCount": 55,
													"color": "#c6e48b"
												},
												{
													"date": "2023-06-19",
													"contributionCount": 10,
													"color": "#c6e48b"
												}
											]
										}
									]
								}
							}
						}
					}
				}`))
			},
			want: []CommitStat{
				{Date: time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC), Commits: 55},
				{Date: time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC), Commits: 10},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(tt.handler)
			defer mockServer.Close()

			c := NewContributionStats(&graphql.GhGraphql{
				C: &graphql.GraphClient{
					Url:    mockServer.URL,
					Client: &http.Client{},
				},
			})

			commitStats, err := c.CommitsByDay()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.ElementsMatchf(t, tt.want, commitStats, "CommitsByDay()")
		})
	}
}

func jsonModelFromFilePath(file string, result interface{}) error {
	// Open the file
	jsonFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	// Read the file
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	// Unmarshal the JSON
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return err
	}

	return nil
}

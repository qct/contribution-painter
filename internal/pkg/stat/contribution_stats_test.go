package stat

import (
	"encoding/json"
	"io"
	"os"
	"rewriting-history/internal/pkg/graphql"
	"sort"
	"testing"

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

	logrus.Info("color 1 --> 4, light to dark")
	for _, s := range sortedStats {
		logrus.Infof(printFormat, s.HumanReadableColor, s.Color, s.TotalCommits, s.Min, s.Max, s.Mean, s.Median)
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

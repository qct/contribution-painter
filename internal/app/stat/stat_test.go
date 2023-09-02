package stat

import (
	"encoding/json"
	"io"
	"os"
	"rewriting-history/internal/pkg/graphql"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorStats(t *testing.T) {
	resp := &graphql.ContributionsCollectionResp{}
	err := jsonModelFromFilePath("mocks/contributions_collection_resp.json", resp)
	assert.NoError(t, err)

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

	// Sort the stats
	sort.Sort(stats)
	for _, stat := range stats {
		// min, max, mean, median
		min := contributionDays(stat.contributionDays).min()
		max := contributionDays(stat.contributionDays).max()
		mean := contributionDays(stat.contributionDays).mean()
		median := contributionDays(stat.contributionDays).median()

		// Print the stats
		t.Logf("%s(%s), count: %d, {%d, %d}, mean: %d, median: %d",
			colorToHuman[stat.color], stat.color, len(stat.contributionDays), min, max, mean, median)
	}
}

func jsonModelFromFilePath(file string, result interface{}) error {
	// Open the file
	jsonFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

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

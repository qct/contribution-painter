package graphql

import (
	"rewriting-history/configs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitsByDay(t *testing.T) {
	cfg := loadConfig(t)

	tests := []struct {
		name    string
		user    string
		token   string
		want    []DailyCommit
		wantErr error
	}{
		{
			name:  "should succeed",
			user:  "qct",
			token: cfg.GhToken,
			want: []DailyCommit{
				{
					//2023-06-18 00:00:00 +0000
					Date:    time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
					Commits: 55,
				},
			},
			wantErr: nil,
		},
	}

	ghGraphql := NewGhGraphql(&cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ghGraphql.CommitsByDay()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Contains(t, got, tt.want[0])
		})
	}
}

func TestMaxCommits(t *testing.T) {
	type args struct {
		from         time.Time
		to           time.Time
		dailyCommits []DailyCommit
	}
	tests := []struct {
		name    string
		args    args
		want    *DailyCommit
		wantErr error
	}{
		{
			name: "should succeed",
			args: args{
				from: time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
				to:   time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC),
				dailyCommits: []DailyCommit{
					{
						Date:    time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
						Commits: 55,
					},
					{
						Date:    time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC),
						Commits: 10,
					},
				},
			},
			want: &DailyCommit{
				Date:    time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
				Commits: 55,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MaxCommits(tt.args.from, tt.args.to, tt.args.dailyCommits)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equalf(t, tt.want, got, "MaxCommits(%v, %v, %v)", tt.args.from, tt.args.to, tt.args.dailyCommits)
		})
	}
}

func loadConfig(t *testing.T) configs.Config {
	var cfg configs.Config
	err := configs.LoadConfig("../../../configs/config.yaml", &cfg)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}
	return cfg
}

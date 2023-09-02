package graphql

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitsByDay(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    []CommitStats
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
			want: []CommitStats{
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

			g := &GhGraphql{
				C: &GraphClient{
					Url:    mockServer.URL,
					Client: &http.Client{},
				},
			}

			commitStats, err := g.CommitsByDay()

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

func TestMaxCommits(t *testing.T) {
	type args struct {
		from         time.Time
		to           time.Time
		dailyCommits []CommitStats
	}
	tests := []struct {
		name    string
		args    args
		want    *CommitStats
		wantErr error
	}{
		{
			name: "should succeed",
			args: args{
				from: time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC),
				to:   time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC),
				dailyCommits: []CommitStats{
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
			want: &CommitStats{
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

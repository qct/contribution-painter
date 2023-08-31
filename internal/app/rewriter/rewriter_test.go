package rewriter

import (
	"reflect"
	"testing"
	"time"
)

func Test_getSunday(t *testing.T) {
	tests := []struct {
		name       string
		now        time.Time
		weekOffset int
		want       time.Time
		wantErr    bool
	}{
		{
			name:       "0 week offset should return the Sunday of the current week",
			now:        time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC),
			weekOffset: 0,
			want:       time.Date(2022, 8, 28, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
		},
		{
			name:       "1 week offset should return the Sunday of the previous week",
			now:        time.Date(2023, 8, 30, 0, 0, 0, 0, time.UTC),
			weekOffset: 1,
			want:       time.Date(2022, 9, 4, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
		},
		{
			name:       "2 week offset should return the Sunday of the week before the previous week",
			now:        time.Date(2023, 8, 30, 0, 0, 0, 0, time.UTC),
			weekOffset: 2,
			want:       time.Date(2022, 9, 11, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSunday(tt.now, tt.weekOffset)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSunday() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSunday() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_latestSunday(t *testing.T) {
	tests := []struct {
		name string
		date *time.Time
		want time.Time
	}{
		{
			name: "latestSunday should return the latest Sunday",
			date: func() *time.Time {
				date := time.Date(2023, 7, 2, 0, 0, 0, 0, time.UTC)
				return &date
			}(),
			want: time.Date(2023, 7, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "latestSunday should return the latest Sunday",
			date: func() *time.Time {
				date := time.Date(2023, 7, 3, 0, 0, 0, 0, time.UTC)
				return &date
			}(),
			want: time.Date(2023, 7, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "latestSunday should return the latest Sunday",
			date: func() *time.Time {
				date := time.Date(2023, 7, 8, 0, 0, 0, 0, time.UTC)
				return &date
			}(),
			want: time.Date(2023, 7, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "latestSunday should return the latest Sunday",
			date: func() *time.Time {
				date := time.Date(2023, 7, 9, 0, 0, 0, 0, time.UTC)
				return &date
			}(),
			want: time.Date(2023, 7, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "latestSunday should return the latest Sunday",
			date: func() *time.Time {
				date := time.Date(2023, 7, 15, 0, 0, 0, 0, time.UTC)
				return &date
			}(),
			want: time.Date(2023, 7, 9, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := latestSunday(tt.date); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("latestSunday() = %v, want %v", got, tt.want)
			}
		})
	}
}

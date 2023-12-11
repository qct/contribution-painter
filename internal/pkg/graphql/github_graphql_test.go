package graphql

import (
	"reflect"
	"testing"
)

func TestGhGraphql_GetContributionCollection(t *testing.T) {
	type fields struct {
		User string
		C    *GraphClient
	}
	tests := []struct {
		name    string
		fields  fields
		want    ContributionsCollectionResp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GhGraphql{
				User: tt.fields.User,
				C:    tt.fields.C,
			}
			got, err := g.GetContributionCollection()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContributionCollection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContributionCollection() got = %v, want %v", got, tt.want)
			}
		})
	}
}

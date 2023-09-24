package dict

import (
	"rewriting-history/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDictionary_GetLetter(t *testing.T) {
	type args struct {
		font          domain.Font
		target        string
		letterSpacing int
		leadingSpace  int
		trailingSpace int
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.Letter
		wantErr error
	}{
		{
			name: "test letter A of Font75",
			args: args{
				font:          domain.Font75,
				target:        "AB",
				letterSpacing: 1,
				leadingSpace:  0,
				trailingSpace: 0,
			},
			want:    []domain.Letter{L75A, L7Space, L75B},
			wantErr: nil,
		},
		{
			name: "test letter A of Font55",
			args: args{
				font:          domain.Font55,
				target:        "A",
				letterSpacing: 1,
				leadingSpace:  0,
				trailingSpace: 0,
			},
			want:    []domain.Letter{L55A},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dictionary{
				font: tt.args.font,
			}
			got, err := d.GetLetter(tt.args.target, tt.args.letterSpacing, tt.args.leadingSpace, tt.args.trailingSpace)

			if tt.wantErr != nil {
				assert.Error(t, err)
			}

			assert.ElementsMatchf(t, tt.want, got, "want: %v, got: %v", tt.want, got)
		})
	}
}

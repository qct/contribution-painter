package simulate

import (
	"errors"
	"rewriting-history/internal/domain"
	"rewriting-history/internal/pkg/dict"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_printMatrix(t *testing.T) {

	tests := []struct {
		name    string
		letters []domain.Letter
	}{
		{
			name: "test letters of 7x5",
			letters: []domain.Letter{
				dict.L75A,
				dict.L75B,
				dict.L75C,
				dict.L75D,
				dict.L75E,
				dict.L75F,
				dict.L75G,
				dict.L75H,
				dict.L75I,
				dict.L75J,
				dict.L75K,
				dict.L75L,
				dict.L75M,
				dict.L75N,
				dict.L75O,
				dict.L75P,
				dict.L75Q,
				dict.L75R,
				dict.L75S,
				dict.L75T,
				dict.L75U,
				dict.L75V,
				dict.L75W,
				dict.L75X,
				dict.L75Y,
				dict.L75Z,
			},
		},
		{
			name: "test letters of 5x5",
			letters: []domain.Letter{
				dict.L55A,
				dict.L55B,
				dict.L55C,
				dict.L55D,
				dict.L55E,
				dict.L55F,
				dict.L55G,
				dict.L55H,
				dict.L55I,
				dict.L55J,
				dict.L55K,
				dict.L55L,
				dict.L55M,
				dict.L55N,
				dict.L55O,
				dict.L55P,
				dict.L55Q,
				dict.L55R,
				dict.L55S,
				dict.L55T,
				dict.L55U,
				dict.L55V,
				dict.L55W,
				dict.L55X,
				dict.L55Y,
				dict.L55Z,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, letter := range tt.letters {
				printMatrix(letter)
				println()
			}
		})
	}
}

func TestSimulator_Simulate(t *testing.T) {
	type args struct {
		target        string
		leadingSpace  int
		letterSpacing int
		trailingSpace int
		topSpace      int
	}
	tests := []struct {
		name     string
		bgLength int
		bgHeight int
		font     domain.Font
		args     args
		wantErr  error
	}{
		{
			name:     "test simulate font 7x5",
			bgLength: 52,
			bgHeight: 7,
			font:     domain.Font75,
			args: args{
				target:        "HELLO",
				leadingSpace:  0,
				letterSpacing: 2,
				trailingSpace: 0,
				topSpace:      0,
			},
		},
		{
			name:     "test simulate font 5x5",
			bgLength: 52,
			bgHeight: 7,
			font:     domain.Font55,
			args: args{
				target:        "HELLO",
				leadingSpace:  0,
				letterSpacing: 2,
				trailingSpace: 0,
				topSpace:      2,
			},
		},
		{
			name:     "too much top space should return error",
			bgLength: 52,
			bgHeight: 7,
			font:     domain.Font55,
			args: args{
				target:        "HELLO",
				leadingSpace:  0,
				letterSpacing: 2,
				trailingSpace: 0,
				topSpace:      3,
			},
			wantErr: errors.New("letters are too high: 5, 7"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Simulator{
				bgLength: tt.bgLength,
				bgHeight: tt.bgHeight,
				dict:     dict.NewDictionary(tt.font),
			}

			err := s.Simulate(tt.args.target, tt.args.letterSpacing, tt.args.leadingSpace, tt.args.trailingSpace, tt.args.topSpace)

			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

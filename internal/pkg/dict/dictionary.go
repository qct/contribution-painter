package dict

import (
	"fmt"
	"rewriting-history/internal/domain"
)

var (
	font75Map = map[rune]domain.Letter{'A': L75A, 'B': L75B, 'C': L75C, 'D': L75D, 'E': L75E, 'F': L75F, 'G': L75G, 'H': L75H, 'I': L75I, 'J': L75J, 'K': L75K, 'L': L75L, 'M': L75M, 'N': L75N, 'O': L75O, 'P': L75P, 'Q': L75Q, 'R': L75R, 'S': L75S, 'T': L75T, 'U': L75U, 'V': L75V, 'W': L75W, 'X': L75X, 'Y': L75Y, 'Z': L75Z, ' ': L7Space}
	font55Map = map[rune]domain.Letter{'A': L55A, 'B': L55B, 'C': L55C, 'D': L55D, 'E': L55E, 'F': L55F, 'G': L55G, 'H': L55H, 'I': L55I, 'J': L55J, 'K': L55K, 'L': L55L, 'M': L55M, 'N': L55N, 'O': L55O, 'P': L55P, 'Q': L55Q, 'R': L55R, 'S': L55S, 'T': L55T, 'U': L55U, 'V': L55V, 'W': L55W, 'X': L55X, 'Y': L55Y, 'Z': L55Z, ' ': L5Space}
)

type Dictionary struct {
	font domain.Font
}

func NewDictionary(font domain.Font) *Dictionary {
	return &Dictionary{font: font}
}

func (d *Dictionary) GetLetter(target string, letterSpacing, leadingSpace, trailingSpace int) ([]domain.Letter, error) {
	var letters []domain.Letter

	var fontMap map[rune]domain.Letter
	switch d.font {
	case domain.Font75:
		fontMap = font75Map
	case domain.Font55:
		fontMap = font55Map
	default:
		return nil, fmt.Errorf("unknown font: %s", d.font)
	}

	// add leading space
	for i := 0; i < leadingSpace; i++ {
		letters = append(letters, fontMap[' '])
	}

	for _, r := range []rune(target) {
		letter, ok := fontMap[r]
		if !ok {
			return nil, fmt.Errorf("unknown letter: %s", string(r))
		}
		letters = append(letters, letter)

		// add space
		for i := 0; i < letterSpacing; i++ {
			letters = append(letters, fontMap[' '])
		}
	}

	// remove last space
	if len(letters) > 0 {
		letters = letters[:len(letters)-letterSpacing]
	}

	// add trailing space
	for i := 0; i < trailingSpace; i++ {
		letters = append(letters, fontMap[' '])
	}

	return letters, nil
}

func (d *Dictionary) FontHeight() int {
	return d.font.Height()
}

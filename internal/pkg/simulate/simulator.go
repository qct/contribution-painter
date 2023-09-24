package simulate

import (
	"errors"
	"fmt"
	"rewriting-history/internal/domain"
	"rewriting-history/internal/pkg/dict"

	"github.com/sirupsen/logrus"
)

// unicode block square
const (
	//bgIcon     = "\u2592 "
	//targetIcon = "\u2588 "
	bgIcon     = "\u25a2 "
	targetIcon = "\u2588 "

	minLength = 5
	minHeight = 5
)

func printMatrix(matrix [][]uint) {
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(matrix[i]); j++ {
			if matrix[i][j] == 1 {
				print(targetIcon)
			} else {
				print(bgIcon)
			}
		}
		println()
	}
}

type Simulator struct {
	bgLength int
	bgHeight int

	dict domain.Dictionary
}

func NewSimulator(length, height int, font domain.Font) *Simulator {
	if length < minLength {
		logrus.Warnf("length is too short: %d, set to %d", length, minLength)
		length = minLength
	}

	if height < minHeight {
		logrus.Warnf("height is too short: %d, set to %d", height, minHeight)
		height = minHeight
	}

	f := font
	if f == "" {
		f = domain.Font75
	}
	return &Simulator{bgLength: length, bgHeight: height, dict: dict.NewDictionary(f)}
}

func (s *Simulator) Simulate(target string, letterSpacing, leadingSpace, trailingSpace, topSpace int) error {
	if len(target) == 0 {
		return errors.New("target is empty")
	}

	letters, err := s.dict.GetLetter(target, letterSpacing, leadingSpace, trailingSpace)
	if err != nil {
		return fmt.Errorf("failed to get letters: %w", err)
	}

	if s.dict.FontHeight()+topSpace > s.bgHeight {
		return fmt.Errorf("letters are too high: %d, %d", s.dict.FontHeight(), s.bgHeight)
	}

	// concat letters
	matrix := make([][]uint, s.bgHeight)
	column := 0
	for _, letter := range letters {
		if column > s.bgLength {
			return fmt.Errorf("letters are too long: %d, %d", column, s.bgLength)
		}

		// add top space
		letterWidth := letter[0]
		for i := 0; i < topSpace; i++ {
			matrix[i] = make([]uint, column+len(letterWidth))
		}

		// add letters to matrix
		for j := topSpace; j < len(matrix); j++ {
			if j-topSpace >= len(letter) {
				matrix[j] = append(matrix[j], make([]uint, len(letterWidth))...)
				continue
			}
			if j-topSpace >= len(letter) {
				continue
			}
			matrix[j] = append(matrix[j], letter[j-topSpace]...)
		}

		column += len(letterWidth)
	}

	printMatrix(matrix)
	return nil
}

package domain

const (
	Font75 Font = "75"
	Font55 Font = "55"
)

// Letter is an m*n 2D array of bools, where m is the height of the letter and n is the width of the letter
// 1 means the dot is filled
// 0 means the dot is empty
type Letter [][]uint

type Letters []Letter

func (l Letters) Length() int {
	length := 0
	for _, letter := range l {
		length += len(letter[0])
	}
	return length
}

// Font includes 7*5 & 5*5
type Font string

func (f Font) Height() int {
	switch f {
	case Font75:
		return 7
	case Font55:
		return 5
	default:
		return 1<<31 - 1
	}
}

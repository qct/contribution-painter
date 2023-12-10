package domain

type Dictionary interface {
	// GetLetters returns the letter with the given string
	GetLetters(target string, letterSpacing, leadingSpace, trailingSpace int) ([]Letter, error)

	// FontHeight returns the length of the font
	FontHeight() int
}

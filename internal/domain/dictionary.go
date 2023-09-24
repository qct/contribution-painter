package domain

type Dictionary interface {
	// GetLetter returns the letter with the given string
	GetLetter(target string, letterSpacing, leadingSpace, trailingSpace int) ([]Letter, error)

	// FontHeight returns the length of the font
	FontHeight() int
}

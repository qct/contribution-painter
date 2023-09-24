package domain

type Simulator interface {
	// Simulate should print the simulation of the target letters to the console
	Simulate(targetLetters string)
}

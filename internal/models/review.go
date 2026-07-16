package models

type Review struct {
	ID      uint
	MovieID uint
	Score   int
	Text    string
}

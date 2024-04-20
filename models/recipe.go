package models

type Ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

type Recipe struct {
	Name        string
	Description string
	Ingredients []Ingredient
	Steps       []string
	Categories  []string

	Link string
}

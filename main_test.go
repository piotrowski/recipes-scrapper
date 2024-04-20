package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseIngredient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  ingredient
	}{
		{
			input: "5g black pepper",
			want: ingredient{
				Name:   "black pepper",
				Amount: 5,
				Unit:   "g",
			},
		}, {
			input: "0.5 teaspoon bicarbonate of soda",
			want: ingredient{
				Name:   "bicarbonate of soda",
				Amount: 0.5,
				Unit:   "teaspoon",
			},
		}, {
			input: "200 g white beans",
			want: ingredient{
				Name:   "white beans",
				Amount: 200,
				Unit:   "g",
			},
		}, {
			input: "2 cups low-sodium chicken broth",
			want: ingredient{
				Name:   "low-sodium chicken broth",
				Amount: 2,
				Unit:   "cup",
			},
		}, {
			input: "2 cans (2 lbs) white beans, such as cannellini, drained",
			want: ingredient{
				Name:   "white beans, such as cannellini, drained",
				Amount: 2,
				Unit:   "lbs",
			},
		}, {
			input: "1 egg",
			want: ingredient{
				Name:   "egg",
				Amount: 1,
			},
		}, {
			input: "Brewed black tea",
			want: ingredient{
				Name: "Brewed black tea",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, parseIngredient(tt.input))
		})
	}
}

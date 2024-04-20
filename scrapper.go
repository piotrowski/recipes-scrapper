package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/piotrowski/recipes-scrapper/models"
	"github.com/piotrowski/recipes-scrapper/scrappers"
)

var (
	_ scrapper = scrappers.NewWikiBooks(nil, "")
)

type scrapper interface {
	GetRecipesLinks() []string
	GetRecipesFromURLs(list []string) []models.Recipe
}

func scrapWebsite(pathToFile string, scr scrapper) {
	links := scr.GetRecipesLinks()
	recipes := scr.GetRecipesFromURLs(links)

	recipesData, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		log.Fatal("Could not marshal recipes: ", err)
	}
	os.WriteFile(pathToFile, recipesData, os.FileMode(0644))
}

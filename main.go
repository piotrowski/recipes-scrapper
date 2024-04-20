package main

import (
	"time"

	"github.com/gocolly/colly"
	"github.com/piotrowski/recipes-scrapper/scrappers"
)

var ignoreList = map[string]bool{
	"https://en.wikibooks.org/wiki/Cookbook:Wheat-Free_Baking_Mix": true,
	"https://en.wikibooks.org/wiki/Cookbook:White_Bread":           true,
	"https://en.wikibooks.org/wiki/Cookbook:Whole_Wheat_Bread":     true,
}

func main() {
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{
		RandomDelay: time.Millisecond * 250,
	})
	c.CacheDir = "./tmp"

	wikiBooks := scrappers.NewWikiBooks(c, "https://en.wikibooks.org/wiki/Category:Recipes")
	scrapWebsite("./recipes.json", wikiBooks)
}

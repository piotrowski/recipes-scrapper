package scrappers

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/piotrowski/recipes-scrapper/models"
)

type wikiBooks struct {
	c *colly.Collector

	startURL   string
	ignoreURLs map[string]bool
}

func NewWikiBooks(c *colly.Collector, startFrom string) *wikiBooks {
	return &wikiBooks{
		c: c,

		startURL: startFrom,
		ignoreURLs: map[string]bool{
			"https://en.wikibooks.org/wiki/Cookbook:Wheat-Free_Baking_Mix": true,
			"https://en.wikibooks.org/wiki/Cookbook:White_Bread":           true,
			"https://en.wikibooks.org/wiki/Cookbook:Whole_Wheat_Bread":     true,
		},
	}
}

func (wb wikiBooks) GetRecipesLinks() []string {
	recipePagesList := []string{}
	wb.c.OnHTML("#mw-pages", func(e *colly.HTMLElement) {
		nextPageURL, isFound := wb.findNextPageInHTML(e)
		if isFound {
			e.Request.Visit(nextPageURL)
		}

		links := wb.findRecipeLinksInHTML(e)
		recipePagesList = append(recipePagesList, links...)
	})
	wb.c.Visit("https://en.wikibooks.org/wiki/Category:Recipes")

	return recipePagesList
}

func (wikiBooks) findNextPageInHTML(e *colly.HTMLElement) (string, bool) {
	var uri string

	nextPageSelector := e.DOM.ChildrenFiltered("#mw-pages > a")
	nextPageSelector.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.Text() != "next page" {
			return true
		}

		uri, _ = s.Attr("href")
		return false
	})

	if uri == "" {
		return "", false
	}

	return e.Request.AbsoluteURL(uri), true
}

func (wb wikiBooks) findRecipeLinksInHTML(e *colly.HTMLElement) []string {
	recipes := []string{}
	e.ForEach("#mw-pages > div.mw-content-ltr > div > div > ul > li > a", func(i int, h *colly.HTMLElement) {
		uri := h.Attr("href")
		if uri == "" {
			return
		}
		url := e.Request.AbsoluteURL(uri)
		if wb.ignoreURLs[url] {
			return
		}
		recipes = append(recipes, url)
	})

	return recipes
}

func (wb wikiBooks) GetRecipesFromURLs(list []string) []models.Recipe {
	recipes := []models.Recipe{}
	wb.c.OnHTML("body", func(e *colly.HTMLElement) {
		url := e.Request.URL
		r, isOk := wb.getRecipe(url.Scheme+"://"+url.Host+url.Path, e)
		if !isOk {
			return
		}
		recipes = append(recipes, r)
	})
	for _, r := range list {
		wb.c.Visit(r)
	}
	return recipes
}

func (wb wikiBooks) getRecipe(link string, e *colly.HTMLElement) (models.Recipe, bool) {
	r := models.Recipe{Link: link}

	recipeName := e.ChildText("#firstHeading")
	recipeName, _ = strings.CutPrefix(recipeName, "Cookbook:")
	r.Name = recipeName

	l := e.DOM.Find("#mw-content-text > div.mw-content-ltr.mw-parser-output")
	l = l.Children()
	startScraping := false
	l.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.HasPrefix(s.Text(), "Ingredients") {
			startScraping = true
			return true
		}
		if !startScraping {
			return true
		}

		if s.Is("h2") {
			return false
		}
		if !s.Is("ul") {
			return true
		}

		ingredientsStr := s.Text()
		ingredients := strings.Split(ingredientsStr, "\n")
		for _, i := range ingredients {
			r.Ingredients = append(r.Ingredients, wb.parseIngredient(i))
		}

		return true
	})
	description := e.ChildText("#mw-content-text > div.mw-content-ltr.mw-parser-output > p:nth-child(4)")
	r.Description = description

	procedureStr := e.ChildText("#mw-content-text > div.mw-content-ltr.mw-parser-output > ol")
	procedure := strings.Split(procedureStr, "\n")
	r.Steps = procedure

	e.ForEach("#mw-normal-catlinks > ul > li", func(i int, h *colly.HTMLElement) {
		if h.Text == "Recipes" {
			return
		}

		r.Categories = append(r.Categories, h.Text)
	})

	if slices.Contains(r.Categories, "Incomplete recipes") {
		return models.Recipe{}, false
	}

	return r, true
}

var (
	regexIngredientWithAmountUnitName = regexp.MustCompile(`(\d*\.?\d+)\D*(cup|lbs|tablespoon|tbsp|teaspoon|oz|g|gram)[s).,]?\s+(.+)`)
	regexIngredientWithAmountName     = regexp.MustCompile(`(\d*\.?\d+)\s(.+)`)
)

func (wb wikiBooks) parseIngredient(input string) models.Ingredient {
	input = strings.Replace(input, "⅛", "0.125", 1)
	input = strings.Replace(input, "¼", "0.25", 1)
	input = strings.Replace(input, "⅓", "0.33", 1)
	input = strings.Replace(input, "½", "0.5", 1)
	input = strings.Replace(input, "⅔", "0.66", 1)
	input = strings.Replace(input, "¾", "0.75", 1)

	ing, isOk := wb.parseIngredientWithAmountUnitName(input)
	if isOk {
		return ing
	}

	ing, isOk = wb.parseIngredientWithAmountName(input)
	if isOk {
		return ing
	}

	return models.Ingredient{
		Name: input,
	}
}

func (wb wikiBooks) parseIngredientWithAmountUnitName(input string) (models.Ingredient, bool) {
	matches := regexIngredientWithAmountUnitName.FindStringSubmatch(input)
	if len(matches) < 4 {
		return models.Ingredient{}, false
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	return models.Ingredient{
		Name:   matches[3],
		Amount: amount,
		Unit:   matches[2],
	}, true
}

func (wb wikiBooks) parseIngredientWithAmountName(input string) (models.Ingredient, bool) {
	matches := regexIngredientWithAmountName.FindStringSubmatch(input)
	if len(matches) < 3 {
		return models.Ingredient{}, false
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	return models.Ingredient{
		Name:   matches[2],
		Amount: amount,
	}, true
}

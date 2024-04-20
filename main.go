package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

type recipe struct {
	Name        string
	Description string
	Ingredients []ingredient
	Steps       []string
	Categories  []string

	Link string
}

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

	recipePagesList := []string{}
	c.OnHTML("#mw-pages", func(e *colly.HTMLElement) {
		nextPageURL, isFound := findNextPageURL(e)
		if isFound {
			e.Request.Visit(nextPageURL)
		}

		currentPageRecipeList := getRecipesLinkList(e)
		recipePagesList = append(recipePagesList, currentPageRecipeList...)
	})
	c.Visit("https://en.wikibooks.org/wiki/Category:Recipes")

	recipes := []recipe{}
	c.OnHTML("body", func(e *colly.HTMLElement) {
		url := e.Request.URL
		r, isOk := scrapRecipeData(url.Scheme+"://"+url.Host+url.Path, e)
		if !isOk {
			return
		}
		recipes = append(recipes, r)
	})
	for _, r := range recipePagesList {
		c.Visit(r)
	}

	// c.Visit("https://en.wikibooks.org/wiki/Cookbook:Xat%C3%B3_(Catalan_Endives_and_Fish)")

	recipesData, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	os.WriteFile("./recipes.json", recipesData, os.FileMode(0644))
}

func findNextPageURL(e *colly.HTMLElement) (string, bool) {
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

	return "https://en.wikibooks.org" + uri, true
}

func getRecipesLinkList(e *colly.HTMLElement) []string {
	list := []string{}
	e.ForEach("#mw-pages > div.mw-content-ltr > div > div > ul > li > a", func(i int, h *colly.HTMLElement) {
		uri := h.Attr("href")
		if uri == "" {
			return
		}
		url := "https://en.wikibooks.org" + uri
		if ignoreList[url] {
			return
		}
		list = append(list, url)
	})

	return list
}

func scrapRecipeData(link string, e *colly.HTMLElement) (recipe, bool) {
	r := recipe{Link: link}

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
			r.Ingredients = append(r.Ingredients, parseIngredient(i))
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
		return recipe{}, false
	}

	return r, true
}

func parseIngredient(input string) ingredient {
	input = strings.Replace(input, "⅛", "0.125", 1)
	input = strings.Replace(input, "¼", "0.25", 1)
	input = strings.Replace(input, "⅓", "0.33", 1)
	input = strings.Replace(input, "½", "0.5", 1)
	input = strings.Replace(input, "⅔", "0.66", 1)
	input = strings.Replace(input, "¾", "0.75", 1)

	ing, isOk := parseIngredientWithAmountUnitName(input)
	if isOk {
		return ing
	}

	ing, isOk = parseIngredientWithAmountName(input)
	if isOk {
		return ing
	}

	return ingredient{
		Name: input,
	}
}

var regexIngredientWithAmountUnitName = regexp.MustCompile(`(\d*\.?\d+)\D*(cup|lbs|tablespoon|tbsp|teaspoon|oz|g|gram)[s).,]?\s+(.+)`)

func parseIngredientWithAmountUnitName(input string) (ingredient, bool) {
	matches := regexIngredientWithAmountUnitName.FindStringSubmatch(input)
	if len(matches) < 4 {
		return ingredient{}, false
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	return ingredient{
		Name:   matches[3],
		Amount: amount,
		Unit:   matches[2],
	}, true
}

var regexIngredientWithAmountName = regexp.MustCompile(`(\d*\.?\d+)\s(.+)`)

func parseIngredientWithAmountName(input string) (ingredient, bool) {
	matches := regexIngredientWithAmountName.FindStringSubmatch(input)
	if len(matches) < 3 {
		return ingredient{}, false
	}

	amount, _ := strconv.ParseFloat(matches[1], 64)
	return ingredient{
		Name:   matches[2],
		Amount: amount,
	}, true
}

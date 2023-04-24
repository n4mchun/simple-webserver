package main

import (
	"os"
	"strings"

	"./scrapper"
	"github.com/labstack/echo"
)

func handleIndex(c echo.Context) error {
	return c.File("templates/index.html")
}

func handleScrape(c echo.Context) error {
	searchWord := strings.ToLower(c.FormValue("searchWord"))
	filename := searchWord + ".csv"
	defer os.Remove(filename)

	scrapper.Scrape(searchWord)
	return c.Attachment(filename, filename)
}

func main() {
	e := echo.New()
	e.GET("/", handleIndex)
	e.POST("/scrape", handleScrape)

	e.Logger.Fatal(e.Start(":1323"))
}

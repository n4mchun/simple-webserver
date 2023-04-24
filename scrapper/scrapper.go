package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type jobInfo struct {
	title     string
	link      string
	date      string
	location  string
	career    string
	education string
	job_type  string
}

var cnt int = 0

func Scrape(searchWord string) {
	baseURL := "https://www.saramin.co.kr/zf_user/search?searchType=search&cat_mcls=2&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&keydownAccess=&panel_type=&search_optional_item=y&search_done=y&panel_count=y&preview=y&recruitSort=relation&recruitPageCount=40&inner_com_type=&show_applied=&quick_apply=&except_read=&ai_head_hunting=&mainSearch=n&searchword=" + searchWord
	c := make(chan []jobInfo)

	pageCount := getPageCount(baseURL)
	for page := 1; page <= pageCount; page++ {
		go extractJobs(baseURL, page, c)
	}

	results := [][]jobInfo{}
	for i := 0; i < pageCount; i++ {
		results = append(results, <-c)
	}
	fmt.Println("Done,", searchWord, "extracted", cnt)

	writeJobs(searchWord, pageCount, results)
}

func writeJobs(searchWord string, pageCount int, jobs [][]jobInfo) {
	file, err := os.Create(searchWord + ".csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Link", "Date", "Location", "Career", "Education", "Type"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, page := range jobs {
		for _, job := range page {
			jobSlice := []string{job.title, job.link, job.date, job.location, job.career, job.education, job.job_type}
			jwErr := w.Write(jobSlice)
			checkErr(jwErr)
		}
	}
}

func extractJobs(baseURL string, page int, c chan []jobInfo) {
	recruitPage := baseURL + "&recruitPage=" + strconv.Itoa(page)
	res, err := http.Get(recruitPage)
	checkStatusCode(res)
	checkErr(err)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	result := []jobInfo{}
	jobsChan := make(chan jobInfo)

	s := doc.Find(".item_recruit")
	for i := 0; i < s.Length(); i++ {
		go extractJobInfo(s.Eq(i), jobsChan)
		cnt += 1
	}
	for i := 0; i < s.Length(); i++ {
		job := <-jobsChan
		result = append(result, job)
	}

	c <- result
}

func extractJobInfo(card *goquery.Selection, c chan<- jobInfo) {
	title, _ := card.Find(".job_tit a").Attr("title")
	link, _ := card.Find(".job_tit a").Attr("href")
	date := card.Find(".job_date span").Text()
	location := card.Find(".job_condition").Eq(0).Find("a").Text()
	career := card.Find(".job_condition span").Eq(1).Text()
	education := card.Find(".job_condition span").Eq(2).Text()
	job_type := card.Find(".job_condition span").Eq(3).Text()

	c <- jobInfo{
		title:     title,
		link:      "https://www.saramin.co.kr" + link,
		date:      date,
		location:  location,
		career:    career,
		education: education,
		job_type:  job_type,
	}
}

func getPageCount(baseURL string) int {
	lastPage := 1
	for recruitPage := 1; ; recruitPage += 10 {
		searchURL := baseURL + "&recruitPage=" + strconv.Itoa(recruitPage)
		res, err := http.Get(searchURL)
		checkStatusCode(res)
		checkErr(err)

		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		checkErr(err)

		nextExist := doc.Find(".btnNext")
		if ret, _ := nextExist.Html(); ret == "" {
			doc.Find(".pagination span").Each(func(i int, s *goquery.Selection) {
				page, _ := s.Html()
				lastPage, _ = strconv.Atoi(page)
			})
			return lastPage
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

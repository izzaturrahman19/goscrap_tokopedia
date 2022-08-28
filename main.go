package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

type Product struct {
	Name        string
	Description string
	Price       string
	Merchant    string
	Rating      string
	ImageLink   string
}

func getHtml(url string) *http.Response {
	// Create HTTP client respone with user agent
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func writeCSV(data []string) {
	fileName := "tokopedia.csv"
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	err = writer.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Write to CSV success")
}

func scrapeChildPageData(doc *goquery.Document) Product {
	product := Product{
		Name:        doc.Find("h1[data-testid='lblPDPDetailProductName']").Text(),
		Description: doc.Find("div[data-testid='lblPDPDescriptionProduk']").Text(),
		Price:       doc.Find("div[data-testid='lblPDPDetailProductPrice']").Text(),
		Merchant:    doc.Find("div[id='pdp_comp-shop_credibility']>div.css-d1nhq9>div>div>a[data-testid='llbPDPFooterShopName']>h2").Text(),
		Rating:      doc.Find("div#pdp_comp-review>div>div>div>div:first-child>div:first-child>p.score-info>span.score").Text(),
		ImageLink:   doc.Find("div.intrinsic>span>img").AttrOr("src", "no image"),
	}
	return product
}

func scrapePageData(doc *goquery.Document, lastCount int) (count int) {
	count = lastCount
	doc.Find("div[data-testid='lstCL2ProductList']>div").Each(func(i int, s *goquery.Selection) {
		resp := getHtml(s.Find("a").AttrOr("href", "no url"))
		defer resp.Body.Close()
		childDoc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		product := scrapeChildPageData(childDoc)
		fmt.Println(product)

		// write to csv
		header := []string{"Name", "Description", "Price", "Merchant", "Rating", "ImageLink"}
		if count == 0 {
			writeCSV(append(header, []string{product.Name, product.Description, product.Price, product.Merchant, product.Rating, product.ImageLink}...))
		} else {
			writeCSV([]string{product.Name, product.Description, product.Price, product.Merchant, product.Rating, product.ImageLink})
		}
		count++
	})

	return count
}

func main() {
	url := "https://www.tokopedia.com/p/handphone-tablet/handphone"
	pageNumber := 1
	count := 0
	for {
		resp := getHtml(url)
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		count = scrapePageData(doc, count)
		if count == 100 {
			break
		}
		pageNumber++
		url = "https://www.tokopedia.com/p/handphone-tablet/handphone" + "?page=" + fmt.Sprint(pageNumber)
	}
}

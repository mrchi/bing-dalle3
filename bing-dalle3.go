package bingdalle3

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BingDalle3 struct {
	cookie string
}

func (bing *BingDalle3) GetTokenBalance() (int, error) {
	url := "https://www.bing.com/images/create"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cookie", bing.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")

	client := http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("resp status error: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(doc.Find("div#token_bal").Text())
	if err != nil {
		return 0, err
	}
	return value, nil
}

func NewBingDalle3(cookie string) *BingDalle3 {
	return &BingDalle3{cookie: cookie}
}

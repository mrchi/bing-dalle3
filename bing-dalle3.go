package bingdalle3

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
	Timeout   = 5 * time.Second
	PageUrl   = "https://www.bing.com/images/create"
	ResultUrl = "https://www.bing.com/images/create/async/results/"
)

type BingDalle3 struct {
	cookie string
}

func (bing *BingDalle3) genUrlForCreatingImage(prompt string) (string, error) {
	fullUrl, err := url.Parse(PageUrl)
	if err != nil {
		return "", err
	}
	queryParams := url.Values{}
	queryParams.Add("q", prompt)
	queryParams.Add("rt", "4")
	queryParams.Add("FORM", "GENCRE")
	fullUrl.RawQuery = queryParams.Encode()
	return fullUrl.String(), nil
}

func (bing *BingDalle3) genUrlForQueryingResult(id string, prompt string) (string, error) {
	fullUrlString, err := url.JoinPath(ResultUrl, id)
	if err != nil {
		return "", err
	}
	fullUrl, err := url.Parse(fullUrlString)
	if err != nil {
		return "", err
	}
	queryParams := url.Values{}
	queryParams.Add("q", prompt)
	fullUrl.RawQuery = queryParams.Encode()
	return fullUrl.String(), nil
}

func (bing *BingDalle3) genUrlForQueryingResultReferer(id string, prompt string) (string, error) {
	fullUrl, err := url.Parse(PageUrl)
	if err != nil {
		return "", err
	}
	queryParams := url.Values{}
	queryParams.Add("q", prompt)
	queryParams.Add("rt", "4")
	queryParams.Add("FORM", "GENCRE")
	queryParams.Add("id", id)
	fullUrl.RawQuery = queryParams.Encode()
	return fullUrl.String(), nil
}

func (bing *BingDalle3) GetTokenBalance() (int, error) {
	req, err := http.NewRequest(http.MethodGet, PageUrl, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Cookie", bing.cookie)
	req.Header.Set("User-Agent", UserAgent)

	client := http.Client{Timeout: Timeout}
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

func (bing *BingDalle3) CreateImage(prompt string) (string, error) {
	fullUrl, err := bing.genUrlForCreatingImage(prompt)
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("q", prompt)
	data.Set("qs", "ds")

	req, err := http.NewRequest(http.MethodPost, fullUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Cookie", bing.cookie)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", fullUrl)

	client := http.Client{
		Timeout: Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		errMsg := doc.Find("div.gil_err_sbt").Text()
		return "", errors.New(errMsg)
	} else if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("resp status error: %s", resp.Status)
	}

	redirectUrl, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		return "", err
	}
	id := redirectUrl.Query().Get("id")
	if id == "" {
		return "", fmt.Errorf("ID not found in redirect URL: %s", redirectUrl.String())
	}
	return id, nil
}

func (bing *BingDalle3) QueryResult(id string, prompt string) ([]string, error) {
	reqUrl, err := bing.genUrlForQueryingResult(id, prompt)
	if err != nil {
		return nil, err
	}
	referUrl, err := bing.genUrlForQueryingResultReferer(id, prompt)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", bing.cookie)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", referUrl)

	client := http.Client{Timeout: Timeout}

	timeoutChan := time.After(10 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	var urls []string
	for {
		select {
		case <-timeoutChan:
			return nil, errors.New("timeout")
		case <-ticker.C:
			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("resp status error: %s", resp.Status)
			}
			doc.Find("img.mimg").Each(func(i int, selection *goquery.Selection) {
				url, _ := selection.Attr("src")
				urls = append(urls, removeQueryParamsForUrl(url))
			})
			if len(urls) > 0 {
				return urls, nil
			}
		}
	}
}

func NewBingDalle3(cookie string) *BingDalle3 {
	return &BingDalle3{cookie: cookie}
}

func removeQueryParamsForUrl(fullUrl string) string {
	url, err := url.Parse(fullUrl)
	if err != nil {
		return fullUrl
	}
	url.RawQuery = ""
	return url.String()
}

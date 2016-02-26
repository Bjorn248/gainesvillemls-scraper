package main

import (
	"bytes"
	"fmt"
	"github.com/Bjorn248/gainesvillemls-scraper/Godeps/_workspace/src/golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

func main() {
	MLSNumbers := getMLSNumbers()
	MLSURLs := getMLSDetails(MLSNumbers)
	fmt.Println(MLSURLs)
}

func getMLSNumbers() []string {
	MLSNums := []string{}

	searchURL := "http://www.gainesvillemls.com"
	searchPath := "/gan/idx/search.php"

	data := url.Values{}
	data.Set("LM_MST_prop_fmtYNNT", "1")
	data.Add("LM_MST_prop_cdYYNT", "1,9,10,11,12,13,14")
	data.Add("LM_MST_mls_noYYNT", "")
	// Minimum Price
	data.Add("LM_MST_list_prcYNNB", "60000")
	// Maximum Price
	data.Add("LM_MST_list_prcYNNE", "150000")
	data.Add("LM_MST_prop_cdYNNL[]", "9")
	// Minimum Square Footage
	data.Add("LM_MST_sqft_nYNNB", "")
	// Maximum Square Footage
	data.Add("LM_MST_sqft_nYNNE", "")
	// Minimum Year Built
	data.Add("LM_MST_yr_bltYNNB", "")
	// Maximum Year Built
	data.Add("LM_MST_yr_bltYNNE", "")
	// Minimum Bedrooms
	data.Add("LM_MST_bdrmsYNNB", "3")
	// Maximum Bedrooms
	data.Add("LM_MST_bdrmsYNNE", "")
	// Minimum Bathrooms
	data.Add("LM_MST_bathsYNNB", "2")
	// Maximum Bathrooms
	data.Add("LM_MST_bathsYNNE", "")
	data.Add("LM_MST_hbathYNNB", "")
	data.Add("LM_MST_hbathYNNE", "")
	// County
	data.Add("LM_MST_countyYNCL[]", "ALA")
	data.Add("LM_MST_str_noY1CS", "")
	data.Add("LM_MST_str_namY1VZ", "")
	data.Add("LM_MST_remarksY1VZ", "")
	data.Add("openHouseStartDt_B", "")
	data.Add("openHouseStartDt_E", "")
	data.Add("ve_info", "")
	data.Add("ve_rgns", "1")
	data.Add("LM_MST_LATXX6I", "")
	data.Add("poi", "")
	data.Add("count", "1")
	data.Add("key", "52633f4973cf845e55b18c8e22ab08d5")
	data.Add("isLink", "0")
	data.Add("custom", "")

	u, _ := url.ParseRequestURI(searchURL)
	u.Path = searchPath
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	request, requestErr := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	if requestErr != nil {
		log.Fatalf("Problem creating new httpRequest", "%s", requestErr)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, responseError := client.Do(request)
	if responseError != nil {
		log.Fatalf("Problem creating new httpRequest", "%s", responseError)
	}

	responseBody := resp.Body
	defer responseBody.Close()

	parsedHTML := html.NewTokenizer(responseBody)

	MLSNumber := ""
	MLSFlag := false

	for {
		tt := parsedHTML.Next()
		switch {
		case tt == html.ErrorToken:
			return MLSNums
		case tt == html.StartTagToken:
			t := parsedHTML.Token()
			if t.String() == `<span class="mls">` {
				MLSFlag = true
			}
		case tt == html.TextToken:
			t := parsedHTML.Token().String()
			tLower := strings.ToLower(t)
			if strings.Contains(tLower, ", fl") {
				// We don't want homes not in Gainesville
				if !strings.Contains(tLower, "gainesville") {
					// Remove the latest MLS number from the array
					MLSNums = MLSNums[:len(MLSNums)-1]
					break
				}
			}
			if MLSFlag == true {
				MLSNumber = t
				MLSNums = append(MLSNums, MLSNumber)
				MLSFlag = false
			}
		}
	}
}

func getMLSDetails(mlsArray []string) []string {
	MLSURLs := []string{}
	// chans := make([]chan string, len(mlsArray))
	responses := make(chan string)
	var wg sync.WaitGroup

	for _, mlsNumber := range mlsArray {
		wg.Add(1)
		go func(mlsNumber string) {
			defer wg.Done()
			responses <- getMLSDetail(mlsNumber)
		}(mlsNumber)
	}

	go func() {
		for response := range responses {
			// We only want responses that are not empty strings
			if response != "" {
				MLSURLs = append(MLSURLs, response)
			}
		}
	}()
	wg.Wait()
	return MLSURLs
}

func getMLSDetail(MLSNumber string) string {
	MLSURL := ""

	searchURL := "http://www.gainesvillemls.com"
	searchPath := "/gan/idx/detail.php"
	data := url.Values{}
	data.Set("key", "52633f4973cf845e55b18c8e22ab08d5")
	data.Add("gallery", "false")
	data.Add("custom", "")
	data.Add("mls", MLSNumber)
	u, _ := url.ParseRequestURI(searchURL)
	u.Path = searchPath
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	request, requestErr := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	if requestErr != nil {
		log.Fatalf("Problem creating new httpRequest", "%s", requestErr)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, responseError := client.Do(request)
	if responseError != nil {
		log.Fatalf("Problem creating new httpRequest", "%s", responseError)
	}

	responseBody := resp.Body
	defer responseBody.Close()

	parsedHTML := html.NewTokenizer(responseBody)

	constructionFlag := false
	constructionCounter := 2
	parkingFlag := false
	parkingCounter := 2

	for {
		tt := parsedHTML.Next()
		switch {
		case tt == html.ErrorToken:
			return MLSURL
		case tt == html.TextToken:
			t := parsedHTML.Token()
			if constructionFlag == true {
				constructionCounter--
				if constructionCounter == 0 {
					if strings.Contains(strings.ToLower(t.String()), "block") {
						MLSURL = fmt.Sprintf("http://www.gainesvillemls.com/gan/idx/index.php?key=52633f4973cf845e55b18c8e22ab08d5&mls=%s\n", MLSNumber)
					}
					constructionFlag = false
				}
			}
			if parkingFlag == true {
				parkingCounter--
				if parkingCounter == 0 {
					if strings.Contains(strings.ToLower(t.String()), "no garage") {
						return ""
					}
					parkingFlag = false
				}
			}
			if t.String() == "Construction-exterior:" {
				constructionFlag = true
			}
			if t.String() == "Parking:" {
				parkingFlag = true
			}

		}
	}
}

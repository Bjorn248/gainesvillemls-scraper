package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/BjornTwitchBot/gainesvillemls-scraper/Godeps/_workspace/src/github.com/garyburd/redigo/redis"
	"github.com/BjornTwitchBot/gainesvillemls-scraper/Godeps/_workspace/src/github.com/robfig/cron"
	"github.com/BjornTwitchBot/gainesvillemls-scraper/Godeps/_workspace/src/golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

var cronScheduler *cron.Cron

// We just want to keep the program running forever
// So that the scheduler can keep running the search function
var waitForever sync.WaitGroup

func main() {
	if os.Getenv("REDIS_HOST_PORT") == "" {
		log.Fatal("REDIS_HOST_PORT not set")
	}
	if os.Getenv("REDIS_PASSWORD") == "" {
		log.Fatal("REDIS_PASSWORD not set")
	}
	if os.Getenv("SENDGRID_API_TOKEN") == "" {
		log.Fatal("SENDGRID_API_TOKEN not set")
	}
	if os.Getenv("EMAIL_FROM_ADDRESS") == "" {
		log.Fatal("EMAIL_FROM_ADDRESS not set")
	}
	if os.Getenv("EMAIL_TO_ADDRESS") == "" {
		log.Fatal("EMAIL_TO_ADDRESS not set")
	}

	flag.Parse()
	// Instantiate redis connection pool
	pool = newPool(*redisServer, *redisPassword)
	poolErr := pool.Get().Err()
	// Check redis connection
	if poolErr != nil {
		log.Fatalf("Something went wrong connecting to Redis! Error is '%s'", poolErr)
	}

	// Instantiate Cron Scheduler
	cronScheduler = cron.New()

	waitForever.Add(1)

	cronScheduler.AddFunc("@hourly", func() {
		// waitForever is now 2, so it will cycle between 2 and 1
		// never reaching 0
		// reaching 0 would cause the program to exit
		waitForever.Add(1)
		MLSPrices := getMLSPrices()
		populateListings(MLSPrices)
		MLSNumbers := returnMLSNumbers(MLSPrices)
		MLSURLs := getMLSDetails(MLSNumbers)
		fmt.Println(MLSURLs)
		if len(MLSURLs) > 0 {
			sendEmail(os.Getenv("EMAIL_TO_ADDRESS"), MLSURLs)
		}
		waitForever.Done()
	})
	cronScheduler.Start()
	waitForever.Wait()
}

func returnMLSNumbers(MLSNumberPrices []string) []string {
	MLSNumbers := []string{}
	for _, MLSPrice := range MLSNumberPrices {
		MLSNumbers = append(MLSNumbers, strings.Split(MLSPrice, "_")[0])
	}
	return MLSNumbers
}

func filterOldListings(listings []string) []string {
	redisConn := pool.Get()
	defer redisConn.Close()

	filteredListings := []string{}

	for _, listing := range listings {
		redisReply, redisError := redis.Bool(redisConn.Do("EXISTS", listing))
		if redisError != nil {
			log.Fatalf("Error reading redis data '%s'", redisError)
		}
		if redisReply == false {
			filteredListings = append(filteredListings, listing)
		}
	}
	return filteredListings
}

func populateListings(listings []string) {
	redisConn := pool.Get()
	defer redisConn.Close()

	for _, listing := range listings {
		_, redisError := redisConn.Do("SET", listing, "true")
		if redisError != nil {
			log.Fatalf("Error inserting data into redis '%s'", redisError)
		}
	}
}

func getMLSPrices() []string {
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
	Price := ""
	PriceFlag := false
	cityFlag := false

	for {
		tt := parsedHTML.Next()
		switch {
		case tt == html.ErrorToken:
			return filterOldListings(MLSNums)
		case tt == html.StartTagToken:
			t := parsedHTML.Token()
			if t.String() == `<span class="mls">` {
				MLSFlag = true
			}
			if t.String() == `<span class="price">` {
				PriceFlag = true
			}
		case tt == html.TextToken:
			t := parsedHTML.Token().String()
			tLower := strings.ToLower(t)
			if strings.Contains(tLower, ", fl") {
				// We don't want homes not in Gainesville
				if strings.Contains(tLower, "gainesville") {
					cityFlag = true
				}
			}
			if MLSFlag == true {
				MLSNumber = t
				MLSFlag = false
			}
			if PriceFlag == true {
				Price = strings.TrimSpace(t)
				MLS_Price := MLSNumber + "_" + Price
				if cityFlag == true {
					MLSNums = append(MLSNums, MLS_Price)
					cityFlag = false
				}
				PriceFlag = false
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

// Google Image Search functionality
package gis

import (
	"fmt"
	"github.com/sdstrowes/gesture/core"
	"github.com/sdstrowes/gesture/util"
	"math/rand"
	neturl "net/url"
	"strings"
	"time"
)

const developerKey = "nope"
const customcx     = "nope"

func Create(bot *core.Gobot, config map[string]interface{}) {
	defaultUrl, useDefault := config["default"].(string)
	exclusions := getExclusions(config)
	bot.ListenFor("^gis (.*)", func(msg core.Message, matches []string) core.Response {
		for _, ex := range(exclusions) {
			if ex == msg.Channel {
				return bot.Stop()
			}
		}
		link, err := search(matches[1])
		if err != nil {
			if useDefault {
				link = defaultUrl
			} else {
				return bot.Error(err)
			}
		}
		msg.Ftfy(link)
		return bot.Stop()
	})
	bot.ListenFor("^gie's (.*)", func(msg core.Message, matches []string) core.Response {
		for _, ex := range(exclusions) {
			if ex == msg.Channel {
				return bot.Stop()
			}
		}
		link, err := search("gie's "+matches[1])
		if err != nil {
			if useDefault {
				link = defaultUrl
			} else {
				return bot.Error(err)
			}
		}
		msg.Ftfy(link)
		return bot.Stop()
	})
}

func getExclusions(config map[string]interface{}) []string {
	result := make([]string, 0)
	exclude, ok := config["exclude"].([]interface{})
	if (!ok) {
		return result
	}
	for _, ex := range(exclude) {
		result = append(result, ex.(string))
	}
	return result
}

type gisItems struct {
	Link string
}

type gisResponse struct {
	Items *[]gisItems
}

// Search queries google for some images, and then randomly selects one
func search(search string) (string, error) {
	searchUrl := "https://www.googleapis.com/customsearch/v1?cx="+customcx+"&key="+developerKey+"&searchType=image&q=" + neturl.QueryEscape(search)
	var gisResponse gisResponse

	// NB: i am a terrible programmer
	for i := 0; i < 5; i++ {
		err := util.UnmarshalUrl(searchUrl, &gisResponse)


		if err == nil {
			break
		}
		if i == 4 {
			return "", err
		}
	}

	if gisResponse.Items == nil {
		return "", fmt.Errorf("No results were returned for query %s", search)
	}

	results := *gisResponse.Items

	if len(results) > 0 {

		// start a goroutine to determine image info for each response result
		// we have to use buffered channels so that the senders don't hang on send after the main method exits
		imageUrlCh := make(chan string, len(results))
		errorsCh := make(chan error, len(results))
		for _, resultUrl := range results {
			go getImageInfo(resultUrl.Link, imageUrlCh, errorsCh)
		}

		// until a timeout is met, build a collection of urls
		totalResults := len(results)
		remainingResults := totalResults
		urls := make([]string, 0, totalResults)
		errors := make([]error, 0, totalResults)
		timeout := time.After(2 * time.Second)

	SEARCH:
		for remainingResults > 0 {
			select {
			case url := <-imageUrlCh:
				urls = append(urls, url)
				remainingResults--
			case err := <-errorsCh:
				errors = append(errors, err)
				remainingResults--
			case <-timeout:
				break SEARCH
			}
		}
		if len(urls) == 0 {
			return "", fmt.Errorf("No image could be found for \"%s\"", search)
		}
		return urls[rand.Intn(len(urls))], nil

	}
	return "", fmt.Errorf("No image could be found for \"%s\"", search)
}

// getImageInfo looks at the header info for the url, and if it is an image, it sends an imageInfo on the channel
func getImageInfo(url string, ch chan<- string, failures chan<- error) {
	imageUrl, contentType, err := util.ResponseHeaderContentType(url)
	if err == nil && strings.HasPrefix(contentType, "image/") {
		url, err := ensureSuffix(imageUrl, "."+contentType[len("image/"):])
		if err != nil {
			failures <- err
		} else {
			ch <- url
		}
	} else {
		failures <- fmt.Errorf("Not an image: %s", url)
	}
}

// ensureSuffix ensures a url ends with suffixes like .jpg, .png, etc
func ensureSuffix(url, suffix string) (string, error) {
	var err error
	unescapedUrl, err := neturl.QueryUnescape(url)
	if err != nil {
		return "", err
	}
	lowerSuffix := strings.ToLower(suffix)
	lowerUrl := strings.ToLower(unescapedUrl)
	if lowerSuffix == ".jpeg" && strings.HasSuffix(lowerUrl, ".jpg") {
		return url, nil
	}
	if lowerSuffix == ".jpg" && strings.HasSuffix(lowerUrl, ".jpeg") {
		return url, nil
	}
	if strings.HasSuffix(lowerUrl, lowerSuffix) {
		return url, nil
	}
	return url, nil
}

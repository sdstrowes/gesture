// A Gesture interface to various YouTubery
package youtube

import (
	"flag"
	"net/http"
	"errors"
	"fmt"
	"github.com/sdstrowes/gesture/core"
	"log"
	"math/rand"
	"net/url"
	"regexp"
	"github.com/google/google-api-go-client/googleapi/transport"
	"github.com/google/google-api-go-client/youtube/v3"
)

// A YouTube plugin

const developerKey = "AIzaSyD2XM3TlPT17JTptv4dP3F31o-bEa3wO78"

var urlCleaner = regexp.MustCompile(`&feature=youtube_gdata_player`)

func Create(bot *core.Gobot, config map[string]interface{}) {
	results, ok := config["results"].(float64)
	if !ok {
		log.Print("Failed to load config for 'youtube' plugin. Using default result count of 1")
		results = 1
	}

	bot.ListenFor("^yt (.*)", func(msg core.Message, matches []string) core.Response {
		link, err := search(matches[1], int(results))
		if err != nil {
			return bot.Error(err)
		}
		if link != "" {
			msg.Ftfy(link)
		}
		return bot.Stop()
	})
}




// Search youtube for the given query string. Returns one of the first N youtube
// results for that search at random (everyone loves entropy!)
// Returns an empty string if there were no results for that query
func search(q string, results int) (link string, err error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	var query      = flag.String("query", q, "Search term")
	var maxResults = flag.Int64("max-results", results, "Max YouTube results")

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").Q(*query).MaxResults(*maxResults)
	response, err := call.Do()

	if err != nil {
		err = errors.New("No video found for search \"" + query + "\"")
		return
	}

	videos := response.Items
	switch l := len(videos); {
	case l > 1:
		ordering := rand.Perm(len(videos))

		for _, i := range ordering {
fmt.Printf("i: %d, id: %s\n", i, response.Items[i].VideoId)

			// Youtube adds a fragment to the end of players accessed via the API. Get
			// rid of that shit.
			link = urlCleaner.ReplaceAllLiteralString(videos[i].Player.Default, "")
		}
	case l == 1:
		link = urlCleaner.ReplaceAllLiteralString(videos[0].Player.Default, "")
	}


	// Iterate through each item and add it to the correct list.
//	for _, item := range response.Items {
//		switch item.Id.Kind {
//		case "youtube#video":
//			videos[item.Id.VideoId] = item.Snippet.Title
//		}
//	}

	return
}

// Generate a search URL for the given query. Returns the requested number of
// search results.
func buildSearchUrl(query string, results int) string {
	escapedQuery := url.QueryEscape(query)
	searchString := "https://gdata.youtube.com/feeds/api/videos?q=%v&max-results=%d&v=2&alt=jsonc"
	return fmt.Sprintf(searchString, escapedQuery, results)
}

// YouTube response types for deserializing JSON
type youTubePlayer struct {
	Default string
	Mobile  string
}

type youTubeItem struct {
	Title       string
	Description string
	Player      youTubePlayer
}

type youTubeData struct {
	Items []youTubeItem
}

type youTubeResponse struct {
	ApiVersion string
	Data       youTubeData
}




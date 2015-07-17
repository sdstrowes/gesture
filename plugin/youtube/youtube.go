// A Gesture interface to various YouTubery
package youtube

import (
	"net/http"
	"errors"
	"github.com/sdstrowes/gesture/core"
	"log"
	"math/rand"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

const developerKey = "nope"

func Create(bot *core.Gobot, config map[string]interface{}) {
	results, ok := config["results"].(float64)
	if !ok {
		log.Print("Failed to load config for 'youtube' plugin. Using default result count of 1")
		results = 1
	}

	bot.ListenFor("^yt (.*)", func(msg core.Message, matches []string) core.Response {
		link, err := search(matches[1], int64(results))
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
func search(q string, results int64) (link string, err error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	call := service.Search.List("id,snippet").Q(q).MaxResults(results)
	response, err := call.Do()
	if err != nil {
		return
	}

	switch l := len(response.Items); {
	case l == 0:
		err = errors.New("No results for \"" + q + "\"")
	case l > 0:
		var n = rand.Intn(len(response.Items))
		link = "\""+response.Items[n].Snippet.Title+"\": https://www.youtube.com/watch?v="+response.Items[n].Id.VideoId;
	}

	return
}

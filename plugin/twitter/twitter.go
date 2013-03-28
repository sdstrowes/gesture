// does twitter related things
package twitter

import (
	"encoding/json"
	"fmt"
	"gesture/plugin"
	"gesture/util"
	"regexp"
	"strings"
)

var (
	commandRegex = regexp.MustCompile(`^https?://(www.)?twitter.com/.*?/status/.*$`)
)

// lol types. we don't need to keep state so i guess we'll just use bool
type TwitterPlugin bool

// lol types
func NewPlugin() TwitterPlugin {
	return TwitterPlugin(false)
}

func (twitter TwitterPlugin) Call(mc plugin.MessageContext) (bool, error) {
	if mc.Command() == "describe" {
		if len(mc.CommandArgs()) > 0 {
			described, err := describe((mc.CommandArgs())[0])
			if err != nil {
				return false, err
			} else {
				mc.Ftfy(described)
			}
		}
		return true, nil
	} else {
		found := false
		for _, word := range strings.Split(mc.Message(), " ") {
			if strings.Contains(word, "twitter.com") {
				if tweet, err := getTweet(word); err == nil && tweet != "" {
					found = true
					mc.Send(tweet)
				}
			}
		}
		if found {
			return true, nil
		}
	}
	return false, nil
}

func getTweet(url string) (result string, err error) {
	parts := strings.Split(url, "/")
	if len(parts) < 1 {
		return "", nil
	}
	tweetId := parts[len(parts)-1]
	if tweetId == "" {
		return "", nil
	}
	body, err := util.GetUrl("https://api.twitter.com/1/statuses/show/" + tweetId + ".json")
	if err != nil {
		return "", err
	}
	var content map[string]interface{}
	if err = json.Unmarshal(body, &content); err != nil {
		return "", err
	}
	response := fmt.Sprintf("%s: %s", parts[3], content["text"].(string))
	return response, nil
}

func describe(user string) (result string, err error) {
	body, err := util.GetUrl("http://api.twitter.com/1/users/lookup.json?include_entities=true&screen_name=" + user)
	if err != nil {
		return "", err
	}
	var jsonResponse []map[string]interface{}
	if err = json.Unmarshal(body, &jsonResponse); err != nil {
		return "", err
	}
	first := jsonResponse[0]
	description := first["description"].(string)
	pic := first["profile_image_url_https"].(string)
	return fmt.Sprintf("\"%s\" %s", description, pic), nil
}

// GetStatus queries a status url and outputs the rewritten text
func GetStatus(url string) (result string, err error) {
	if !commandRegex.MatchString(url) {
		return "", nil
	}
	pieces := strings.Split(url, "/")
	id := pieces[len(pieces)-1]
	body, err := util.GetUrl("https://api.twitter.com/1/statuses/show/" + id + ".json")
	if err != nil {
		return "", err
	}
	var jsonResponse map[string]interface{}
	json.Unmarshal(body, &jsonResponse)
	result = jsonResponse["text"].(string)
	return
}

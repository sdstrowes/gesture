/*
 The markov plugin listens passively to all messages and creates markov
 chains for each nick.
*/
package markov

import (
	"fmt"
	"github.com/sdstrowes/gesture/core"
	"github.com/sdstrowes/gesture/state"
	"log"
	"math/rand"
	"strings"
	"strconv"
	"sync"
	"time"
)

type markovState struct {
	PrefixLength int
	Chains       map[string]map[string][]string  // map[user]map[prefix][]chains
	SuffixCounts map[string]map[string]map[string]int // map[user]map[prefix]map[suffix]
}

const (
	maxWords       = 20
	maxChainLength = 1000
)

var (
	// todo: make prefix length configurable
	markov      = markovState{PrefixLength: 1, Chains: make(map[string]map[string][]string), SuffixCounts: make(map[string]map[string]map[string]int)}
	//markov      = markovState{PrefixLength: 1, Chains: make(map[string]map[string][]string)}
	mutex       sync.Mutex
	lastInduction	time.Time
	pluginState = state.NewState("markov")
)

var induction = [...]string {"1. Welcome to ##gucs. This is an *un*official GU CS channel. Banter and realtalk are welcome.",
                             "2. Please report your dealios to knifa on a regular basis.",
                             "3. Please familiarise yourself with the channel anthem: https://www.youtube.com/watch?v=fA2AiedVCmw",
                             "4. Play nice with the bot."}

var insult_a = [...]string {"a lazy", "a stupid", "an insecure", "an idiotic", "a slimy", "a slutty", "a smelly", "a pompous", "a communist", "an elitist"}
var insult_b = [...]string {"douche", "ass", "turd", "rectum", "butt", "cock", "shit", "crotch", "prick", "boner", "dick"}
var insult_c = [...]string {"pilot", "canoe", "captain", "pirate", "hammer", "jockey", "waffle", "goblin", "biscuit", "clown", "monster", "hound", "dragon", "balloon"}

func Create(bot *core.Gobot, config map[string]interface{}) {
	if err := pluginState.Load(&markov); err != nil {
		log.Printf("Could not load plugin state: %s", err)
	}

// ONE OFF
//	//Chains       map[string]map[string][]string  // map[user]map[prefix][]chains
//	//SuffixCounts map[string]map[string]map[string]int // map[user]map[prefix]map[suffix]
//
	for user := range markov.Chains {
		fmt.Printf("USER: %s\n", user);
		/* create the user if they don't exist: counts[sds] */
		userMap, ok := markov.SuffixCounts[user]
		if !ok {
			userMap := make(map[string]map[string]int)
			markov.SuffixCounts[user] = userMap
		}

		/* create the first word: counts[sds][foo] */
		for word := range markov.Chains[user] {
			weightedWords, ok := userMap[word]
			if !ok {
				weightedWords := make(map[string]int)
				markov.SuffixCounts[user][word] = weightedWords
			}

			for i := 0; i < len(markov.Chains[user][word]); i++ {
				nextword := markov.Chains[user][word][i]
				_, ok := weightedWords[nextword]
				if !ok {
					markov.SuffixCounts[user][word][nextword] = 1
				}
			}
		}
	}
// ONE OFF

	rand.Seed(time.Now().UTC().UnixNano())

	bot.ListenFor("^ *!insult *$", func(msg core.Message, matches []string) core.Response {
		user := msg.User

		msg.Send(user+", you're nothing but "+insult_a[rand.Intn(len(insult_a))]+" "+insult_b[rand.Intn(len(insult_b))]+"-"+insult_c[rand.Intn(len(insult_c))])
		return bot.Stop()
	})

	bot.ListenFor("^ *!insult +(.+)*$", func(msg core.Message, matches []string) core.Response {
		msg.Send(matches[1]+", you're nothing but "+insult_a[rand.Intn(len(insult_a))]+" "+insult_b[rand.Intn(len(insult_b))]+"-"+insult_c[rand.Intn(len(insult_c))])
		return bot.Stop()
	})

	bot.ListenFor("^ *!quityourjob +(.+)*$", func(msg core.Message, matches []string) core.Response {
		msg.Send(matches[1]+", quit your job!")
		return bot.Stop()
	})

	bot.ListenFor("^ *!getajob +(.+)*$", func(msg core.Message, matches []string) core.Response {
		msg.Send(matches[1]+", get a job!")
		return bot.Stop()
	})

	bot.ListenFor("^ *!induct *$", func(msg core.Message, matches []string) core.Response {
		elapsed := time.Now().Sub(lastInduction)
		if elapsed.Seconds() < 60 {
			msg.Reply("Rule "+strconv.Itoa(len(induction))+", dumbass.")
		} else {
			lastInduction = time.Now()
			for i := 0; i < len(induction); i++ {
				msg.Reply(induction[i])
			}
		}
		return bot.Stop()
	})

	bot.ListenFor("^ *!induct +(.+)", func(msg core.Message, matches []string) core.Response {
		elapsed := time.Now().Sub(lastInduction)
		if elapsed.Seconds() < 60 {
			msg.Reply("Rule "+strconv.Itoa(len(induction))+", dumbass.")
		} else {
			lastInduction = time.Now()
			for i := 0; i < len(induction); i++ {
				msg.Send(matches[1]+": "+induction[i])
			}
		}
		return bot.Stop()
	})

	// generate a chain using any word in a given string
	// i.e.: "gumbot: hi there" starts a chain from "hi" or "there"
	bot.ListenFor(fmt.Sprintf("^ *%s:.*$", bot.Name), func(msg core.Message, matches []string) core.Response {
		mutex.Lock()
		defer mutex.Unlock()

		tokens := strings.Split(matches[0], " ");
		if (len(tokens) <= 1) {
			return bot.Stop()
		}
		tokens2 := tokens[1:len(tokens)];

		i := rand.Intn(len(tokens2))

		fmt.Printf("Choosing element %v of %v\n", i, len(tokens2))

		output, err := generateRandomSeeded(tokens2[i])
		if err != nil {
			return bot.Error(err)
		}
		msg.Send(output)
		return bot.Stop()
	})

	bot.ListenFor("^ *markov *$", func(msg core.Message, matches []string) core.Response {
		mutex.Lock()
		defer mutex.Unlock()
		output, err := generateRandom()
		if err != nil {
			return bot.Error(err)
		}
		msg.Send(output)
		return bot.Stop()
	})

	// generate a chain for the specified user
	bot.ListenFor("^ *markov +(.+)", func(msg core.Message, matches []string) core.Response {
		mutex.Lock()
		defer mutex.Unlock()
		output, err := generate(matches[1])
		if err != nil {
			return bot.Error(err)
		}
		msg.Send(output)
		return bot.Stop()
	})

	// listen to everything
	bot.ListenFor("(.*)", func(msg core.Message, matches []string) core.Response {
		mutex.Lock()
		defer mutex.Unlock()
		user := msg.User
		text := matches[0]
		record(user, text)

		foobar := rand.Intn(100)
		if foobar < 1 {
// Throw this in, to shake the alg up.
//			tokens := strings.Split(matches[0], " ");
//			if (len(tokens) <= 1) {
//				return bot.Stop()
//			}
//			tokens2 := tokens[1:len(tokens)];
//
//			i := rand.Intn(len(tokens2))
//
//			fmt.Printf("Choosing element %v of %v\n", i, len(tokens2))
//
//			output, err := generateRandomSeeded(tokens2[i])
//			if err != nil {
//				return bot.Error(err)
//			}
//			msg.Send(output)
//			return bot.Stop()

			fmt.Printf("DING DING: random winner!\n")

			output, err := generateRandom()
			if err != nil {
				return bot.Error(err)
			}
			msg.Send(output)
		}
		return bot.KeepGoing()
	})
}

// getChainMap gets the map for a particular user, or a new map with all of the data for all users
func getChainMap(user string) (map[string][]string, error) {
	if user != "" {
		userMap, ok := markov.Chains[user]
		if !ok {
			return nil, fmt.Errorf("No chain could be found for %s", user)
		}
		return userMap, nil
	}
	if len(markov.Chains) == 0 {
		return nil, fmt.Errorf("No chains could be found")
	}
	// combine all of the users' maps
	result := make(map[string][]string)
	for _, userChainMap := range markov.Chains {
		// userChainMap is a map[string][]string
		for prefix, userChain := range userChainMap {
			chain := result[prefix]
			if chain != nil {
				chain = make([]string,0)
			}
			for _, chainItem := range userChain {
				chain = append(chain, chainItem)
			}
			result[prefix] = chain
		}
	}
	return result, nil
}

func generateRandom() (string, error) {
	return generate("")
}

func generateRandomSeeded(seed string) (string, error) {
	chainMap, err := getChainMap("")
	if err != nil {
		return "", err
	}
	p := newPrefix(markov.PrefixLength)
	var words []string
	words = append(words, seed)
	p.shift(seed)

	fmt.Printf("p: %s\n", p.String())

	for i := 0; i < maxWords; i++ {
		choices := chainMap[p.String()]
		fmt.Printf("choices: len: %v\n", len(choices))
		if len(choices) == 0 {
			break
		}
		choice := rand.Intn(len(choices))
		fmt.Printf("choice: %v: %s\n", choice, choices[choice])
		next := choices[choice]
		words = append(words, next)
		p.shift(next)
	}
	return strings.Join(words, " "), nil
}

func generate(user string) (string, error) {
	chainMap, err := getChainMap(user)
	if err != nil {
		return "", err
	}
	p := newPrefix(markov.PrefixLength)
	var words []string
	for i := 0; i < maxWords; i++ {
		choices := chainMap[p.String()]
		if len(choices) == 0 {
			break
		}
		i := rand.Intn(len(choices))
		next := choices[i]
		fmt.Printf("DING: Chosen %v from %v: %s\n", i, len(choices), next)
		words = append(words, next)
		p.shift(next)
	}

	var neuwords []string
	_, ok := markov.SuffixCounts[user]
	if !ok {
		//weights := make(map[string]map[string]int)
		markov.SuffixCounts[user] = make(map[string]map[string]int)
	}
	p = newPrefix(markov.PrefixLength)
	choices := chainMap[""]
	next := choices[rand.Intn(len(choices))]
	neuwords = append(neuwords, next)
	p.shift(next)
	for i := 1; i < maxWords; i++ {
		weightedChoices := markov.SuffixCounts[user][p.String()]
		if (len(weightedChoices) == 0) {
			fmt.Printf("NEU: WEIGHTS: %v has no weights\n", p.String())
			break
		}
		total := 0
		for i := range weightedChoices {
			total += weightedChoices[i]
			//fmt.Printf("WEIGHT: %v -> %v: %v\n", p.String(), i, weightedChoices[i])
		}
		r := rand.Intn(total) + 1
		fmt.Printf("NEU: Probabilities: number elements: %v, total weight: %v, rand: %v\n", len(weightedChoices), total, r)
		total = 0
		var j string
		for j = range weightedChoices {
			total += weightedChoices[j]
			if (total > r) {
				break
			}
		}
		fmt.Printf("NEU: Stopped at %s: w[j]:%v total:%v\n", j, weightedChoices[j], total);
		//next := weightedChoices[j]
		neuwords = append(neuwords, j)
		p.shift(j)
	}
	fmt.Printf("NEU: %s\n", strings.Join(neuwords, " "))

	return strings.Join(words, " "), nil
}

// record breaks up the text into tokens and then creates chains for that user
func record(user, text string) error {
	p := newPrefix(markov.PrefixLength)
	tokens := strings.Split(text, " ")
	userMap, ok := markov.Chains[user]
	if !ok {
		markov.Chains[user] = make(map[string][]string)
		//markov.SuffixWeights[user] = make(map[string]uint32)
		userMap = markov.Chains[user]
		//userWeights = markov.SuffixWeights[user]
	}
	for _, token := range tokens {
		if strings.HasPrefix("http", token) {
			continue
		}
		str := p.String()
		if !contains(userMap[str], token) {
			userMap[str] = append(userMap[str], token)
			//userWeights[str]
			p.shift(token)
			// only allow maxChainLength items in a particular chain for a prefix
			if len(userMap[str]) > maxChainLength {
				userMap[str] = userMap[str][len(userMap[str])-maxChainLength:]
			}
		}
	}

	userSuffixes, ok := markov.SuffixCounts[user]
	if !ok {
		markov.SuffixCounts[user] = make(map[string]map[string]int)
		userSuffixes = markov.SuffixCounts[user]
	}

	for i := 1; i < len(tokens); i++ {
		fmt.Printf(":: %s -> %s:\t", tokens[i-1], tokens[i]);

		var lastword = tokens[i-1]
		var nextword = tokens[i]

		var weight, ok = userSuffixes[lastword][nextword]
		if !ok {
			userSuffixes[lastword] = make(map[string]int)
			userSuffixes[lastword][nextword] = 0
		}
		fmt.Printf("weight: %u\n", weight)
		userSuffixes[lastword][nextword] += 1
	}

	return pluginState.Save(markov, false)
}

func contains(tokens []string, token string) bool {
	for _, word := range tokens {
		if word == token {
			return true
		}
	}
	return false
}

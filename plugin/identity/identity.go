// replies when someone mentions the bot's name
package identity

import (
	"fmt"
	"github.com/sdstrowes/gesture/core"
)

func Create(bot *core.Gobot, config map[string]interface{}) {
	name := bot.Name

	bot.ListenFor(fmt.Sprintf("(?i)kill %s", name), func(msg core.Message, matches []string) core.Response {
		msg.Reply("EAT SHIT")
		return bot.Stop()
	})

	bot.ListenFor(fmt.Sprintf("(?i)(hey|h(a?)i|hello) %s", name), func(msg core.Message, matches []string) core.Response {
		msg.Send(fmt.Sprintf("why, hello there %s", msg.User))
		return bot.Stop()
	})

	bot.ListenFor(fmt.Sprintf("(?i)(H|h)appy new year(!*|,) %s", name), func(msg core.Message, matches []string) core.Response {
		msg.Send(fmt.Sprintf("Happy new year to you too, %s!", msg.User))
		return bot.Stop()
	})
}


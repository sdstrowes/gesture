// Talkin'
package core

import (
	"fmt"
	"github.com/sdstrowes/gesture/rewrite"
	"github.com/sdstrowes/gesture/util"
	irc "github.com/fluffle/goirc/client"
)

type Message struct {
	conn    *irc.Conn
	line    *irc.Line
	User    string
	Channel string
	Text    string
}

const maxMsgSize int = 490

func (msg *Message) Names() []string {
	result := make([]string, 0)
	nicks := msg.conn.StateTracker().GetChannel(msg.Channel).Nicks
	for nk, _ := range nicks {
		result = append(result, nk)
	}
	return result
}

func (msg *Message) Send(message string) {
	for _, chunk := range util.StringSplitN(rewrite.Rewrite(message), maxMsgSize) {
		msg.conn.Privmsg(msg.Channel, chunk)
	}
}

func (msg *Message) SendPriv(message string) {
	for _, chunk := range util.StringSplitN(rewrite.Rewrite(message), maxMsgSize) {
		msg.conn.Privmsg(msg.User, chunk)
	}
}

func (msg *Message) Reply(message string) {
	msg.Send(fmt.Sprintf("%s: %s", msg.User, message))
}

func (msg *Message) Ftfy(message string) {
	msg.Send(fmt.Sprintf("%s: ftfy -> %s", msg.User, message))
}

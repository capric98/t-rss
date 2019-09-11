package irctool

import (
	"crypto/tls"
	"fmt"

	irc "github.com/thoj/go-ircevent"
)

func newIRC(addr string, user string, channel string, key string) {
	ircconn := irc.IRC("Test", user)
	ircconn.VerboseCallbackHandler = true
	ircconn.UseTLS = true
	ircconn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	ircconn.AddCallback("001", func(e *irc.Event) { ircconn.Join(channel) })
	ircconn.AddCallback("PRIVMSG", func(e *irc.Event) {
		fmt.Println(e.Message())
	})
	err := ircconn.Connect(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	ircconn.SendRaw(key)
	ircconn.Loop()
}

package ticker

import "github.com/capric98/t-rss/feed"

// Ticker struct
type Ticker struct {
	c chan []feed.Item
}

// C :)
func (t *Ticker) C() <-chan []feed.Item {
	return t.c
}

// Stop :)
func (t *Ticker) Stop() {
	close(t.c)
}

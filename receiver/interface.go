package receiver

import "github.com/capric98/t-rss/feed"

// Receiver interface
type Receiver interface {
	Push(*feed.Item, []byte) error
	Name() string
}

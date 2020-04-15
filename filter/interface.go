package filter

import "github.com/capric98/t-rss/feed"

// Filter interface.
type Filter interface {
	Check(*feed.Item) error
}

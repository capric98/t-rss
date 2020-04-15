package filter

import (
	"fmt"

	"github.com/capric98/t-rss/feed"
	"github.com/capric98/t-rss/unit"
)

type contentSizeFilter struct {
	min, max int64
}

// NewContentSizeFilter :)
func NewContentSizeFilter(min, max int64) Filter {
	return &contentSizeFilter{
		min: min,
		max: max,
	}
}

// Check meets Filter.Check() interface.
func (f *contentSizeFilter) Check(v *feed.Item) error {
	if v.Len == 0 {
		return nil
	} // check it later
	if v.Len < f.min {
		return fmt.Errorf("content_size: %v < minSize:%v", unit.FormatSize(v.Len), unit.FormatSize(f.min))
	}
	if v.Len > f.max {
		return fmt.Errorf("content_size: %v > maxSize:%v", unit.FormatSize(v.Len), unit.FormatSize(f.max))
	}
	return nil
}

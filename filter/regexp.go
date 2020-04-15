package filter

import (
	"fmt"

	"github.com/capric98/t-rss/feed"
	"github.com/capric98/t-rss/setting"
)

type regexpFilter struct {
	accept, reject []setting.Reg
}

// NewRegexpFilter :)
func NewRegexpFilter(accept, reject []setting.Reg) Filter {
	regf := &regexpFilter{
		accept: accept,
		reject: reject,
	}
	if len(regf.accept) == 0 {
		regf.accept = nil
	}
	return regf
}

// Check meets Filter.Check() interface.
func (f *regexpFilter) Check(v *feed.Item) error {
	for _, r := range f.reject {
		// if r.R.MatchString(v.Description) {
		// 	return true, r.C
		// }
		if r.R.MatchString(v.Title) || r.R.MatchString(v.Author) {
			return fmt.Errorf("regexp: matched - %v", r.C)
		}
	}
	if f.accept != nil {
		for _, r := range f.accept {
			if r.R.MatchString(v.Title) || r.R.MatchString(v.Author) {
				return nil
			}
		}
		return fmt.Errorf("regexp: no match of Accept")
	}
	return nil
}

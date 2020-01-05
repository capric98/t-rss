package core

import (
	"fmt"

	"github.com/capric98/t-rss/torrents/bencode"
)

func (w *worker) editTorrent(data []byte) (en []byte, e error) {
	defer func() {
		if err := recover(); err != nil {
			e = err.(error)
		}
	}()

	results, err := bencode.Decode(data)
	if err != nil || len(results) != 1 {
		w.log(fmt.Sprintf("Edit tracker: %v", err), 1)
	}
	torrent := results[0]

	if w.Config.DeleteT != nil {
		announce := torrent.Dict("announce")
		for _, v := range w.Config.DeleteT {
			if v.R.Match(announce.BStr()) {
				w.log(fmt.Sprintf(" + edit tracker: \"%s\" matches \"%s\", delete announce.", announce.BStr(), v.C), 0)
				torrent.Delete("announce")
				break
			}
		}
		announceList := torrent.Dict("announce-list")
		for _, v := range w.Config.DeleteT {
			for i := announceList.Len(); i > 0; i-- {
				subList := announceList.List(i - 1)
				for s := subList.Len(); s > 0; s-- {
					if v.R.Match(subList.List(s - 1).BStr()) {
						w.log(fmt.Sprintf(" + edit tracker: \"%s\" matches \"%s\", delete part of announce-list.", subList.List(s-1).BStr(), v.C), 0)
						subList.DeleteN(s - 1)
						break
					}
				}
				if subList.Len() == 0 {
					announceList.DeleteN(i - 1)
				}
			}
			if announceList.Len() == 0 {
				torrent.Delete("announce")
			}
		}

	}

	if w.Config.AddT != nil && len(w.Config.AddT) >= 1 {
		if torrent.Dict("announce") == nil {
			w.log(" + edit tracker: announce add "+"\""+w.Config.AddT[0]+"\"", 0)
			_ = torrent.AddPart("announce", bencode.NewBStr(w.Config.AddT[0]))
			w.Config.AddT = w.Config.AddT[1:]
		}
		if len(w.Config.AddT) == 0 {
			return torrent.Encode()
		}

		if torrent.Dict("announce-list") == nil {
			_ = torrent.AddPart("announce-list", bencode.NewEmptyList())
		}
		list := torrent.Dict("announce-list")
		list.AnnounceList(w.Config.AddT)
	}
	w.log(fmt.Sprintf(" + edit tracker: check %v", torrent.Check()), 0)

	return torrent.Encode()
}

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
			if v.Match(announce.BStr()) {
				torrent.Delete("announce")
				break
			}
		}
		announceList := torrent.Dict("announce-list")
		for _, v := range w.Config.DeleteT {
			for i := announceList.Len(); i > 0; i-- {
				subList := announceList.List(i - 1)
				for s := subList.Len(); s > 0; s-- {
					if v.Match(subList.List(s - 1).BStr()) {
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
	w.log("Success Delete", 1)
	if w.Config.AddT != nil && len(w.Config.AddT) >= 1 {
		if torrent.Dict("announce") == nil {
			w.log("announce add", 1)
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
	w.log("Success Add", 1)
	w.log(torrent.Check(), 1)

	return torrent.Encode()
}

package trss

import (
	"fmt"

	"github.com/capric98/t-rss/bencode"
	"github.com/sirupsen/logrus"
)

func (w *worker) editTorrent(data []byte) (en []byte) {
	defer func() { _ = recover() }()

	log := w.logger().WithFields(logrus.Fields{
		"@func": "editTorrent",
	})

	results, err := bencode.Decode(data)
	if err != nil || len(results) != 1 {
		log.Warn("decode: ", err)
		return
	}
	torrent := results[0]

	for _, reg := range w.edit.Tracker.Delete {
		announce := torrent.Dict("announce")
		if announce != nil {
			if reg.R.Match(announce.BStr()) {
				log.Debug(fmt.Sprintf(" + edit tracker: \"%s\" matches \"%s\", delete announce.", announce.BStr(), reg.C), 0)
				torrent.Delete("announce")
				continue
			}
		}
		announceList := torrent.Dict("announce-list")
		if announceList != nil {
			for i := announceList.Len(); i > 0; i-- {
				subList := announceList.List(i - 1)
				for s := subList.Len(); s > 0; s-- {
					if reg.R.Match(subList.List(s - 1).BStr()) {
						log.Debug(fmt.Sprintf(" + edit tracker: \"%s\" matches \"%s\", delete part of announce-list.", subList.List(s-1).BStr(), reg.C), 0)
						subList.DeleteN(s - 1)
						break
					}
				}
				if subList.Len() == 0 {
					announceList.DeleteN(i - 1)
				}
			}
		}
		if announceList.Len() == 0 {
			torrent.Delete("announce-list")
		}
	}

	var waitList []string
	for _, add := range w.edit.Tracker.Add {
		if torrent.Dict("announce") == nil {
			log.Debug("announce add "+"\""+add+"\"", 0)
			_ = torrent.AddPart("announce", bencode.NewBStr(add))
			continue
		}
		waitList = append(waitList, add)
	}
	if torrent.Dict("announce-list") == nil {
		_ = torrent.AddPart("announce-list", bencode.NewEmptyList())
	}
	list := torrent.Dict("announce-list")
	list.AnnounceList(waitList)
	log.Debug("check ", torrent.Check())

	en, _ = torrent.Encode()

	return
}

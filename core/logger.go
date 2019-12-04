package core

import (
	"log"
)

func (w *worker) log(info interface{}, level int) {
	if info == nil {
		return
	}
	if level>=w.loglevel {
		log.Println(info)
	}
}
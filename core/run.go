package core

import (
	"sync"
)

func (w *worker)run(wg *sync.WaitGroup) {
	defer wg.Done()
}
package main

import (
	"testing"
	"time"

	"gitlab.clearwayintegration.com/go/modcomm/v3/dto"
)

const TOTAL_ITEMS = 100000

func TestBasics(t *testing.T) {
	q := NewActiveQueue[*dto.ModuleData]()
	for i := 1; i <= TOTAL_ITEMS; i++ {
		q.Push(&dto.ModuleData{Count: uint64(i)})
		//log.Printf("Pushing item #%d, count: %d", i, q.Count())
	}
	rcv := q.Receive()
	for i := 1; i <= TOTAL_ITEMS; i++ {
		_, ok := <-rcv
		if !ok {
			t.Error("queue died")
		}
		//log.Printf("Got: %d, count: %d", data.Count, q.Count())
	}
	time.Sleep(1 * time.Millisecond)
	if fc := q.Count(); fc != 0 {
		t.Errorf("Final count is wrong: want 0, got %d", fc)
	}
}

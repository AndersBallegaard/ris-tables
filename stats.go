package main

import (
	"time"
)

type EventStats struct {
	CurrentEvents  uint64
	LastTestTime   time.Time
	LastTestEvents uint64
	Rate           float64
}

func (e *EventStats) add() {
	e.CurrentEvents += 1
}

func (e *EventStats) getNumberOfEvents() uint64 {
	return e.CurrentEvents
}

func (e *EventStats) getRate() float64 {
	if time.Now().Sub(e.LastTestTime).Seconds() > 60 {
		e.Rate = (float64(e.CurrentEvents) - float64(e.LastTestEvents)) / time.Now().Sub(e.LastTestTime).Seconds()
		e.LastTestTime = time.Now()
	}
	return e.Rate
}

package main

import (
	"time"
)

type EventStats struct {
	currentEvents  uint64
	lastTestTime   time.Time
	lastTestEvents uint64
	rate           float64
}

func (e *EventStats) add() {
	e.currentEvents += 1
}

func (e *EventStats) getNumberOfEvents() uint64 {
	return e.currentEvents
}

func (e *EventStats) getRate() float64 {
	if time.Now().Sub(e.lastTestTime).Seconds() > 60 {
		e.rate = (float64(e.currentEvents) - float64(e.lastTestEvents)) / time.Now().Sub(e.lastTestTime).Seconds()
		e.lastTestTime = time.Now()
	}
	return e.rate
}

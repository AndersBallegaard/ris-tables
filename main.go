package main

import (
	"log"
)

func main() {
	log.Println("RIS-Tables server")
	collectors := CollectorGroup{}
	stats := EventStats{}

	RRCInfo := new(RRCInfoAPIResp)
	rrcinforesp := GenericHTTPGet("https://stat.ripe.net/data/rrc-info/data.json")
	RRCInfo.UnmarshalJSON(rrcinforesp)
	collectors.init_collector(*RRCInfo)

	go collectorWorker(&collectors, &stats)
	apiLoadAndRun(&collectors, &stats)

}

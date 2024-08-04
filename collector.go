package main

import (
	//"bufio"
	"encoding/json"
	"log"
	"net/http"
	"net/netip"
	"strings"
)

type RISAnnouncement struct {
	Next_hop string
	Prefixes []string
}

type RISEventData struct {
	Host          string
	Peer          string
	Peer_asn      string
	Path          []uint32
	Type          string
	Origin        string
	Withdrawals   []string
	Announcements []RISAnnouncement
	State         string
}

type RISEvent struct {
	Type string
	Data RISEventData
}

func (e *RISEvent) process(collectors *CollectorGroup, stats *EventStats) {
	rrc := strings.Split(strings.ToLower(e.Data.Host), ".")[0]
	collectors.collectors[rrc].EventHandler(e)
	stats.add()
}

func (c *RISCollector) EventHandler(e *RISEvent) {
	if e.Type == "ris_message" && e.Data.Type == "UPDATE" {
		c.update(e)
	}
	if e.Type == "RIS_PEER_STATE" && e.Data.State == "down" {
		c.peer_down(e)
	}
}

func (s *SAFI_Unicast) addBgpPath(prefix string, nexthop string, event *RISEvent) {
	p, err := netip.ParsePrefix(prefix)
	ErrorParserFatal(err)

	// Make sure the route exists in the BST
	s.insertRouteIfNew(p, s.version)

	route, err := s.Routes.getExactRoute(p)
	ErrorParserFatal(err)
	route.addPathFromRis(ASN(event.Data.Peer_asn), event.Data.Path, nexthop, event.Data.Origin, event.Data.Peer)
}

func (c *RISCollector) update(e *RISEvent) {
	if e.Data.Announcements != nil {
		for _, annoncement := range e.Data.Announcements {
			for _, prefix := range annoncement.Prefixes {
				// A better way to find the ip version must exist

				if strings.Contains(prefix, ".") {
					c.Routing_table.Ipv4.Unicast.addBgpPath(prefix, annoncement.Next_hop, e)
				}
				if strings.Contains(prefix, ":") {
					c.Routing_table.Ipv6.Unicast.addBgpPath(prefix, annoncement.Next_hop, e)
				}

			}
		}
	}
	if e.Data.Withdrawals != nil {
		for _, prefix := range e.Data.Withdrawals {
			log.Println("Withdrawl recived for", prefix, "from", e.Data.Peer_asn)
			p, err := netip.ParsePrefix(prefix)
			ErrorParserFatal(err)
			peer, err := netip.ParseAddr(e.Data.Peer)
			if strings.Contains(prefix, ":") {

				r, err := c.Routing_table.Ipv6.Unicast.Routes.getExactRoute(p)
				if err == nil {
					r.DeletePath(peer)
				}
			}

		}
	}
}
func (c *RISCollector) peer_down(e *RISEvent) {
	log.Println("Peer down recived, but cleanup not implemented yet")
}

func collectorWorker(collectors *CollectorGroup, stats *EventStats) {
	url := "https://ris-live.ripe.net/v1/stream/?format=json&client=ris-tables-anderstb@anderstb.dk"
	resp, err := http.Get(url)
	ErrorParserFatal(err)
	//t, _ := io.ReadAll(resp.Body)
	dec := json.NewDecoder(resp.Body)
	for {
		var event RISEvent
		err := dec.Decode(&event)
		if err != nil {

			// TODO: Fix unmarshal errors
			log.Println(err)
			foo := make(map[string]interface{})
			dec.Decode(&foo)
			log.Println(foo)

			//break
		}
		event.process(collectors, stats)
	}
	log.Println("FUCK")

}

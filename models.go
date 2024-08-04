package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/praserx/ipconv"
)

type IP_VERSION uint

const IPV4 = IP_VERSION(4)
const IPV6 = IP_VERSION(6)

type ASN string

type RRCFwdTable struct {
	Ipv4 *SAFIUnicastFwdTable
	Ipv6 *SAFIUnicastFwdTable
}

type fwdTableEntry struct {
	Prefix         *netip.Prefix
	Nexthop        *netip.Addr
	Nexthop_AS     ASN
	Nexthop_ASPATH []ASN
	Origin         string
}

func (f *fwdTableEntry) Len() uint64 {
	v := uint64(0)

	v += uint64(f.Prefix.Bits())
	addr := f.Prefix.Addr()
	if strings.Contains(addr.String(), ":") {
		bi, _ := ipconv.IPv6ToBigInt(net.ParseIP(addr.String()))

		v += uint64(bi.Int64())

	} else {
		i, _ := ipconv.IPv4ToInt(net.ParseIP(addr.String()))
		v += uint64(i)
	}

	return v
}

type SAFIUnicastFwdTable struct {
	Table []fwdTableEntry
	Count uint32
}

type BGPPath struct {
	Peer_asn ASN
	Peer_ip  netip.Addr
	Aspath   []ASN
	Nexthop  netip.Addr
	Origin   string
}

type RouteBSTNode struct {
	Version IP_VERSION
	Prefix  netip.Prefix
	Paths   []*BGPPath
	Right   *RouteBSTNode
	Left    *RouteBSTNode
}

func (r *RouteBSTNode) bestPath() fwdTableEntry {
	f := fwdTableEntry{Prefix: &r.Prefix}
	for _, path := range r.Paths {
		if f.Nexthop == nil {
			f.Nexthop = &path.Nexthop
			f.Nexthop_AS = path.Peer_asn
			f.Nexthop_ASPATH = path.Aspath
			f.Origin = path.Origin
		} else {
			// Waaaay too basic implementation but whatever
			if len(path.Aspath) < len(f.Nexthop_ASPATH) {
				f.Nexthop = &path.Nexthop
				f.Nexthop_AS = path.Peer_asn
				f.Nexthop_ASPATH = path.Aspath
				f.Origin = path.Origin
			}

		}
	}
	return f
}

func (r *RouteBSTNode) getForwardingTables() []fwdTableEntry {
	var fwdtable []fwdTableEntry

	fwdtable = append(fwdtable, r.bestPath())
	if r.Right != nil {
		rfw := r.Right.getForwardingTables()
		fwdtable = append(fwdtable, rfw...)
	}
	if r.Left != nil {
		lfw := r.Left.getForwardingTables()
		fwdtable = append(fwdtable, lfw...)
	}

	return fwdtable
}

func (r *RouteBSTNode) insertRouteIfNew(prefix netip.Prefix, version IP_VERSION) error {

	if prefix.Addr() == r.Prefix.Addr() && prefix.Bits() == r.Prefix.Bits() {
		return nil
	}
	rp, err := r.Prefix.MarshalBinary()
	ErrorParserFatal(err)
	p, err := prefix.MarshalBinary()
	rps := string(rp)
	ps := string(p)
	//print(string(p), " The address should be fully expanded")
	ErrorParserFatal(err)
	if ps < rps {
		if r.Left == nil {
			//log.Println("Adding prefix", prefix)
			r.Left = &RouteBSTNode{Prefix: prefix, Version: version}
		} else {
			r.Left.insertRouteIfNew(prefix, version)
		}
	} else {
		if r.Right == nil {
			//log.Println("Adding prefix", prefix)
			r.Right = &RouteBSTNode{Prefix: prefix, Version: version}
		} else {
			r.Right.insertRouteIfNew(prefix, version)
		}
	}
	return nil
}

func (r *RouteBSTNode) getExactRoute(prefix netip.Prefix) (*RouteBSTNode, error) {
	if r == nil {
		return nil, errors.New("No matching route was found")
	}
	rBits, err := r.Prefix.MarshalBinary()
	ErrorParserFatal(err)
	rs := string(rBits)

	pBits, err := prefix.MarshalBinary()
	ErrorParserFatal(err)
	ps := string(pBits)

	if rs == ps {
		return r, nil
	}
	if ps < rs {
		if r.Left == nil {
			return nil, errors.New("No matching route was found")
		} else {
			return r.Left.getExactRoute(prefix)
		}
	} else {
		if r.Right == nil {
			return nil, errors.New("No matching route was found")
		} else {
			return r.Right.getExactRoute(prefix)
		}
	}

}

func (r *RouteBSTNode) addPathFromRis(Peer_asn ASN, Aspath []uint32, RisNexthop string, origin string, Peer_ip_str string) {
	Peer_ip, err := netip.ParseAddr(Peer_ip_str)
	Nexthop, err := netip.ParseAddr(strings.Split(RisNexthop, ",")[0])
	ErrorParserFatal(err)
	if r.Paths == nil {
		r.Paths = make([]*BGPPath, 0)
	}
	foundMatch := false
	for _, path := range r.Paths {
		if Peer_asn == path.Peer_asn && Nexthop == path.Nexthop && Peer_ip == path.Peer_ip {
			foundMatch = true
			var asp []ASN
			for _, as := range Aspath {
				asp = append(asp, ASN(strconv.Itoa(int(as))))
			}
			path.Aspath = asp
			path.Origin = origin

		}
	}
	if !foundMatch {
		var asp []ASN
		for _, as := range Aspath {
			asp = append(asp, ASN(strconv.Itoa(int(as))))
		}
		r.Paths = append(r.Paths, &BGPPath{Peer_asn: Peer_asn, Aspath: asp, Nexthop: Nexthop, Origin: origin, Peer_ip: Peer_ip})
	}

}

func (r *RouteBSTNode) DeletePath(Peer_ip netip.Addr) {

	var new_Paths []*BGPPath
	for _, path := range r.Paths {
		if path.Peer_ip != Peer_ip {
			new_Paths = append(new_Paths, path)
		}
	}
	r.Paths = new_Paths

}

type SAFI interface {
	get()
}

type SAFI_Unicast struct {
	Version IP_VERSION
	Routes  *RouteBSTNode
}

func (s *SAFI_Unicast) insertRouteIfNew(prefix netip.Prefix, version IP_VERSION) {
	if s.Routes == nil {
		s.Routes = &RouteBSTNode{Prefix: prefix, Version: version}
	} else {
		s.Routes.insertRouteIfNew(prefix, version)
	}
}

type AFI struct {
	Version IP_VERSION
	Unicast SAFI_Unicast
}

type BGPTable struct {
	Ipv4 AFI
	Ipv6 AFI
}

type Collector interface {
	init_tables()
}

type RISCollector struct {
	Name          string
	Location      string
	Routing_table BGPTable
}

func (g *CollectorGroup) init_collector(info RRCInfoAPIResp) {
	if g.Collectors == nil {
		g.Collectors = make(map[string]*RISCollector)
		g.Stats = EventStats{}
		for _, rrc := range info.Data.Rrcs {
			c := RISCollector{Name: strings.ToLower(rrc.Name), Location: rrc.Geographical_location}
			c.init_tables()
			g.Collectors[c.Name] = &c
		}
	}

}

func (c *RISCollector) init_tables() {
	c.Routing_table = BGPTable{}
	c.Routing_table.Ipv4 = AFI{Version: IPV4}
	c.Routing_table.Ipv4.Unicast = SAFI_Unicast{Version: IPV4}

	c.Routing_table.Ipv6 = AFI{Version: IPV6}
	c.Routing_table.Ipv6.Unicast = SAFI_Unicast{Version: IPV6}
}

func (c *SAFI_Unicast) getForwardingTables() *SAFIUnicastFwdTable {
	fwd := SAFIUnicastFwdTable{}
	fwd.Table = c.Routes.getForwardingTables()
	fwd.Count = uint32(len(fwd.Table))
	return &fwd
}

func (c *RISCollector) getForwardingTables() RRCFwdTable {
	fwd := RRCFwdTable{}
	fwd.Ipv4 = c.Routing_table.Ipv4.Unicast.getForwardingTables()
	fwd.Ipv6 = c.Routing_table.Ipv6.Unicast.getForwardingTables()

	return fwd
}

type CollectorGroup struct {
	Collectors map[string]*RISCollector
	Stats      EventStats
}

func loadFromDisk() CollectorGroup {
	filename := "ris.json"
	ncg := CollectorGroup{}
	// Read from file if posible
	if _, err := os.Stat(filename); err == nil {
		log.Println("Found data file, loading from disk")
		dat, err := os.ReadFile(filename)
		ErrorParserFatal(err)
		json.Unmarshal(dat, &ncg)
	}
	return ncg
}
func (cg *CollectorGroup) persistantSync() {
	filename := "ris.json"
	s30, _ := time.ParseDuration("30s")
	for {
		time.Sleep(s30)
		ob, err := json.MarshalIndent(cg, "", "	")
		if err != nil {
			log.Println(err)
		}
		os.WriteFile(filename, ob, 0664)
	}
}

type RRCInfoAPIRespRRC struct {
	Name                  string
	Geographical_location string
}

type RRCInfoAPIRespRRCS struct {
	Rrcs []RRCInfoAPIRespRRC
}

type RRCInfoAPIResp struct {
	Data RRCInfoAPIRespRRCS
}

package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func apiLoadAndRun(collectors *CollectorGroup, stats *EventStats) {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/", apiListCollectorsBuilder(collectors))
		api.GET("/status", apiStatsBuilder(collectors, stats))
		rrc := api.Group("/rrc")
		{
			rrc.GET("/bgp/*rrc", apiGetRRCBGPTableBuilder(collectors))
			rrc.GET("/fwd/*rrc", apiGetRRCFwdTableBuilder(collectors))
		}

	}

	router.Run("[::]:8085")
}

type StatsModel struct {
	CollectorCount       uint
	CollectorNames       []string
	TotalEventsProcessed uint64
	EventsPerSec         float64
	RateCalculated       time.Time
}

func apiStatsBuilder(collectors *CollectorGroup, stats *EventStats) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		o := StatsModel{}
		o.EventsPerSec = stats.getRate()
		o.TotalEventsProcessed = stats.getNumberOfEvents()
		o.CollectorCount = 0
		o.RateCalculated = stats.lastTestTime

		for _, collector := range collectors.collectors {
			o.CollectorCount += 1
			o.CollectorNames = append(o.CollectorNames, collector.Name)
		}

		c.JSON(http.StatusOK, o)
	}
	return gin.HandlerFunc(fn)
}

type ListCollectorsModel struct {
	Name         string
	Location     string
	BgpLink      string
	FwdTableLink string
	StatusLink   string
}

func apiListCollectorsBuilder(collectors *CollectorGroup) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var r []ListCollectorsModel

		for _, collector := range collectors.collectors {
			rrc := ListCollectorsModel{Name: collector.Name, Location: collector.Location, BgpLink: c.Request.Host + "/api/rrc/bgp/" + collector.Name, FwdTableLink: c.Request.Host + "/api/rrc/fwd/" + collector.Name, StatusLink: c.Request.Host + "/api/status"}
			r = append(r, rrc)
		}

		c.JSON(http.StatusOK, r)
	}
	return gin.HandlerFunc(fn)
}

func apiGetRRCBGPTableBuilder(collectors *CollectorGroup) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		found := false
		server := c.Param("rrc")[1:]

		for _, rrc := range collectors.collectors {
			if rrc.Name == server {
				c.JSON(http.StatusOK, rrc)
				found = true
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, "{}")
		}

	}
	return gin.HandlerFunc(fn)
}

func apiGetRRCFwdTableBuilder(collectors *CollectorGroup) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		found := false
		server := c.Param("rrc")[1:]

		for _, rrc := range collectors.collectors {
			if rrc.Name == server {
				c.JSON(http.StatusOK, rrc.getForwardingTables())
				found = true
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, "{}")
		}

	}
	return gin.HandlerFunc(fn)
}

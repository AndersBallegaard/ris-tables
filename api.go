package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func apiLoadAndRun(collectors *CollectorGroup) {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/", apiListCollectorsBuilder(collectors))
		rrc := api.Group("/rrc")
		{
			rrc.GET("/bgp/*rrc", apiGetRRCBGPTableBuilder(collectors))
			rrc.GET("/fwd/*rrc", apiGetRRCFwdTableBuilder(collectors))
		}

	}

	router.Run("[::]:8085")
}

type ListCollectorsModel struct {
	Name         string
	Location     string
	BgpLink      string
	FwdTableLink string
}

func apiListCollectorsBuilder(collectors *CollectorGroup) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var r []ListCollectorsModel

		for _, collector := range collectors.collectors {
			rrc := ListCollectorsModel{Name: collector.Name, Location: collector.Location, BgpLink: c.Request.Host + "/api/rrc/bgp/" + collector.Name, FwdTableLink: c.Request.Host + "/api/rrc/fwd/" + collector.Name}
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

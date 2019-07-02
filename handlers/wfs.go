package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/paulmach/orb/geojson"
	"github.com/sirupsen/logrus"
)

func (h *HTTPServer) setupWFSHandlers(opt Options) {
	h.router.GET("/collections/:collection/items", getFeatures(opt))
}

func getFeatures(opt Options) func(*gin.Context) {
	return func(c *gin.Context) {
		collection := c.Param("collection")

		bboxstr := c.Query("bbox")

		limitstr := c.Query("limit")
		if limitstr != "" {
			limitstr = fmt.Sprintf("&limit=%s", limitstr)
		}

		timestr := c.Query("time")
		if timestr != "" {
			timestr = fmt.Sprintf("&time=%s", timestr)
		}

		propertiesFilterStr := ""
		params := c.Request.URL.Query()
		for k, v := range params {
			if k != "time" && k != "bbox" && k != "limit" {
				propertiesFilterStr = fmt.Sprintf("%s&%s=%s", propertiesFilterStr, k, v)
			}
		}

		q := fmt.Sprintf("%s/collections/%s/items?bbox=%s%s%s%s", opt.WFSURL, collection, bboxstr, limitstr, timestr, propertiesFilterStr)
		logrus.Debugf("WFS query: %s", q)
		resp, err := http.Get(q)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Error requesting WFS service. err=%s", err)})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("WFS invocation status != 200. status=%d", resp.StatusCode)
			data1, err1 := ioutil.ReadAll(resp.Body)
			if err1 != nil {
				logrus.Warnf("Error reading failed WFS response body. err=%s", err1)
				c.JSON(resp.StatusCode, gin.H{"message": msg})
			} else {
				c.JSON(resp.StatusCode, gin.H{"message": msg, "body": string(data1)})
			}
			return
		}

		var fc geojson.FeatureCollection
		data, err0 := ioutil.ReadAll(resp.Body)
		if err0 != nil {
			msg := fmt.Sprintf("Error reading WFS service response. err=%s", err0)
			logrus.Errorf(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"message": msg})
			return
		}

		logrus.Debugf("WFS response bytes: %d", len(data))
		err = json.Unmarshal(data, &fc)
		if err != nil {
			msg := fmt.Sprintf("Error parsing WFS service response. err=%s", err)
			logrus.Errorf(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"message": msg})
			return
		}
		logrus.Debugf("WFS response feature count: %d", len(fc.Features))

		c.JSON(http.StatusOK, fc)
	}
}
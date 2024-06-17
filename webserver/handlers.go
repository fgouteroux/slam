package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/fgouteroux/slam/redis"
	alertmanagerTmpl "github.com/fgouteroux/slam/template"
)

func (ws *webserver) healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func (ws *webserver) handleWebhook(c *gin.Context) error {
	log.Debug(c.Request.Header)

	msg := &alertmanagerTmpl.Data{}
	err := c.ShouldBindJSON(msg)
	if err != nil {
		return err
	}

	channelName := c.Param("channel")
	templateName := c.Query("template")

	key := Hash(msg.GroupKey)

	renderedStr, err := renderTemplate(msg, templateName)
	if err != nil {
		return err
	}

	color := "danger"
	if msg.Status == "resolved" {
		color = "good"
	}

	var channelID, timestamp string
	
	// Try to get the value from the cache
	if ws.Cache == "local" {
		cached, ok := localCache.Get(key)
		if ok {
			data := cached.Value.(map[string]string)
			timestamp = data["timestamp"]
			channelID = data["channelID"]
		}
	} else if ws.Cache == "redis" {
		jsonData, err := redis.Get(redisCache, key)
		if err == nil {
			var data map[string]string
			json.Unmarshal([]byte(jsonData), &data)
			timestamp = data["timestamp"]
			channelID = data["channelID"]
		}
	}

	if msg.Status == "firing" {
		channelID, ts, err := ws.sendSlackMessage(
			channelName,
			msg.CommonAnnotations["summary"],
			msg.CommonAnnotations["title_link"],
			renderedStr,
			"",
			color,
			"",
			false,
		)
		if err != nil {
			return err
		}
		data := map[string]string{
			"status":    msg.Status,
			"timestamp": ts,
			"channelID": channelID,
		}

		if ws.Cache == "local" {
			localCache.Set(key, data)
		} else if ws.Cache == "redis" {
			dataraw, _ := json.Marshal(data)
			err = redis.Set(redisCache, key, []byte(dataraw), 0)
			if err != nil {
				log.Errorf("Could not set key in redis: %v", err)
			}
		}

	} else {

		if timestamp != "" {
			// update color of original message
			_, timestamp, err = ws.sendSlackMessage(
				channelID,
				msg.CommonAnnotations["summary"],
				msg.CommonAnnotations["title_link"],
				renderedStr,
				fmt.Sprintf("Resolved at %s", timeNowToDateTimeFormatted()),
				color,
				timestamp,
				true,
			)
			if err != nil {
				return err
			}
			// remove key from cache
			if ws.Cache == "local" {
				localCache.Del(key)
			} else if ws.Cache == "redis" {
				redis.Delete(redisCache, key)
			}

		} else {
			log.Infof("Key '%s' not found in cache, couldn't update original message.", key)
			_, timestamp, err = ws.sendSlackMessage(
				channelName,
				msg.CommonAnnotations["summary"],
				msg.CommonAnnotations["title_link"],
				renderedStr,
				fmt.Sprintf("Resolved at %s", timeNowToDateTimeFormatted()),
				color,
				"",
				false,
			)
			if err != nil {
				return err
			}
		}
	}
	_, err = c.Writer.WriteString("ok")
	if err != nil {
		return err
	}
	return nil
}

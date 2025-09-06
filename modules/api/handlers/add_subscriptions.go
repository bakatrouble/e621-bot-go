package handlers

import (
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddSubscriptionsHandler(c *gin.Context) {
	store := c.Value("store").(*storage.Storage)

	var body subsRequestBody
	var err error

	if err = c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Invalid request body: %s", err.Error()),
		})
		return
	}

	var existingSubs map[string]struct{}
	if existingSubs, err = store.GetSubsMap(c); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	conflicts := make([]string, 0)
	invalids := make([]string, 0)
	for _, sub := range body.Subs {
		sub = strings.ToLower(sub)

		if _, exists := existingSubs[sub]; exists {
			conflicts = append(conflicts, sub)

			if _, err = utils.QueryParser.ParseString("", sub); err != nil {
				invalids = append(invalids, sub)
				continue
			}
		}
	}
	if len(invalids) > 0 {
		slices.Sort(invalids)
		c.JSON(400, gin.H{"status": "error", "message": "Some subscriptions are invalid", "invalid": invalids})
		return
	}
	if len(conflicts) > 0 {
		slices.Sort(conflicts)
		c.JSON(409, gin.H{"status": "error", "message": "Some subscriptions already exist", "conflicts": conflicts})
		return
	}

	for _, sub := range body.Subs {
		if err = store.AddSub(c, sub); err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}
	}

	slices.Sort(body.Subs)
	c.JSON(200, gin.H{"status": "ok", "added": body.Subs})
}

package handlers

import (
	"e621-bot-go/storage"
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func DeleteSubscriptionsHandler(c *gin.Context) {
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

	missing := make([]string, 0)
	for _, sub := range body.Subs {
		sub = strings.ToLower(sub)

		if _, exists := existingSubs[sub]; !exists {
			missing = append(missing, sub)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		c.JSON(404, gin.H{"status": "error", "message": "Some subscriptions were not found", "missing": missing})
		return
	}
}

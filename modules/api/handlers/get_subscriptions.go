package handlers

import (
	"e621-bot-go/storage"

	"github.com/gin-gonic/gin"
)

func GetSubscriptionsHandler(c *gin.Context) {
	store := c.Value("store").(*storage.Storage)
	var subs []string
	var err error
	if subs, err = store.GetSubs(c); err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success", "subscriptions": subs})
}

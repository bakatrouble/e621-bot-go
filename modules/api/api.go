package api

import (
	"context"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/samber/slog-gin"
)

type subsRequestBody struct {
	Subs []string `json:"subs" binding:"required"`
}

func StartAPI(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	store := ctx.Value("store").(*storage.Storage)

	logger := utils.NewLogger("api")
	logger.With("host", "127.0.0.1").With("port", config.Api.Port).Info("starting api")

	defer wg.Done()

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.With("method", httpMethod).With("path", absolutePath).With("handlers", nuHandlers).Debug("registered route")
	}
	gin.DebugPrintFunc = func(format string, v ...any) {
		logger.Debug(format, v...)
	}
	if config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	setContextValues := func(c *gin.Context) {
		c.Set("config", config)
		c.Set("store", store)
		c.Set("logger", logger)
	}

	router, _ := graceful.Default(
		graceful.WithAddr(fmt.Sprintf("127.0.0.1:%d", config.Api.Port)),
	)

	router.Use(sloggin.New(logger))
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	router.Use(func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if !slices.Contains(config.Api.Keys, apiKey) {
			c.JSON(403, gin.H{"status": "error", "message": "Forbidden"})
			c.Abort()
			return
		}
		setContextValues(c)
		c.Next()
	})

	router.GET("/api/subscriptions", func(c *gin.Context) {
		store := c.Value("store").(*storage.Storage)
		var subs []string
		var err error
		if subs, err = store.GetSubs(c); err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success", "subscriptions": subs})
	})

	router.POST("/api/subscriptions", func(c *gin.Context) {
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
		for _, sub := range body.Subs {
			if _, exists := existingSubs[sub]; exists {
				conflicts = append(conflicts, sub)
			}
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
	})

	router.DELETE("/api/subscriptions", func(c *gin.Context) {
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
			if _, exists := existingSubs[sub]; !exists {
				missing = append(missing, sub)
			}
		}
		if len(missing) > 0 {
			slices.Sort(missing)
			c.JSON(404, gin.H{"status": "error", "message": "Some subscriptions were not found", "missing": missing})
			return
		}
	})

	if err := router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.With("err", err).Error("failed to start web api")
	}
}

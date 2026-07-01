package api

import (
	"context"
	"e621-bot-go/modules/api/handlers"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"errors"
	"net/http"
	"slices"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/graceful"
	"github.com/gin-gonic/gin"
	"github.com/samber/slog-gin"
)

func StartAPI(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)
	store := ctx.Value("store").(*storage.Storage)
	metrics := ctx.Value("metrics").(*utils.Metrics)

	logger := utils.NewLogger("api")
	logger.With("bind", config.Api.Bind).Info("starting api")

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
		graceful.WithAddr(config.Api.Bind),
	)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "X-API-Key")
	corsConfig.AllowAllOrigins = true

	router.Use(sloggin.New(logger))
	router.Use(gin.Recovery())
	router.Use(cors.New(corsConfig))

	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/" || c.Request.URL.Path == "" {
			c.JSON(http.StatusOK, gin.H{"hello": "world"})
			c.Abort()
			return
		}

		apiKey := c.GetHeader("x-api-key")
		if !slices.Contains(config.Api.Keys, apiKey) {
			c.JSON(403, gin.H{"status": "error", "message": "Forbidden"})
			c.Abort()
			return
		}
		setContextValues(c)
		c.Next()
	})

	router.GET("/api/subscriptions", handlers.GetSubscriptionsHandler)
	router.POST("/api/subscriptions", handlers.AddSubscriptionsHandler)
	router.DELETE("/api/subscriptions", handlers.DeleteSubscriptionsHandler)

	router.GET("/metrics", gin.WrapH(metrics.PromHandler()))

	if err := router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.With("err", err).Error("failed to start web api")
		panic(err)
	}
}

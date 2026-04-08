package api

import (
	"context"
	"e621-bot-go/modules/api/handlers"
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

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "X-API-Key")
	corsConfig.AllowAllOrigins = true

	router.Use(sloggin.New(logger))
	router.Use(gin.Recovery())
	router.Use(cors.New(corsConfig))

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

	router.GET("/api/subscriptions", handlers.GetSubscriptionsHandler)
	router.POST("/api/subscriptions", handlers.AddSubscriptionsHandler)
	router.DELETE("/api/subscriptions", handlers.DeleteSubscriptionsHandler)

	if err := router.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.With("err", err).Error("failed to start web api")
		panic(err)
	}
}

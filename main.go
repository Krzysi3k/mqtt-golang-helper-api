package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

func main() {
	// r := gin.New()
	// r.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/"}}))

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.0.123:6379",
		Password: "",
		DB:       0,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/get-redis-data", GetRedisData(ctx, rdb))
	r.GET("/redis-info", GetRedisInfo(ctx, rdb))
	r.Run("0.0.0.0:5001")
}

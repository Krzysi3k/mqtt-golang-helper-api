package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

func GetRedisData(ctx context.Context, rdb *redis.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		keyName := c.Query("data")
		val, err := rdb.Get(ctx, keyName).Result()
		if err != nil {
			c.JSON(404, gin.H{"payload": "key not found"})
			return
		}

		result := make(map[string]interface{})
		if strings.Contains(val, "{") {
			json.Unmarshal([]byte(val), &result)
		} else if keyName == "docker-metrics-cpu" || keyName == "docker-metrics-mem" || keyName == "termometr-payload" {
			c.String(200, val)
			return
		} else {
			result["payload"] = val
		}
		c.JSON(200, result)
	}
}

func GetRedisInfo(ctx context.Context, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := []string{
			"vibration-sensor",
			"door-state",
			"rotate-option",
			"washing-state",
		}
		val := rdb.MGet(ctx, keys...).Val()
		combinedOutput := []map[string]interface{}{}
		for i := 0; i < len(keys); i++ {
			combinedOutput = append(combinedOutput, map[string]interface{}{keys[i]: val[i]})
		}
		redisKeys := rdb.Keys(ctx, "*").Val()
		combinedOutput = append(combinedOutput, map[string]interface{}{"Redis keys-in-use": len(redisKeys)})
		merged := mergeMaps(combinedOutput...)
		c.JSON(200, merged)
	}
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			value, ok := v.(string)
			if ok {
				if strings.Contains(value, "{") || strings.Contains(value, "[") {
					tmpObj := make(map[string]interface{})
					json.Unmarshal([]byte(value), &tmpObj)
					result[k] = tmpObj
					continue
				}
				result[k] = value
			} else {
				result[k] = v
			}
		}
	}
	return result
}
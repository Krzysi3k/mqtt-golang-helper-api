package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

// GET /get-redis-data?data=REDISKEY
func GetRedisData(ctx context.Context, rdb *redis.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		keyName := c.Query("data")
		if keyName == "" {
			c.JSON(400, gin.H{"payload": "missing query string"})
			return
		}
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

// GET /redis-info
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

// GET /docker-info?items=containers|images
func GetDockerInfo(ctx context.Context, dockerClient *client.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		queryParam := c.Query("items")

		if queryParam == "containers" {
			containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: false})
			logError(err)
			conArr := []string{}
			for _, container := range containers {
				conName := strings.Replace(container.Names[0], "/", "", -1)
				line := fmt.Sprintf("%v - %v", conName, container.Status)
				conArr = append(conArr, line)
			}
			c.JSON(200, gin.H{"containers": conArr})
		} else if queryParam == "images" {
			images, err := dockerClient.ImageList(ctx, types.ImageListOptions{All: true})
			logError(err)
			imgArr := []string{}
			for _, img := range images {
				if strings.Contains(img.RepoTags[0], "<none>") {
					continue
				}
				line := fmt.Sprintf("%v, %vMB", img.RepoTags[0], (img.Size / 1024 / 1024))
				imgArr = append(imgArr, line)
			}
			c.JSON(200, gin.H{"images": imgArr})
		} else {
			c.JSON(400, gin.H{"payload": "wrong or missing query string"})
		}
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

func logError(err error) {
	if err != nil {
		log.Println(err)
	}
}

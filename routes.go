package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
		if keyName == "docker-metrics-cpu" || keyName == "docker-metrics-mem" || keyName == "termometr-payload" {
			c.String(200, val)
			return
		}
		if strings.Contains(val, "[") || strings.Contains(val, "{") {
			c.Data(http.StatusOK, "application/json", []byte(val))
		} else {
			c.JSON(200, gin.H{"payload": val})
		}
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
		var sb strings.Builder
		sb.WriteString("{")
		for i := 0; i < len(keys); i++ {
			if val[i] != nil {
				if v, ok := val[i].(string); ok {
					if strings.Contains(v, "{") || strings.Contains(v, "[") {
						sb.WriteString(`"` + keys[i] + `":` + v + ",")
					} else {
						sb.WriteString(`"` + keys[i] + `":"` + v + `",`)
					}
				}
			}
		}
		redisKeys := rdb.Keys(ctx, "*").Val()
		sb.WriteString(fmt.Sprintf("\"Redis keys-in-use\":%v}", len(redisKeys)))
		// sBuilder.WriteString("}")
		// output := sb.String()
		// jsonOut := output[0:len(output)-2] + "}"
		c.Data(http.StatusOK, "application/json", []byte(sb.String()))
	}
}

// GET /docker-info?items=containers|images
func GetDockerInfo(ctx context.Context, dockerClient *client.Client) gin.HandlerFunc {

	return func(c *gin.Context) {
		queryParam := c.Query("items")

		if queryParam == "containers" {
			containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
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

func logError(err error) {
	if err != nil {
		log.Println(err)
	}
}

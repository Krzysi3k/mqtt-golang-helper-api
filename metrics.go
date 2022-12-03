package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "cpu_temperature_celsius",
	Help: "Current temperature of the CPU.",
})

func init() {
	prometheus.MustRegister(cpuTemp)
}

func PrometheusMetrics(ctx context.Context) gin.HandlerFunc {
	h := promhttp.Handler()
	cpuTemp.Set(21)
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

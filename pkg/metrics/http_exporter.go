package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GinHandleFunc is a gin handler which can expose metrics in http server
func GinHandleFunc(ctx *gin.Context) {
	promhttp.HandlerFor(GetSystemPrometheusRegistry(), promhttp.HandlerOpts{}).ServeHTTP(ctx.Writer, ctx.Request)
}

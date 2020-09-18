package ginhttp

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func ginIndex(c *gin.Context) {
	pprof.Index(c.Writer, c.Request)
}

func ginCmdline(c *gin.Context) {
	pprof.Cmdline(c.Writer, c.Request)
}

func ginProfile(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func ginSymbol(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func ginTrace(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

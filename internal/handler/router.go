package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yayakyogi/coffee-pos/pkg/response"
)

// NewRouter membuat dan mengonfigurasi Gin engine: mode dipilih berdasarkan
// appEnv, middleware bawaan dipasang, lalu route group /api/v1 dibuat.
func NewRouter(appEnv string) *gin.Engine {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// gin.New() dipakai (bukan gin.Default()) agar middleware dipasang eksplisit.
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.OK(c, "server is running", nil)
		})
	}

	return router
}

package http

import (
	"github.com/gin-gonic/gin"

	"azret/internal/http/handlers"
)

func NewRouter(users *handlers.UsersHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.Status(200) })
	// Swagger UI served from embedded files at /swagger
	r.StaticFS("/swagger", swaggerFileSystem())
	// Ensure /swagger (without trailing slash) serves the UI (not a 404)
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/users", users.GetUser)
	}
	return r
}

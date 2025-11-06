package routes

import (
	v1 "server-monitoring/internal/routes/v1"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	srv := gin.Default()
	r := srv.Group("/api")

	{
		v1.RegisterV1Routes(r)
	}
	return srv
}

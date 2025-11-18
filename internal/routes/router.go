package routes

import (
	v1 "server-monitoring/internal/routes/v1"
	"server-monitoring/pkg/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	config := config.Load()
	srv := gin.Default()

	srv.Use(cors.New(cors.Config{
		AllowOrigins:     config.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r := srv.Group("/api")

	{
		v1.RegisterV1Routes(r)
	}
	return srv
}

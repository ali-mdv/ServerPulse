package routes

import (
	v1 "server-monitoring/internal/routes/v1"
	"server-monitoring/pkg/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	cfg := config.Load()
	srv := gin.Default()

	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// "*" with AllowCredentials is rejected by browsers. When the wildcard is
	// requested, echo the caller's Origin back so credentials keep working.
	if len(cfg.Origins) == 1 && cfg.Origins[0] == "*" {
		corsCfg.AllowOriginFunc = func(origin string) bool { return true }
	} else if len(cfg.Origins) > 0 {
		corsCfg.AllowOrigins = cfg.Origins
	}

	srv.Use(cors.New(corsCfg))

	r := srv.Group("/api")

	{
		v1.RegisterV1Routes(r, cfg)
	}
	return srv
}

package v1

import (
	"server-monitoring/pkg/config"

	"github.com/gin-gonic/gin"
)

func RegisterV1Routes(r *gin.RouterGroup, cfg *config.Config) {
	api := r.Group("/v1")

	{
		RegisterAuthRoutes(api)
		RegisterUserRoutes(api)
		RegisterDockerRoutes(api)
		RegisterReportRoutes(api)
		RegisterSystemRoutes(api)
		RegisterPM2Routes(api, cfg)
	}

}

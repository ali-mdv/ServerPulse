package v1

import (
	"github.com/gin-gonic/gin"
)

func RegisterV1Routes(r *gin.RouterGroup) {
	api := r.Group("/v1")

	{
		RegisterUserRoutes(api)
	}

}

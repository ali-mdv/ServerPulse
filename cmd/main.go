package main

import (
	"fmt"
	"server-monitoring/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	srv := gin.Default()
	srv.Run(fmt.Sprintf(":%d", cfg.Port))
}

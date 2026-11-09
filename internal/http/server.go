package http

import "github.com/gin-gonic/gin"

func Init() error {
	api := gin.Default()

	return api.Run("0.0.0.0:8080")
}

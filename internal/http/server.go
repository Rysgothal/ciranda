package http

import (
	"ciranda/internal/database"
	"ciranda/internal/utils"

	"github.com/gin-gonic/gin"
)

func Init() error {
	api := gin.Default()

	// Testando conexão com banco
	db, err := database.Open(utils.GetLocalPath())
	if err != nil {
		return err
	}

	api.GET("/books", listBooks)

	db.Close()
	return api.Run("0.0.0.0:8080")
}

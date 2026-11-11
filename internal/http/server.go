package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) Run() error {
	api := gin.Default()

	api.GET("/books", s.listBooks)

	return api.Run("0.0.0.0:8080")
}

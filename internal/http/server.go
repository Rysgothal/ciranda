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

	api.POST("/users", s.CreateUser)
	api.DELETE("/users/:id", s.DeleteUser)
	api.PATCH("/users/:id", s.UpdateUser)

	return api.Run("0.0.0.0:8080")
}

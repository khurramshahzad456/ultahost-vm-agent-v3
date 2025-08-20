package server

import (
	"fmt"
	"ultahost-ai-gateway/internal/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func NewServer() *Server {
	r := gin.Default()
	return &Server{Engine: r}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", config.AppConfig.Port)
	return s.Engine.Run(addr)
}

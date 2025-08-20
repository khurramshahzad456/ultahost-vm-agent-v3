package server

import (
	"ultahost-ai-gateway/internal/api"
	"ultahost-ai-gateway/internal/websocket"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {

	r.GET("/agent/connect", websocket.HandleAgentWebSocket)
	// r.POST("/agent/send-command", api.HandleSendCommand)
	// r.POST("/agent/register", api.HandleAgentRegister)
	r.POST("/agent/register", api.InstallTokenMiddleware(), api.HandleAgentRegister)
	r.Use(api.AuthMiddleware())

	r.POST("/chat", api.HandleChat)
	r.POST("/agent/enable", api.HandleEnableUltaAI)

}

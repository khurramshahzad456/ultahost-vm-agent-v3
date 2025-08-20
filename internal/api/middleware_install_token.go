// internal/api/middleware_install_token.go
package api

import (
	"net/http"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

// InstallTokenMiddleware checks install token validity
func InstallTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var body struct {
		// 	InstallToken string `json:"install_token" binding:"required"`
		// 	VpsId        string `json:"vps_id" binding:"required"`
		// }

		var body utils.TokenData

		// Parse JSON body first
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid install_token"})
			c.Abort()
			return
		}

		// Verify & consume token
		_, ok := utils.ConsumeInstallToken(body.Token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// fmt.Println("----vpsid: ", tokenData.VPSID)
		// Store data for handler use
		c.Set("tokenData", body)

		// Let the request proceed
		c.Next()
	}
}

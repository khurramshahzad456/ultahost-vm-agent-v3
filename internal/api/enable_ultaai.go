package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

type EnableUltaAIRequest struct {
	UserID string `json:"user_id" binding:"required"`
	VPSID  string `json:"vps_id" binding:"required"`
}

// Generate secure random token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func HandleEnableUltaAI(c *gin.Context) {
	var req EnableUltaAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := generateRandomToken(16)
	fmt.Println("========================")

	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate token")
		return
	}
	fmt.Println("------------------")

	utils.SaveInstallToken(token, req.UserID, req.VPSID, 15*time.Minute)

	curlCmd := fmt.Sprintf(
		`curl -s https://193.109.193.72/install.sh | bash -s -- --token=%s`,
		token,
	)
	c.String(http.StatusOK, curlCmd)
}

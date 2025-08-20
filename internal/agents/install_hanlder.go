package agents

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EnableUltaAIRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	VPSID  uint `json:"vps_id" binding:"required"`
}

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

	// Generate one-time token
	token, err := generateRandomToken(16)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return only curl command pointing to .sh file
	curlCmd := fmt.Sprintf(
		`curl -s https://install.ultaai.com/install.sh | bash -s -- --token=%s`,
		token,
	)

	c.String(http.StatusOK, curlCmd)
}

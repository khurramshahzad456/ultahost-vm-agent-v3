package api

import (
	"fmt"
	"net/http"
	"os"
	"ultahost-ai-gateway/internal/agents"
	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleChat(c *gin.Context) {
	var req *models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.UserToken = c.GetString("user_token")

	category, err := ai.ClassifyPromptCategory(&models.CategoryRequest{
		Query: req.Message,
		Categories: []string{
			"billing",
			"vps",
			"domain",
			"products",
			"support",
			"server_metrics",
			"vm_command",
			"unknown",
			"wordpress",
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI classifier failed", "details": err.Error()})
		return
	}
	fmt.Println(" Received token in", category, "agent:", req.UserToken)

	switch category {
	case "vps", "vm_command", "server_metrics", "wordpress":
		
		resp, err := agents.HandleVPS(req, agents.VPSFunctionList)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})

	case "billing":
		resp, err := agents.HandleBilling(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})

	case "domain":
		resp, err := agents.HandleDomain(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})

	case "products", "product_info", "hosting_plans":
		resp, err := agents.HandleProducts(req, agents.ProductsFunctionList)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})

	default:
		c.JSON(http.StatusNotImplemented, gin.H{"response": "I couldn’t process this request. Please rephrase or try again."})
	}
}

func InitAgent(c *gin.Context) {
	var req struct {
		InstallToken string `json:"install_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.InstallToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid install token"})
		return
	}

	agentID := "agent_" + uuid.NewString()

	signatureKey, err := utils.GenerateHMACKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate HMAC key"})
		return
	}

	identityToken, err := utils.GenerateJWTToken(agentID, []byte(os.Getenv("SIGNING_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	certPEM, privPEM := utils.GenerateSelfSignedCert(agentID)

	response := map[string]string{
		"identity_token": identityToken,
		"signature_key":  signatureKey,
		"certificate":    certPEM,
		"private_key":    privPEM,
	}

	c.JSON(http.StatusOK, response)
}

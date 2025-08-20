package agents

import (
	"strings"
	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
)

// List of available product-related functions

func HandleProducts(req *models.ChatRequest, functionList []string) (string, error) {
	functionName, err := ai.ClassifyFunctionWithinAgent(req.Message, functionList)
	if err != nil {
		return "", err
	}
	functionName = strings.ToLower(functionName)

	switch functionName {
	case "getallproducts":
		return getAllProducts(req)
	case "getproductpackage":
		return getProductPackage(req)
	case "getallpackages":
		return getAllPackages(req)
	default:
		return "I couldn't match your request to a known product function.", nil
	}
}

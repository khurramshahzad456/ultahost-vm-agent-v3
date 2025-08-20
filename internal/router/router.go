package router

// import (
// 	"errors"
// 	"ultahost-ai-gateway/internal/agents"
// 	"ultahost-ai-gateway/internal/pkg/models"
// )

// func RouteRequest(req models.ChatRequest) (string, error) {
// 	intent := ClassifyIntent(req.Message)

// 	switch intent {
// 	case "billing":
// 		return agents.HandleBilling(req)
// 	case "domain":
// 		return agents.HandleDomain(req)
// 	case "vps":
// 		return agents.HandleVPS(req)
// 	default:
// 		return "", errors.New("Unable to classify intent")
// 	}
// }

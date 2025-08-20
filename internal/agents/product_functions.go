package agents

import (
	"io"
	"net/http"
	"time"
	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
)

var ProductsFunctionList = []string{
	"getAllProducts",
	"getAllPackages",
	"getProductPackage",
	"getProductsByName",
	"getProductsByName",
}

// List of available product-related functions

func getAllProducts(req *models.ChatRequest) (string, error) {
	url := "https://api.ultahost.dev/products"

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.UserToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	rawOutput := string(body)
	summary, err := ai.SummarizeResponse(rawOutput)
	if err != nil {
		return rawOutput, nil
	}
	return summary, nil
}

// getAllPackages fetches all hosting packages and summarizes the response
func getAllPackages(req *models.ChatRequest) (string, error) {
	url := "https://api.ultahost.dev/products/all-package"

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.UserToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	rawOutput := string(body)
	summary, err := ai.SummarizeResponse(rawOutput)
	if err != nil {
		return rawOutput, nil
	}
	return summary, nil
}

// getProductPackage fetches a specific product-package combo and summarizes the response
func getProductPackage(req *models.ChatRequest) (string, error) {
	url := "https://api.ultahost.dev/products/package?product=dedicated-hosting&package=ulta-x3"

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.UserToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	rawOutput := string(body)
	summary, err := ai.SummarizeResponse(rawOutput)
	if err != nil {
		return rawOutput, nil
	}
	return summary, nil
}

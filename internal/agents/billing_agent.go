package agents

import (
	"ultahost-ai-gateway/internal/pkg/models"
)

func HandleBilling(req *models.ChatRequest) (string, error) {
	// url := "https://api.ultahost.dev/invoices"

	// payload := []byte(`{"query":"` + req.Message + `"}`)
	// httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	// if err != nil {
	// 	return "", err
	// }

	// httpReq.Header.Set("Authorization", "Bearer "+req.UserToken)
	// httpReq.Header.Set("Content-Type", "application/json")

	// client := &http.Client{Timeout: 10 * time.Second}
	// resp, err := client.Do(httpReq)
	// if err != nil {
	// 	return "", err
	// }
	// defer resp.Body.Close()

	// body, _ := io.ReadAll(resp.Body)
	// return string(body), nil
	return "✅ VPS agent called successfully with message: " + req.Message, nil
}

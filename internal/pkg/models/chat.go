package models

type ChatRequest struct {
	Message   string `json:"message"`
	UserToken string `json:"-"`
	VPSID     string   `json:"vps_id,omitempty"`   
    Args      []string `json:"args,omitempty"`
}

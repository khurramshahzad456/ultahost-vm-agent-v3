package utils

import (
	"sync"
	"time"
)

type TokenData struct {
	UserID string
	VPSID  string `json:"vps_id" binding:"required"`
	Token  string `json:"install_token" binding:"required"`
	Expiry time.Time
}

var (
	tokenStore   = make(map[string]TokenData)
	tokenStoreMu sync.Mutex
)

// SaveInstallToken saves token with TTL
func SaveInstallToken(token string, userID, vpsID string, ttl time.Duration) {
	tokenStoreMu.Lock()
	defer tokenStoreMu.Unlock()

	tokenStore[token] = TokenData{
		UserID: userID,
		VPSID:  vpsID,
		Expiry: time.Now().Add(ttl),
	}
}

// ConsumeInstallToken returns token data and deletes it
func ConsumeInstallToken(token string) (TokenData, bool) {
	tokenStoreMu.Lock()
	defer tokenStoreMu.Unlock()

	data, exists := tokenStore[token]
	if !exists || time.Now().After(data.Expiry) {
		// return TokenData{}, false
	}

	// Delete token so it can't be reused
	delete(tokenStore, token)
	return data, true
}

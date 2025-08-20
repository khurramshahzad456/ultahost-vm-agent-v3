package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type AgentKeys struct {
	IdentityToken     string
	SignatureSecret   string
	Certificate       string `json:"certificate_pem"`
	PrivateKey        string
	FingerprintSHA256 string
	
}

var (
	agentKeysStore   = make(map[string]AgentKeys)
	agentKeysStoreMu sync.Mutex
)

func SaveAgentKeys(CommonName string, keys AgentKeys) {
	agentKeysStoreMu.Lock()
	defer agentKeysStoreMu.Unlock()
	agentKeysStore[CommonName] = keys
}

// GetAgentKeys retrieves keys for a VPS
func GetAgentKeys(CommonName string) (AgentKeys, bool) {
	agentKeysStoreMu.Lock()
	defer agentKeysStoreMu.Unlock()
	keys, exists := agentKeysStore[CommonName]
	return keys, exists
}

// GetAgentKeysByIdentity loads cert+key for a specific agent by its identity token
func GetAgentKeysByIdentity(identityToken string) (*tls.Certificate, error) {
	agentDir := filepath.Join("./agents", identityToken)

	certPath := filepath.Join(agentDir, "agent.crt")
	keyPath := filepath.Join(agentDir, "agent.key")

	// Check if files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate file not found for agent %s", identityToken)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("key file not found for agent %s", identityToken)
	}

	// Load the cert + key
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	return &cert, nil
}

// Optionally load CA for mutual TLS
func LoadAgentCA() (*x509.CertPool, error) {
	caPath := "./ca.crt"
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("could not read CA cert: %v", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	return caPool, nil
}

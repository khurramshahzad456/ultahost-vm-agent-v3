package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

type AgentRegisterRequest struct {
	InstallToken string `json:"install_token" binding:"required"`
	VPSID        string `json:"vps_id" binding:"required"`
}

type AgentRegisterResponse struct {
	IdentityToken   string `json:"identity_token"`
	SignatureSecret string `json:"signature_secret"`
	Certificate     string `json:"certificate"` // base64 PEM cert
	PrivateKey      string `json:"private_key"` // base64 PEM key
}

func generateSecret(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSelfSignedCert() (certPEM []byte, keyPEM []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"UltaHost Agent"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM, nil
}

func HandleAgentRegister(c *gin.Context) {
	tokenData, exists := c.Get("tokenData")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token data missing"})
		return
	}
	td := tokenData.(utils.TokenData)

	fmt.Printf(" td: %+v \n", td)
	fmt.Printf(" TOKENDDATA: %+v \n", tokenData)

	identityToken := generateSecret(32)
	signatureSecret := generateSecret(32)
	keys, clientCertPEM, clientKeyPEM, err := ProceedCerts(td.VPSID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cert"})
		return
	}

	cert, err := GetCrt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cert"})
		return
	}
	fingerprint, err := utils.ComputeCertFingerprintSHA256(clientCertPEM)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fingerprint calc failed"})
		return
	}
	utils.SaveAgentKeys("Agent_"+td.VPSID, utils.AgentKeys{
		IdentityToken:     identityToken,
		SignatureSecret:   signatureSecret,
		Certificate:       base64.StdEncoding.EncodeToString(clientCertPEM),
		PrivateKey:        base64.StdEncoding.EncodeToString(clientKeyPEM),
		FingerprintSHA256: fingerprint,
	})

	keys["IdentityToken"] = identityToken
	keys["SignatureSecret"] = signatureSecret
	keys["FingerprintSHA256"] = fingerprint
	keys["Cert"] = string(cert)

	payload, err := json.Marshal(keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "json marshal failed"})
		return
	}

	encryptedPayload, err := encryptAESGCM(encryptionKey, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", encryptedPayload)
}

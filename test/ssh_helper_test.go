package test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func generateSshKeyPair(outputDir string, name string) error {
	err := os.RemoveAll(outputDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyFile, err := os.OpenFile(filepath.Join(outputDir, name+".pem"), os.O_CREATE|os.O_WRONLY, 0400)
	if err != nil {
		return err
	}
	err = pem.Encode(privateKeyFile, privateKeyBlock)
	if err != nil {
		return err
	}
	defer func(privatePem *os.File) {
		_ = privatePem.Close()
	}(privateKeyFile)

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err != nil {
		return err
	}
	publicKeyFile, err := os.OpenFile(filepath.Join(outputDir, name+".pub"), os.O_CREATE|os.O_WRONLY, 0444)
	if err != nil {
		return err
	}
	_, err = publicKeyFile.Write(publicKeyBytes)
	if err != nil {
		return err
	}
	defer func(publicPem *os.File) {
		_ = publicPem.Close()
	}(publicKeyFile)

	return nil
}

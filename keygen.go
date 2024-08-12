package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
)

// Generate and save keys for a given node ID
func generateAndSaveKeys(nodeID int) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}

	publicKeyPath := fmt.Sprintf("Keys/Node%d/Node%d_ED25519_PUB", nodeID, nodeID)
	privateKeyPath := fmt.Sprintf("Keys/Node%d/Node%d_ED25519_PRIV", nodeID, nodeID)

	// Ensure the directory exists
	os.MkdirAll(fmt.Sprintf("Keys/Node%d", nodeID), 0700)

	// Write public key
	if err := ioutil.WriteFile(publicKeyPath, publicKey, 0644); err != nil {
		fmt.Println("Error saving public key:", err)
	}

	// Write private key
	if err := ioutil.WriteFile(privateKeyPath, privateKey, 0644); err != nil {
		fmt.Println("Error saving private key:", err)
	}
}

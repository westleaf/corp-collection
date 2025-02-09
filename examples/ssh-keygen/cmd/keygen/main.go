package main

import (
	"fmt"
	"os"

	sshkeygen "github.com/westleaf/corp-collection/examples/ssh-keygen"
)

func main() {
	var (
		privateKey []byte
		publicKey  []byte
		err        error
	)
	if privateKey, publicKey, err = sshkeygen.GenerateKeys(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = os.WriteFile("somekey.pem", privateKey, 0600); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if err = os.WriteFile("somekey.pub", publicKey, 0644); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

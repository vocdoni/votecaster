package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

const (
	FID = 398983
)

func main() {
	http.HandleFunc("/", handleRequest)
	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Generate Ed25519 key pair
	_, privateKey, err := GenerateKeyPair()
	if err != nil {
		http.Error(w, "Error generating key pair", http.StatusInternalServerError)
		return
	}

	// Set a deadline for signing request (example purpose)
	deadline := GetDeadline()

	// Make API request
	_, deeplinkUrl, err := CreateSignedKeyRequest(privateKey, FID, deadline)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error making signed key request: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Deep link URL: %s\n", deeplinkUrl)

	// Generate the QR code
	qrCode, err := GenerateQRCode(deeplinkUrl)
	if err != nil {
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	// Display the information
	fmt.Fprintf(w, "<h1>Warpcast API Integration</h1>")
	fmt.Fprintf(w, "<p><strong>URL with Token:</strong> <a href='%s'>%s</a></p>", deeplinkUrl, deeplinkUrl)
	fmt.Fprintf(w, "<p><strong>Public Key:</strong> %x</p>", privateKey.Public())
	fmt.Fprintf(w, "<p><strong>Private Key:</strong> %x</p>", privateKey)
	fmt.Fprintf(w, "<p><strong>QR Code:</strong><br><img src='data:image/png;base64,%s'/></p>", base64.StdEncoding.EncodeToString(qrCode))
}

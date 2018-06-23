package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

// GenerateSecureID creates a secure ID to use
// as a CSRF key or a Message ID
func GenerateSecureID() string {
	// Inspired from gorilla/securecookie
	k := make([]byte, 32)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

// DomainHashFromAddress returns a byte slice containing the SHA1 hash
// of the domain in an email address.
func DomainHashFromAddress(addr string) (string, error) {
	// Parse out the email domain
	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		return "", err
	}
	parts := strings.Split(parsed.Address, "@")
	if len(parts) != 2 {
		return "", err
	}
	domain := parts[1]
	// Generate the hash of the domain provided
	domainHash := sha1.Sum([]byte(domain))
	return hex.EncodeToString(domainHash[:]), nil
}

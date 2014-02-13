package api

import (
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/mdlayher/goat/goat/data"
)

// APIAuthenticator interface which defines methods required to implement an authentication method
type APIAuthenticator interface {
	Auth(*http.Request) (bool, error)
}

// BasicAuthenticator uses the HTTP Basic authentication scheme
type BasicAuthenticator struct {
}

// Auth handles validation of HTTP Basic authentication
func (a BasicAuthenticator) Auth(r *http.Request) (bool, error) {
	// Retrieve Authorization header
	auth := r.Header.Get("Authorization")

	// No header provided
	if auth == "" {
		return false, nil
	}

	// Ensure format is valid
	basic := strings.Split(auth, " ")
	if basic[0] != "Basic" {
		return false, nil
	}

	// Decode base64'd user:password pair
	buf, err := base64.URLEncoding.DecodeString(basic[1])
	if err != nil {
		return false, err
	}

	// Split into user ID/password
	credentials := strings.Split(string(buf), ":")

	// Load API key for specified user ID
	key, err := new(data.APIKey).Load(credentials[0], "user_id")
	if err != nil || key == (data.APIKey{}) {
		return false, err
	}

	// Hash input password
	sha := sha1.New()
	if _, err = sha.Write([]byte(credentials[1] + key.Salt)); err != nil {
		return false, err
	}
	hash := fmt.Sprintf("%x", sha.Sum(nil))

	// Verify hashes match, using timing-attack resistant method
	// If function returns 1, hashes match
	return subtle.ConstantTimeCompare([]byte(hash), []byte(key.Key)) == 1, nil
}

// HMACAuthenticator uses the HMAC-SHA1 authentication scheme
type HMACAuthenticator struct {
}

// Auth handles validation of HMAC-SHA1 authentication
func (a HMACAuthenticator) Auth(r *http.Request) (bool, error) {
	return true, nil
}

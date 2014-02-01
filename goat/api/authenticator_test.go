package api

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/mdlayher/goat/goat/common"
	"github.com/mdlayher/goat/goat/data"
)

// TestBasicAuthenticator verifies that BasicAuthenticator.Auth works properly
func TestBasicAuthenticator(t *testing.T) {
	log.Println("TestBasicAuthenticator()")

	// Load config
	config := common.LoadConfig()
	common.Static.Config = config

	// Generate mock user
	user := data.UserRecord{
		Username:     "test",
		Passkey:      "abcdef0123456789",
		TorrentLimit: 10,
	}

	// Save mock user
	if !user.Save() {
		t.Fatalf("Failed to save mock user")
	}

	// Load mock user to fetch ID
	user = user.Load(user.Username, "username")
	if user == (data.UserRecord{}) {
		t.Fatalf("Failed to load mock user")
	}

	// Generate an API key and salt
	pass := "deadbeef"
	salt := "test"

	sha := sha1.New()
	sha.Write([]byte(pass + salt))
	hash := fmt.Sprintf("%x", sha.Sum(nil))

	// Generate mock API key
	key := data.APIKey{
		UserID: user.ID,
		Key:    hash,
		Salt:   salt,
	}

	// Save mock API key
	if !key.Save() {
		t.Fatalf("Failed to save mock data.APIKey")
	}

	// Load mock data.APIKey to fetch ID
	key = key.Load(key.Key, "key")
	if key == (data.APIKey{}) {
		t.Fatalf("Failed to load mock data.APIKey")
	}

	// Generate mock HTTP request
	r := http.Request{}
	headers := map[string][]string{
		"Authorization": {"Basic " + base64.URLEncoding.EncodeToString([]byte(user.Username+":"+pass))},
	}
	r.Header = headers

	// Perform authentication request
	auth := new(BasicAuthenticator).Auth(&r)
	if !auth {
		t.Fatalf("Failed to authenticate using BasicAuthenticator")
	}

	// Delete mock user
	if !user.Delete() {
		t.Fatalf("Failed to delete mock user")
	}

	// Delete mock API key
	if !key.Delete() {
		t.Fatalf("Failed to delete mock data.APIKey")
	}
}

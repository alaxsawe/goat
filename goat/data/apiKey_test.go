package data

import (
	"log"
	"testing"

	"github.com/mdlayher/goat/goat/common"
)

// TestAPIKey verifies that APIKey creation, save, load, and delete work properly
func TestAPIKey(t *testing.T) {
	log.Println("TestAPIKey()")

	// Load config
	config, err := common.LoadConfig()
	if err != nil {
		t.Fatalf("Could not load configuration: %s", err.Error())
	}
	common.Static.Config = config

	// Generate mock APIKey
	key := new(APIKey)
	if err := key.Create(1); err != nil {
		t.Fatalf("Failed to create mock APIKey: %s", err.Error())
	}

	// Verify key can be saved
	if err := key.Save(); err != nil {
		t.Fatalf("Failed to save APIKey: %s", err.Error())
	}

	// Verify key can be loaded using matching pubkey
	key2, err := key.Load(key.Pubkey, "pubkey")
	if err != nil || key2 == (APIKey{}) {
		t.Fatal("Failed to load APIKey: %s", err.Error())
	}

	// Verify key is the same as previous one
	if key.Pubkey != key2.Pubkey {
		t.Fatalf("key.Pubkey, expected %s, got %s", key.Pubkey, key2.Pubkey)
	}

	// Verify key can be deleted
	if err := key2.Delete(); err != nil {
		t.Fatalf("Failed to delete APIKey: %s", err.Error())
	}
}

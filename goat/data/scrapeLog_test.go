package data

import (
	"log"
	"net/url"
	"testing"

	"github.com/mdlayher/goat/goat/common"
)

// TestScrapeLog verifies that scrape log creation, save, load, and delete work properly
func TestScrapeLog(t *testing.T) {
	log.Println("TestScrapeLog()")

	// Load config
	config, err := common.LoadConfig()
	if err != nil {
		t.Fatalf("Could not load configuration: %s", err.Error())
	}
	common.Static.Config = config

	// Generate fake scrape query
	query := url.Values{}
	query.Set("info_hash", "deadbeef000000000000")
	query.Set("ip", "127.0.0.1")

	// Generate struct from query
	scrape := new(ScrapeLog)
	err = scrape.FromValues(query)
	if err != nil {
		t.Fatalf("Failed to create scrape from values: %s", err.Error())
	}

	// Verify proper hex encode of info hash
	if scrape.InfoHash != "6465616462656566303030303030303030303030" {
		t.Fatalf("InfoHash, expected \"6465616462656566303030303030303030303030\", got %s", scrape.InfoHash)
	}

	// Verify same IP
	if scrape.IP != "127.0.0.1" {
		t.Fatalf("IP, expected \"127.0.0.1\", got %s", scrape.IP)
	}

	// Verify scrape can be saved
	if err := scrape.Save(); err != nil {
		t.Fatalf("Failed to save ScrapeLog: %s", err.Error())
	}

	// Verify scrape can be loaded using hex info hash
	scrape2, err := scrape.Load("6465616462656566303030303030303030303030", "info_hash")
	if scrape2 == (ScrapeLog{}) || err != nil {
		t.Fatal("Failed to load ScrapeLog: %s", err.Error())
	}

	// Verify scrape is the same as previous one
	if scrape.InfoHash != scrape2.InfoHash {
		t.Fatalf("scrape.InfoHash, expected %s, got %s", scrape.InfoHash, scrape2.InfoHash)
	}

	// Verify scrape can be deleted
	if err := scrape2.Delete(); err != nil {
		t.Fatalf("Failed to delete ScrapeLog: %s", err.Error())
	}
}

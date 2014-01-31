package goat

import (
	"log"
	"net/url"
	"testing"
)

// TestUDPTrackerAnnounce verifies that the UDP tracker announce output format is correct
func TestUDPTrackerAnnounce(t *testing.T) {
	log.Println("TestUDPTrackerAnnounce()")

	// Generate mock fileRecord
	file := fileRecord{
		InfoHash: "6465616462656566",
		Verified: true,
	}

	// Save mock file
	if !file.Save() {
		t.Fatalf("Failed to save mock file")
	}

	// Generate fake announce query
	query := url.Values{}
	query.Set("info_hash", "deadbeef")
	query.Set("ip", "127.0.0.1")
	query.Set("port", "5000")
	query.Set("uploaded", "0")
	query.Set("downloaded", "0")
	query.Set("left", "0")
	query.Set("numwant", "50")

	// Create a UDP tracker, trigger an announce
	tracker := udpTracker{TransID: []byte("0123")}
	res := tracker.Announce(query, file)

	// Decode response
	announce, err := new(udpAnnounceResponsePacket).FromBytes(res)
	if err != nil {
		t.Fatalf("Failed to decode UDP announce response")
	}
	log.Println(announce)

	// Verify correct action
	if announce.Action != 1 {
		t.Fatalf("Incorrect UDP action, expected 1")
	}

	// Delete mock file
	if !file.Delete() {
		t.Fatalf("Failed to delete mock file")
	}
}

// TestUDPTrackerError verifies that the UDP tracker error format is correct
func TestUDPTrackerError(t *testing.T) {
	log.Println("TestUDPTrackerError()")

	// Create a UDP tracker, trigger an error
	tracker := udpTracker{TransID: []byte("0123")}
	res := tracker.Error("Testing")

	// Decode response
	errRes, err := new(udpErrorResponsePacket).FromBytes(res)
	if err != nil {
		t.Fatalf("Failed to decode UDP error response")
	}
	log.Println(errRes)

	// Verify correct action
	if errRes.Action != 3 {
		t.Fatalf("Incorrect UDP action, expected 3")
	}

	// Verify correct error
	if errRes.Error != "Testing" {
		t.Fatalf("Incorrect UDP error, expected 'Testing'")
	}
}

// TestUDPTrackerScrape verifies that the UDP tracker scrape format is correct
func TestUDPTrackerScrape(t *testing.T) {
	log.Println("TestUDPTrackerScrape()")

	// Generate mock fileRecord
	file := fileRecord{
		InfoHash: "6465616462656566",
		Verified: true,
	}

	// Save mock file
	if !file.Save() {
		t.Fatalf("Failed to save mock file")
	}

	// Store file in slice
	files := make([]fileRecord, 0)
	files = append(files[:], file)

	// Create a UDP tracker, trigger a scrape
	tracker := udpTracker{TransID: []byte("0123")}
	res := tracker.Scrape(files)
	log.Println(string(res))

	// Decode response
	scrape, err := new(udpScrapeResponsePacket).FromBytes(res)
	if err != nil {
		t.Fatalf("Failed to decode UDP scrape response")
	}
	log.Println(scrape)

	// Verify correct action
	if scrape.Action != 2 {
		t.Fatalf("Incorrect UDP action, expected 2")
	}

	// Delete mock file
	if !file.Delete() {
		t.Fatalf("Failed to delete mock file")
	}
}

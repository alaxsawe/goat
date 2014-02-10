package udp

import (
	"bytes"
	"log"
	"testing"
)

// TestScrapeRequest verifies that ScrapeRequest binary marshal and unmarshal work properly
func TestScrapeRequest(t *testing.T) {
	log.Println("TestScrapeRequest()")

	// Generate mock ScrapeRequest
	scrape := ScrapeRequest{
		ConnID: 1,
		Action: 2,
		TransID: 1,
		InfoHashes: [][]byte{[]byte("abcdef0123456789abcd"), []byte("0123456789abcdef0123")},
	}

	// Marshal to binary representation
	out, err := scrape.MarshalBinary()
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeRequest to binary: %s", err.Error())
	}

	// Unmarshal scrape from binary representation
	scrape2 := new(ScrapeRequest)
	if err := scrape2.UnmarshalBinary(out); err != nil {
		t.Fatalf("Failed to unmarshal ScrapeRequest from binary: %s", err.Error())
	}

	// Verify scrapes are identical
	if scrape.ConnID != scrape2.ConnID || !bytes.Equal(scrape.InfoHashes[0], scrape2.InfoHashes[0]) {
		t.Fatalf("ScrapeRequest results do not match")
	}
}

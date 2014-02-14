package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mdlayher/goat/goat/data"
)

// Table driven tests to iterate over and test the API router
var apiTests = []struct {
	method string
	url    string
	code   int
}{
	{"GET", "/api/", 404},
	{"GET", "/api/files/a", 400},
	{"GET", "/api/abcdef", 404},
	{"GET", "/api/files", 200},
	{"GET", "/api/files/1", 200},
	{"GET", "/api/status", 200},
	{"GET", "/api/users", 200},
	{"GET", "/api/users/1", 200},
	{"PUT", "/api/", 405},
}

// TestRouter verifies that the API router is working properly
func TestRouter(t *testing.T) {
	log.Println("TestRouter()")

	// Iterate all API tests
	for _, test := range apiTests {
		// Generate mock HTTP request
		r, err := http.NewRequest(test.method, "http://localhost:8080"+test.url, nil)
		if err != nil {
			t.Fatalf("Failed to create HTTP request")
		}

		// Capture HTTP writer response with recorder
		w := httptest.NewRecorder()

		// Invoke API router
		Router(w, r, data.UserRecord{})

		// Validate input
		if w.Code != test.code {
			t.Fatalf("Test %s %s, expected HTTP %d, got HTTP %d", test.method, test.url, test.code, w.Code)
		}

		log.Printf("OK - %s %s -> HTTP %d", test.method, test.url, w.Code)
		log.Printf(w.Body.String())
	}
}

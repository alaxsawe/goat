package goat

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// apiError represents an error response from the API
type apiError struct {
	Error string `json:"error"`
}

// apiRouter handles the routing of HTTP API requests
func apiRouter(w http.ResponseWriter, r *http.Request) {
	// API is read-only, at least for the time being
	if r.Method != "GET" {
		http.Error(w, string(apiErrorResponse("Method not allowed")), 405)
		return
	}

	// Log API calls
	log.Printf("API: [http %s] %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)

	// Split request path
	urlArr := strings.Split(r.URL.Path, "/")

	// Verify API method set
	if len(urlArr) < 3 {
		http.Error(w, string(apiErrorResponse("No API call")), 404)
		return
	}

	// Check for an ID
	ID := -1
	if len(urlArr) == 4 {
		i, err := strconv.Atoi(urlArr[3])
		if err != nil || i < 1 {
			http.Error(w, string(apiErrorResponse("Invalid integer ID")), 400)
			return
		} else {
			ID = i
		}
	}

	// Response buffer
	res := make([]byte, 0)

	// Choose API method
	switch urlArr[2] {
	// Files on tracker
	case "files":
		res = getFilesJSON(ID)
	// Server status
	case "status":
		res = getStatusJSON()
	// Return error response
	default:
		http.Error(w, string(apiErrorResponse("Undefined API call")), 404)
		return
	}

	// If requested, compress response using gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Add("Content-Encoding", "gzip")

		// Write gzip'd response
		gz := gzip.NewWriter(w)
		if _, err := gz.Write(res); err != nil {
			log.Println(err.Error())
			return
		}

		if err := gz.Close(); err != nil {
			log.Println(err.Error())
		}

		return
	}

	if _, err := w.Write(res); err != nil {
		log.Println(err.Error())
	}

	return
}

// apiErrorResponse returns an apiError as JSON
func apiErrorResponse(msg string) []byte {
	res := apiError{
		msg,
	}

	out, err := json.Marshal(res)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return out
}

package goat

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
)

// Handshake for UDP tracker protocol
const udpInitID = 4497486125440

// UDP errors
var (
	// udpActionError is returned when a client requests an invalid tracker action
	udpActionError = errors.New("udp: client did not send a valid UDP tracker action")
	// udpHandshakeError is returned when a client does not send the proper handshake ID
	udpHandshakeError = errors.New("udp: client did not send proper UDP tracker handshake")
	// udpIntegerError is returned when a client sends an invalid integer parameter
	udpIntegerError = errors.New("udp: client sent an invalid integer parameter")
	// udpWriteError is returned when the tracker cannot generate a proper response
	udpWriteError = errors.New("udp: tracker cannot generate UDP tracker response")
)

// UDP address to connection ID map
var udpAddrToID map[string]uint64 = map[string]uint64{}

// Handle incoming UDP connections and return response
func handleUDP(l *net.UDPConn, sendChan chan bool, recvChan chan bool) {
	// Create shutdown function
	go func(l *net.UDPConn, sendChan chan bool, recvChan chan bool) {
		// Wait for done signal
		<-sendChan

		// Close listener
		if err := l.Close(); err != nil {
			log.Println(err.Error())
		}

		log.Println("UDP listener stopped")
		recvChan <- true
	}(l, sendChan, recvChan)

	// Loop and read connections
	for {
		buf := make([]byte, 2048)
		rlen, addr, err := l.ReadFromUDP(buf)

		// Count incoming connections
		atomic.AddInt64(&static.UDP.Current, 1)
		atomic.AddInt64(&static.UDP.Total, 1)

		// Triggered on graceful shutdown
		if err != nil {
			// Ignore connection closing error, caused by stopping network listener
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Println(err.Error())
				panic(err)
			}

			return
		}

		// Verify length is at least 16 bytes
		if rlen < 16 {
			log.Println("Invalid length")
			continue
		}

		// Spawn a goroutine to handle the connection and send back the response
		go func(l *net.UDPConn, buf []byte, addr *net.UDPAddr) {
			// Capture initial response from buffer
			res, err := parseUDP(buf, addr)
			if err != nil {
				// Client sent a malformed UDP handshake
				log.Println(err.Error())

				// If error, client did not handshake correctly, so boot them with error message
				_, err2 := l.WriteToUDP(res, addr)
				if err2 != nil {
					log.Println(err2.Error())
				}

				return
			}

			// Write response
			_, err = l.WriteToUDP(res, addr)
			if err != nil {
				log.Println(err.Error())
			}

			return
		}(l, buf, addr)
	}
}

// Parse a UDP byte buffer, return response from tracker
func parseUDP(buf []byte, addr *net.UDPAddr) ([]byte, error) {
	// Current connection ID (initially handshake, then generated by tracker)
	connID := binary.BigEndian.Uint64(buf[0:8])
	// Action integer (connect: 0, announce: 1)
	action := binary.BigEndian.Uint32(buf[8:12])
	// Transaction ID, to match between requests
	transID := buf[12:16]

	// Action switch
	switch action {
	// Connect
	case 0:
		// Validate UDP tracker handshake
		if connID != udpInitID {
			return udpTrackerError("Invalid UDP tracker handshake", transID), udpHandshakeError
		}

		res := bytes.NewBuffer(make([]byte, 0))

		// Action
		err := binary.Write(res, binary.BigEndian, uint32(0))
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Could not generate UDP tracker response", transID), udpWriteError
		}

		// Transaction ID
		err = binary.Write(res, binary.BigEndian, transID)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Could not generate UDP tracker response", transID), udpWriteError
		}

		// Generate a connection ID, which will be expected for this client next call
		expID := uint64(randRange(1, 1000000000))

		// Store this client's address and ID in map
		udpAddrToID[addr.String()] = expID

		// Connection ID, generated for this session
		err = binary.Write(res, binary.BigEndian, expID)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Could not generate UDP tracker response", transID), udpWriteError
		}

		return res.Bytes(), nil
	// Announce
	case 1:
		// Ensure connection ID map contains this IP address
		expID, ok := udpAddrToID[addr.String()];
		if !ok {
			return udpTrackerError("Client must properly handshake before announce", transID), udpHandshakeError
		}

		// Validate expected connection ID using map
		if connID != expID {
			return udpTrackerError("Invalid UDP connection ID", transID), udpHandshakeError
		}

		// Clear this IP from the connection map
		delete(udpAddrToID, addr.String())

		// Generate connection query
		query := url.Values{}

		// Mark client as UDP
		query.Set("udp", "1")

		// Transaction ID
		transID := buf[12:16]

		// Info hash
		query.Set("info_hash", string(buf[16:36]))

		// Skipped: peer_id: buf[36:56]

		// Downloaded
		t, err := strconv.ParseInt(hex.EncodeToString(buf[56:64]), 16, 64)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: downloaded", transID), udpIntegerError
		}
		query.Set("downloaded", strconv.FormatInt(t, 10))

		// Left
		t, err = strconv.ParseInt(hex.EncodeToString(buf[64:72]), 16, 64)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: left", transID), udpIntegerError
		}
		query.Set("left", strconv.FormatInt(t, 10))

		// Uploaded
		t, err = strconv.ParseInt(hex.EncodeToString(buf[72:80]), 16, 64)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: uploaded", transID), udpIntegerError
		}
		query.Set("uploaded", strconv.FormatInt(t, 10))

		// Event
		t, err = strconv.ParseInt(hex.EncodeToString(buf[80:84]), 16, 32)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: event", transID), udpIntegerError
		}
		event := strconv.FormatInt(t, 10)

		// Convert event to actual string
		switch event {
		case "0":
			query.Set("event", "")
		case "1":
			query.Set("event", "completed")
		case "2":
			query.Set("event", "started")
		case "3":
			query.Set("event", "stopped")
		}

		// IP address
		t, err = strconv.ParseInt(hex.EncodeToString(buf[84:88]), 16, 32)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: ip", transID), udpIntegerError
		}
		query.Set("ip", strconv.FormatInt(t, 10))

		// If no IP address set, use the UDP source
		if query.Get("ip") == "0" {
			query.Set("ip", strings.Split(addr.String(), ":")[0])
		}

		// Key
		query.Set("key", hex.EncodeToString(buf[88:92]))

		// Numwant
		query.Set("numwant", hex.EncodeToString(buf[92:96]))

		// If numwant is hex max value, default to 50
		if query.Get("numwant") == "ffffffff" {
			query.Set("numwant", "50")
		}

		// Port
		t, err = strconv.ParseInt(hex.EncodeToString(buf[96:98]), 16, 32)
		if err != nil {
			log.Println(err.Error())
			return udpTrackerError("Invalid integer parameter: port", transID), udpIntegerError
		}
		query.Set("port", strconv.FormatInt(t, 10))

		// Trigger an anonymous announce
		return trackerAnnounce(userRecord{}, query, transID), nil
	default:
		return udpTrackerError("Invalid action", transID), udpActionError
	}
}

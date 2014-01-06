package goat

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

// Handshake for UDP tracker protocol
const InitID = 4497486125440

// UDPConnHandler handles incoming UDP network connections
type UDPConnHandler struct {
}

// Handle incoming UDP connections and return response
func (u UDPConnHandler) Handle(l *net.UDPConn, udpDoneChan chan bool) {
	// Create shutdown function
	go func(l *net.UDPConn, udpDoneChan chan bool) {
		// Wait for done signal
		Static.ShutdownChan <- <-Static.ShutdownChan

		// Close listener
		l.Close()
		udpDoneChan <- true
	}(l, udpDoneChan)

	first := true
	for {
		buf := make([]byte, 2048)
		rlen, addr, err := l.ReadFromUDP(buf)

		// Triggered on graceful shutdown
		if err != nil {
			return
		}

		// Verify length is at least 16 bytes
		if rlen < 16 {
			Static.LogChan <- "Invalid length"
			continue
		}

		// Current connection ID (initially handshake, then generated by tracker)
		connID := binary.BigEndian.Uint64(buf[0:8])
		// Action integer (connect: 0, announce: 1)
		action := binary.BigEndian.Uint32(buf[8:12])
		// Transaction ID, to match between requests
		transID := buf[12:16]

		// On first run, verify valid connection ID
		if first {
			if connID != InitID {
				Static.LogChan <- "Invalid connection handshake"
				_, err = l.WriteToUDP(UDPTrackerError("Invalid connection handshake", transID), addr)
				if err != nil {
					Static.LogChan <- err.Error()
					return
				}
				continue
			}
			first = false
		}

		// Action switch
		switch action {
		// Connect
		case 0:
			res := bytes.NewBuffer(make([]byte, 0))

			// Action
			err = binary.Write(res, binary.BigEndian, uint32(0))
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}

			// Transaction ID
			err = binary.Write(res, binary.BigEndian, transID)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}

			// Connection ID, generated for this session
			err = binary.Write(res, binary.BigEndian, uint64(RandRange(0, 1000000000)))
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}

			_, err := l.WriteToUDP(res.Bytes(), addr)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}

			continue
		// Announce
		case 1:
			query := map[string]string{}

			// Ignoring these for now, because clients function sanely without them
			// Connection ID: buf[0:8]
			// Action: buf[8:12]

			// Mark client as UDP
			query["udp"] = "1"

			// Transaction ID
			transID := buf[12:16]

			// Info hash
			query["info_hash"] = string(buf[16:36])

			// Skipped: peer_id: buf[36:56]

			// Downloaded
			t, err := strconv.ParseInt(hex.EncodeToString(buf[56:64]), 16, 64)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["downloaded"] = strconv.FormatInt(t, 10)

			// Left
			t, err = strconv.ParseInt(hex.EncodeToString(buf[64:72]), 16, 64)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["left"] = strconv.FormatInt(t, 10)

			// Uploaded
			t, err = strconv.ParseInt(hex.EncodeToString(buf[72:80]), 16, 64)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["uploaded"] = strconv.FormatInt(t, 10)

			// Event
			t, err = strconv.ParseInt(hex.EncodeToString(buf[80:84]), 16, 32)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["event"] = strconv.FormatInt(t, 10)

			// Convert event to actual string
			switch query["event"] {
			case "0":
				query["event"] = ""
			case "1":
				query["event"] = "completed"
			case "2":
				query["event"] = "started"
			case "3":
				query["event"] = "stopped"
			}

			// IP address
			t, err = strconv.ParseInt(hex.EncodeToString(buf[84:88]), 16, 32)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["ip"] = strconv.FormatInt(t, 10)

			// If no IP address set, use the UDP source
			if query["ip"] == "0" {
				query["ip"] = strings.Split(addr.String(), ":")[0]
			}

			// Key
			query["key"] = hex.EncodeToString(buf[88:92])

			// Numwant
			query["numwant"] = hex.EncodeToString(buf[92:96])

			// If numwant is hex max value, default to 50
			if query["numwant"] == "ffffffff" {
				query["numwant"] = "50"
			}

			// Port
			t, err = strconv.ParseInt(hex.EncodeToString(buf[96:98]), 16, 32)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
			query["port"] = strconv.FormatInt(t, 10)

			// Trigger an anonymous announce
			resChan := make(chan []byte)
			go TrackerAnnounce(UserRecord{}, query, transID, resChan)

			_, err = l.WriteToUDP(<-resChan, addr)
			close(resChan)
			if err != nil {
				Static.LogChan <- err.Error()
				return
			}
		default:
			Static.LogChan <- "Invalid action"
			continue
		}
	}
}

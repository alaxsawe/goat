package api

import (
	"encoding/json"

	"github.com/mdlayher/goat/goat/data"
)

// getUsersJSON returns a JSON representation of one or more data.UserRecords
func getUsersJSON(ID int) ([]byte, error) {
	// Check for a valid integer ID
	if ID > 0 {
		// Load user
		user, err := new(data.UserRecord).Load(ID, "id")
		if err != nil {
			return nil, err
		}

		// Create JSON represenation
		jsonUser, err := user.ToJSON()
		if err != nil {
			return nil, err
		}

		// Marshal into JSON
		res, err := json.Marshal(jsonUser)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	// Load all users
	users, err := new(data.UserRecordRepository).All()
	if err != nil {
		return nil, err
	}

	// Convert all users to JSON representation
	jsonUsers := make([]data.JSONUserRecord, 0)
	for _, u := range users {
		j, err := u.ToJSON()
		if err != nil {
			return nil, err
		}

		jsonUsers = append(jsonUsers[:], j)
	}

	// Marshal into JSON
	res, err := json.Marshal(jsonUsers)
	if err != nil {
		return nil, err
	}

	return res, err
}
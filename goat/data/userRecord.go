package data

import (
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/mdlayher/goat/goat/common"
)

// UserRecord represents a user on the tracker
type UserRecord struct {
	ID           int
	Username     string
	Passkey      string
	TorrentLimit int `db:"torrent_limit"`
}

// Create a UserRecord, using defined parameters
func (u *UserRecord) Create(username string, torrentLimit int) error {
	// Set username and torrent limit
	u.Username = username
	u.TorrentLimit = torrentLimit

	// Randomly generate a new passkey
	sha := sha1.New()
	if _, err := sha.Write([]byte(common.RandString())); err != nil {
		return err
	}

	u.Passkey = fmt.Sprintf("%x", sha.Sum(nil))
	return nil
}

// Delete UserRecord from storage
func (u UserRecord) Delete() bool {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// Delete UserRecord
	if err = db.DeleteUserRecord(u.ID, "id"); err != nil {
		log.Println(err.Error())
		return false
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return true
}

// Save UserRecord to storage
func (u UserRecord) Save() bool {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// Save UserRecord
	if err := db.SaveUserRecord(u); err != nil {
		log.Println(err.Error())
		return false
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return true
}

// Load UserRecord from storage
func (u UserRecord) Load(id interface{}, col string) UserRecord {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		log.Println(err.Error())
		return u
	}

	// Load UserRecord by specified column
	u, err = db.LoadUserRecord(id, col)
	if err != nil {
		log.Println(err.Error())
		return UserRecord{}
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return u
}

// Uploaded loads this user's total upload
func (u UserRecord) Uploaded() (int64, error) {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		return 0, err
	}

	// Retrieve total bytes user has uploaded
	uploaded, err := db.GetUserUploaded(u.ID)
	if err != nil {
		return 0, err
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return uploaded, nil
}

// Downloaded loads this user's total download
func (u UserRecord) Downloaded() (int64, error) {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		return 0, err
	}

	// Retrieve total bytes user has downloaded
	downloaded, err := db.GetUserDownloaded(u.ID)
	if err != nil {
		return 0, err
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return downloaded, nil
}

// Seeding counts the number of torrents this user is seeding
func (u UserRecord) Seeding() (int, error) {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		return 0, err
	}

	// Retrieve total number of torrents user is actively seeding
	seeding, err := db.GetUserSeeding(u.ID)
	if err != nil {
		return 0, err
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return seeding, nil
}

// Leeching counts the number of torrents this user is leeching
func (u UserRecord) Leeching() (int, error) {
	// Open database connection
	db, err := DBConnect()
	if err != nil {
		return 0, err
	}

	// Retrieve total number of torrents user is actively leeching
	leeching, err := db.GetUserSeeding(u.ID)
	if err != nil {
		return 0, err
	}

	if err := db.Close(); err != nil {
		log.Println(err.Error())
	}

	return leeching, nil
}

package goat

// Generates a unique hash for a given struct with its ID, to be used as ID for storage
type Hashable interface {
	Hash() string
}

// Defines methods to save, load, and delete data for an object
type Persistent interface {
	Save() bool
	Load(interface{}, string) interface{}
	Delete() bool
}

type ErrorRes struct {
	ErrLocation string
	Time        int64
	Error       string
}

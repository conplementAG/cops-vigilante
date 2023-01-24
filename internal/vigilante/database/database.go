package database

// Database is a simple interface for any key-value store
type Database interface {
	// Set sets the key with value, creating a new key or overwriting an existing one
	Set(key string, value interface{})

	// Get returns the value found by key. Returns nil if nothing found
	Get(key string) interface{}
}

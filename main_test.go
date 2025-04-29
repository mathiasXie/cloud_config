package cloud_config

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Call Init to initialize the database and start the refreshConfig goroutine
	Init(db)

	err = SaveConfig("server", "web server config", "{\"host\":\"localhost\"}", "web server config host,port and other config")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	host := GetConfig("server")["host"]
	if host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", host)
	}
	for {
	}
}

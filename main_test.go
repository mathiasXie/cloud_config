package cloud_config

import (
	"fmt"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestInit(t *testing.T) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local", "mathias", "123456", "192.168.6.109", "30306", "spring")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Call Init to initialize the database and start the refreshConfig goroutine
	Init(db, "llm-rpc")

	err = SaveConfig("server4", "web server config", "{\"host\":\"localhost\"}", "web server config host,port and other config")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	} else {
		t.Logf("save config success")
	}
	config := GetConfig("server4")
	t.Logf("config: %+v\n", config)
	host := config["host"]
	if host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", host)
	} else {
		t.Logf("host: %s", host)
	}
	RemoveConfig("server4")
}

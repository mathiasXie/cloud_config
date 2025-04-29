package cloud_config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type cloudConfig map[string]string

var (
	db         *gorm.DB
	configMap  = make(map[string]cloudConfig)
	configLock sync.RWMutex
)

func Init(configDB *gorm.DB) {
	db = configDB

	// Check if the table exists. If it doesn't exist, create it.
	if !db.Migrator().HasTable(&CloudConfig{}) {
		if err := db.Migrator().CreateTable(&CloudConfig{}); err != nil {
			log.Fatalf("Failed to create cloud_configs table: %v", err)
		}
		log.Println("cloud_configs table created")
	}

	// Timed refresh configuration
	go refreshConfig()
}

func loadConfigFromDB() {
	var configs []CloudConfig
	result := db.Where("deleted_at IS NULL").Find(&configs)
	if result.Error != nil {
		log.Fatalf("Failed to query cloud_configs table: %v", result.Error)
	}

	configLock.Lock()
	defer configLock.Unlock()

	for _, config := range configs {
		var cfg map[string]string
		err := json.Unmarshal([]byte(config.ConfigValue), &cfg)
		if err != nil {
			log.Printf("config %s can not marshal", config.ConfigKey)
			continue
		}
		configMap[config.ConfigKey] = cfg
	}

	log.Printf("Loaded %d cloud configs from the database", len(configMap))
}

func refreshConfig() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		loadConfigFromDB()
	}
}

func GetConfig(key string) map[string]string {
	configLock.RLock()
	defer configLock.RUnlock()

	config, ok := configMap[key]
	if !ok {
		fmt.Printf("Config with key '%s' not found\n", key)
		return map[string]string{}
	}

	return config
}

func SaveConfig(key string, data string) error {
	configLock.Lock()
	defer configLock.Unlock()

	var cfg map[string]string
	err := json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Printf("config %s can not marshal", key)
		return err
	}

	cfgModel := &CloudConfig{}
	cfgModel.ConfigKey = key
	cfgModel.ConfigValue = data
	result := db.Model(&CloudConfig{}).Save(cfgModel)
	if result.Error != nil {
		return result.Error
	}

	configMap[key] = cfg
	return nil
}

func RemoveConfig(key string) {
	configLock.Lock()
	defer configLock.Unlock()

	// Soft delete config
	cfgModel := &CloudConfig{}
	cfgModel.ConfigKey = key
	cfgModel.DeleteAt = time.Now()
	result := db.Model(&CloudConfig{}).Save(&cfgModel)
	if result.Error != nil {
		log.Fatalf("Failed to delete config in the database: %v", result.Error)
	}
	delete(configMap, key)
}

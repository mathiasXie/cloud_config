package cloud_config

import (
	"encoding/json"
	"errors"
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
	namespace  string
)

func Init(configDB *gorm.DB, configNamespace string) {
	db = configDB
	namespace = configNamespace
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
	result := db.Where("deleted_at IS NULL AND namespace=?", namespace).Find(&configs)
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

func SaveConfig(key, name, data, description string) error {
	configLock.Lock()
	defer configLock.Unlock()

	var cfg map[string]string
	err := json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Printf("config %s can not marshal", key)
		return err
	}

	// Check if the config already exists
	var existingConfig CloudConfig
	cfgModel := &CloudConfig{}
	existErr := db.Where("namespace = ? and config_key = ?", namespace, key).First(&existingConfig)
	if existErr != nil {
		if errors.Is(existErr.Error, gorm.ErrRecordNotFound) {
			cfgModel.Id = existingConfig.Id
		} else {
			return existErr.Error
		}
	}

	cfgModel.ConfigKey = key
	cfgModel.Namespace = namespace
	cfgModel.ConfigValue = data
	cfgModel.ConfigName = name
	cfgModel.Description = description

	result := db.Model(&CloudConfig{}).Save(cfgModel)
	if result.Error != nil {
		return result.Error
	}

	configMap[key] = cfg
	return nil
}

func RemoveConfig(namespace, key string) {
	configLock.Lock()
	defer configLock.Unlock()

	// Soft delete config
	cfgModel := &CloudConfig{}
	cfgModel.Namespace = namespace
	cfgModel.ConfigKey = key
	cfgModel.DeletedAt = time.Now()
	result := db.Model(&CloudConfig{}).Save(&cfgModel)
	if result.Error != nil {
		log.Fatalf("Failed to delete config in the database: %v", result.Error)
	}
	delete(configMap, key)
}

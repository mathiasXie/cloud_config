package cloud_config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	db         *gorm.DB
	configMap  = make(map[string]map[string]string)
	configLock sync.RWMutex
	namespace  string
)

func Init(configDB *gorm.DB, configNamespace string) {
	db = configDB
	namespace = configNamespace
	// Check if the table exists. If it doesn't exist, create it.
	if !db.Migrator().HasTable(&CloudConfig{}) {
		if err := db.Debug().Migrator().CreateTable(&CloudConfig{}); err != nil {
			log.Fatalf("Failed to create cloud_configs table: %v", err)
		}
		log.Println("cloud_configs table created")
	}

	// Timed refresh configuration
	go refreshConfig()
}

func loadConfigFromDB() {
	var configs []CloudConfig
	result := db.Where("namespace=?", namespace).Find(&configs)
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
	existErr := db.Where("namespace = ? and config_key = ?", namespace, key).First(&existingConfig).Error
	if existErr != nil && existErr != gorm.ErrRecordNotFound {
		return existErr
	} else {
		cfgModel.Id = existingConfig.Id
	}

	cfgModel.ConfigKey = key
	cfgModel.Namespace = namespace
	cfgModel.ConfigValue = data
	cfgModel.ConfigName = name
	cfgModel.Description = description

	result := db.Save(cfgModel)
	if result.Error != nil {
		return result.Error
	}

	configMap[key] = cfg
	fmt.Printf("=======configMap: %+v\n", configMap)
	return nil
}

func RemoveConfig(key string) {
	configLock.Lock()
	defer configLock.Unlock()

	// Soft delete config
	result := db.Where("namespace = ? AND config_key = ?", namespace, key).Delete(&CloudConfig{})
	if result.Error != nil {
		log.Fatalf("Failed to delete config in the database: %v", result.Error)
	}
	delete(configMap, key)
}

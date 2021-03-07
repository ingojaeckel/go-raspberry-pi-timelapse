package conf

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const (
	settingsFile = "timelapse-settings.db"
	bucket       = "settings"
	settingsKey  = "s"
)

var missingBucketError = errors.New("missing bucket")
var settingsNotFound = errors.New("settings not found")

var initialConfiguration = Settings{
	DebugEnabled:           false,
	SecondsBetweenCaptures: 1800, // 30 min
	OffsetWithinHour:       900,  // 15 min
	// Default resolution: 3280x2464 (8 MP). 66%: 2186x1642 (3.5 MP), 50%: 1640x1232 (2 MP)
	PhotoResolutionWidth:    2186,
	PhotoResolutionHeight:   1642,
	PreviewResolutionWidth:  640,
	PreviewResolutionHeight: 480,
	RotateBy:                0,
	ResolutionSetting:       0,
	Quality:                 100,
}

type Settings struct {
	SecondsBetweenCaptures  int
	OffsetWithinHour        int
	PhotoResolutionWidth    int
	PhotoResolutionHeight   int
	PreviewResolutionWidth  int
	PreviewResolutionHeight int
	RotateBy                int
	ResolutionSetting       int
	Quality                 int
	DebugEnabled            bool
}

func LoadConfiguration() (*Settings, error) {
	if areSettingsMissing(settingsFile) {
		log.Println("Creating initial settings file..")
		return WriteConfiguration(initialConfiguration)
	}
	log.Println("Settings file exists.")

	db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	val, exists, err := get(db, settingsKey)
	if err != nil {
		log.Fatal("Error loading settings: ", err.Error())
		return nil, err
	}
	if !exists {
		return nil, settingsNotFound
	}

	var existingSettings Settings
	err = json.Unmarshal([]byte(val), &existingSettings)
	return &existingSettings, err
}

func WriteConfiguration(s Settings) (*Settings, error) {
	log.Printf("Write configuration: %v\n", s)
	db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	if err != nil {
		log.Fatalf("Failed to create settings file: %s", err.Error())
		return nil, err
	}

	marshalled, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("Failed to marshal config: %s", err.Error())
		return nil, err
	}
	if err := set(db, settingsKey, marshalled); err != nil {
		log.Fatal("Failed to write settings: ", err.Error())
		return nil, err
	}
	return &s, nil
}

func areSettingsMissing(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func set(db *bolt.DB, key string, value []byte) error {
	log.Printf("setting '%s': '%s' -> '%v'\n", bucket, key, value)
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return missingBucketError
		}
		return b.Put([]byte(key), value)
	})
}

func get(db *bolt.DB, key string) (value string, exists bool, err error) {
	log.Printf("getting '%s': '%s'\n", bucket, key)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		if b == nil {
			return missingBucketError
		}

		v := b.Get([]byte(key))
		if v == nil {
			value = ""
			exists = false
			return nil
		}
		value = string(v)
		exists = true
		log.Printf("Found value '%s': '%s' -> '%s'\n", bucket, key, value)
		return nil
	})
	return value, exists, err
}

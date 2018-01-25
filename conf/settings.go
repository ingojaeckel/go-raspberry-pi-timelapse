package conf

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"time"
)

const (
	settingsFile = "timelapse-settings.db"
	bucket       = "settings"
	settingsKey  = "s"
)

type Settings struct {
	SecondsBetweenCaptures  int
	OffsetWithinHour        int
	PhotoResolutionWidth    int
	PhotoResolutionHeight   int
	PreviewResolutionWidth  int
	PreviewResolutionHeight int
	DebugEnabled            bool
}

type setting struct {
	key   string
	value int
}

func LoadConfiguration() (*Settings, error) {
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		log.Printf("Creating initial settings file..")

		db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
		defer db.Close()

		if err != nil {
			log.Fatalf("Failed to create initial settings file: %s", err.Error())
			return nil, err
		}

		s := Settings{
			DebugEnabled:           false,
			SecondsBetweenCaptures: 1800, // 30 min
			OffsetWithinHour:       900,  // 15 min
			// Default resolution: 3280x2464 (8 MP). 66%: 2186x1642 (3.5 MP), 50%: 1640x1232 (2 MP)
			PhotoResolutionWidth:    2186,
			PhotoResolutionHeight:   1642,
			PreviewResolutionWidth:  640,
			PreviewResolutionHeight: 480,
		}

		marshalled, err := json.Marshal(s)
		if err != nil {
			log.Fatalf("Failed to marshal initial config: %s", err.Error())
			return nil, err
		}
		// TODO write as []byte instead
		val := string(marshalled)
		if err := set(db, settingsKey, val); err != nil {
			log.Fatalf("Failed to write settings: %s", err.Error())
			return nil, err
		}

		return &s, nil
	}
	log.Println("Settings file exists.")

	db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	val, exists, err := get(db, settingsKey)
	if err != nil {
		log.Fatalf("Error loading settings: %s", err.Error())
		return nil, err
	}
	if !exists {
		return nil, errors.New("Settings not found")
	}

	var existingSettings Settings
	err = json.Unmarshal([]byte(val), &existingSettings)
	return &existingSettings, err
}

func set(db *bolt.DB, key, value string) error {
	log.Printf("setting '%s': '%s' -> '%s'\n", bucket, key, value)
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}

		b := tx.Bucket([]byte(bucket))
		if b == nil {
			log.Fatalf("Bucket %s does not exist", bucket)
			return errors.New("Missing bucket")
		}
		return b.Put([]byte(key), []byte(value))
	})
}

func get(db *bolt.DB, key string) (value string, exists bool, err error) {
	log.Printf("getting '%s': '%s'\n", bucket, key)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		if b == nil {
			log.Fatalf("Bucket %s does not exist", bucket)
			return errors.New("Missing bucket")
		}

		v := b.Get([]byte(key))
		if v == nil {
			value = ""
			exists = false
			return nil
		}
		value = string(v)
		exists = true
		return nil
	})
	return value, exists, err
}

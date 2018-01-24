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
)

type Settings struct {
	SecondsBetweenCaptures int
	OffsetWithinHour       int
	PhotoResolutionWidth   int
	PhotoResolutionHeight  int
}

type setting struct {
	key   string
	value int
}

func LoadConfiguration() (*Settings, error) {

	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		// create initial settings file

		db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
		defer db.Close()

		if err != nil {
			log.Fatalf("Failed to create initial settings file: %s", err.Error())
			return nil, err
		}

		s := Settings{
			SecondsBetweenCaptures: 1800,
			OffsetWithinHour:       1800,
			PhotoResolutionWidth:   1800,
			PhotoResolutionHeight:  1800,
		}

		marshalled, err := json.Marshal(s)
		if err != nil {
			log.Fatalf("Failed to marshal initial config: %s", err.Error())
			return nil, err
		}
		// TODO write as []byte instead
		if err := set(db, "settings", string(marshalled)); err != nil {
			log.Fatalf("Failed to write settings: %s", err.Error())
			return nil, err
		}

		return &s, nil
	}

	db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()

	val, exists, err := get(db, "settings")
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
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func get(db *bolt.DB, key string) (value string, exists bool, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
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

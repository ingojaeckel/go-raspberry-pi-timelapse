package conf

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	settingsFile                  = "timelapse-settings.db"
	bucket                        = "settings"
	settingSecondsBetweenCaptures = "secondsBetweenCaptures"
	settingOffsetWithinHour       = "offsetWithinHour"
	settingPhotoResolutionWidth   = "photoResolutionWidth"
	settingPhotoResolutionHeight  = "photoResolutionHeight"
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

func LoadConfiguration() error {

	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		// create initial settings file

		db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			log.Fatalf("Failed to create initial settings file: %s", err.Error())
			return err
		}

		settings := []setting{
			{settingSecondsBetweenCaptures, 1800},
			{settingOffsetWithinHour, 1800},
			{settingPhotoResolutionWidth, 1800},
			{settingPhotoResolutionHeight, 1800},
		}
		if err := setInts(db, settings); err != nil {
			log.Fatalf("Failed to initialize settings file with initial settings: %s", err.Error())
			return err
		}
	}

	db, err := bolt.Open(settingsFile, 0600, &bolt.Options{Timeout: 1 * time.Second})

	getInt(db, settingSecondsBetweenCaptures)

	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	return nil
}

func setInts(db *bolt.DB, settings []setting) error {
	for _, s := range settings {
		if err := setInt(db, s.key, s.value); err != nil {
			return err
		}
	}
	return nil
}

func setInt(db *bolt.DB, key string, value int) error {
	strVal := strconv.Itoa(value)
	return set(db, key, strVal)
}

func set(db *bolt.DB, key, value string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(value))
		return err
	})
}

func getInt(db *bolt.DB, key string) (int, bool, error) {
	v, exists, err := get(db, key)

	if exists && err == nil {
		intVal, err := strconv.Atoi(v)
		return intVal, exists, err
	}
	// Either the key does not exist, or there was an error. In either case there is nothing to convert to int.
	// Return -1 to indicate that this is the case..
	return -1, exists, err
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
